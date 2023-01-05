package client

import (
	"bytes"
	"crypto/tls"
	r "github.com/birneee/webpage-requisites-go"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/logging"
	log "github.com/sirupsen/logrus"
	"http-perf-go/internal"
	"io"
	"net"
	"net/http"
	u "net/url"
	"regexp"
	"strings"
	"sync/atomic"
	"time"
)

// TODO 0rtt
type Config struct {
	Urls                  []*u.URL
	TLSCertFile           string
	Qlog                  bool
	QlogPrefix            string
	PageRequisites        bool
	ParallelRequests      int
	ProxyConfig           *quic.ProxyConfig
	AllowEarlyHandover    bool
	ExtraStreamEncryption bool
	UserAgent             string
	UrlBlacklist          []*regexp.Regexp
}

type client struct {
	config               *Config
	httpClient           *http.Client
	totalReceivedBytes   atomic.Int64
	totalQuicConnections atomic.Uint32
	totalGetRequests     atomic.Int64
	totalHttpErrors      atomic.Int64
}

// Run blocks until everything is downloaded
func Run(config Config) error {
	certPool, err := internal.SystemCertPoolWithAdditionalCert(config.TLSCertFile)
	if err != nil {
		return err
	}

	client := &client{
		config: &config,
	}

	tlsConf := &tls.Config{
		RootCAs: certPool,
	}

	tracers := make([]logging.Tracer, 0)

	tracers = append(tracers, internal.NewEventTracer(internal.Handlers{
		UpdatePath: func(odcid logging.ConnectionID, newRemote net.Addr) {
			log.Infof("migrated QUIC connection %s to %s", odcid.String(), newRemote)
		},
		StartedConnection: func(odcid logging.ConnectionID, local, remote net.Addr, srcConnID, destConnID logging.ConnectionID) {
			log.Infof("started QUIC connection %s", odcid.String())
			client.totalQuicConnections.Add(1)
		},
		ClosedConnection: func(odcid logging.ConnectionID, err error) {
			log.Infof("closed QUIC connection %s", odcid.String())
		},
	}))

	if config.Qlog {
		tracers = append(tracers, internal.NewQlogTracer(config.QlogPrefix, func(filename string) {
			log.Infof("created qlog file: %s", filename)
		}))
	}

	quicConf := &quic.Config{
		Tracer:                logging.NewMultiplexedTracer(tracers...),
		ProxyConf:             config.ProxyConfig,
		EnableActiveMigration: true,
		AllowEarlyHandover:    config.AllowEarlyHandover,
	}

	if config.ExtraStreamEncryption {
		quicConf.ExtraStreamEncryption = quic.EnforceExtraStreamEncryption
	} else {
		quicConf.ExtraStreamEncryption = quic.DisableExtraStreamEncryption
	}

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConf,
		QuicConfig:      quicConf,
	}
	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}

	client.httpClient = hclient

	urlQueue := internal.NewDistinctChannel[u.URL](1024)

	pendingRequests := internal.NewCondHelper(0)

	for _, url := range config.Urls {
		distinct := urlQueue.Add(*url)
		if distinct {
			pendingRequests.UpdateState(func(s int) int { return s + 1 })
		}
	}

	firstRequestTime := time.Now()
	for i := 0; i < config.ParallelRequests; i++ {
		go func() {
			for {
				url := urlQueue.Next()
				if client.isUrlIgnored(url) {
					log.Infof("skip blacklisted url: %s", url.String())
				} else {
					receivedBytes, err := client.download(&url, func(url *u.URL) {
						distinct := urlQueue.Add(*url)
						if distinct {
							pendingRequests.UpdateState(func(s int) int { return s + 1 })
						}
					})
					client.totalReceivedBytes.Add(receivedBytes)
					if err != nil {
						log.Errorf("failed to download %s: %v", url.String(), err)
					}
				}
				pendingRequests.UpdateState(func(s int) int { return s - 1 })
			}
		}()
	}

	pendingRequests.Wait(func(s int) bool { return s == 0 })

	log.Infof("total bytes received: %d B, time: %.3f s, get requests: %d, http errors: %d, quic connections: %d", client.totalReceivedBytes.Load(), time.Now().Sub(firstRequestTime).Seconds(), client.totalGetRequests.Load(), client.totalHttpErrors.Load(), client.totalQuicConnections.Load())

	return nil
}

// return received bytes
func (c *client) download(url *u.URL, onFindRequisite func(*u.URL)) (int64, error) {
	c.totalGetRequests.Add(1)
	log.Infof("GET %s", url)
	start := time.Now()
	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("user-agent", c.config.UserAgent)
	rsp, err := c.httpClient.Do(req)
	if err != nil {
		return 0, err
	}

	//TODO convert HTML and CSS with other encodings to UTF-8
	contentType := strings.ToLower(strings.Split(rsp.Header.Get("Content-Type"), ";")[0])

	var received int64
	var requisites []*u.URL
	var stop time.Time

	if c.config.PageRequisites && contentType == internal.MIME_TYPE_TEXT_HTML {
		html, err := io.ReadAll(rsp.Body)
		if err != nil {
			return 0, err
		}
		stop = time.Now()
		received = int64(len(html))
		requisites, err = r.GetHtmlRequisites(bytes.NewReader(html))
		if err != nil {
			return 0, err
		}
	} else if c.config.PageRequisites && contentType == internal.MIME_TYPE_TEXT_CSS {
		css, err := io.ReadAll(rsp.Body)
		if err != nil {
			return 0, err
		}
		stop = time.Now()
		received = int64(len(css))
		requisites, err = r.GetCssRequisites(string(css))
		if err != nil {
			return 0, err
		}
	} else {
		received, err = io.Copy(internal.DiscardWriter{}, rsp.Body)
		if err != nil {
			return 0, err
		}
		stop = time.Now()
	}

	if isHttpStatusError(rsp.StatusCode) {
		c.totalHttpErrors.Add(1)
	}
	log.Infof("got %s %s %d, %d byte, %f s", url, rsp.Proto, rsp.StatusCode, received, stop.Sub(start).Seconds())

	for _, requisite := range requisites {
		absolute := url.ResolveReference(requisite)
		onFindRequisite(absolute)
	}

	return received, nil
}

func isHttpStatusError(statusCode int) bool {
	return statusCode < 200 || statusCode >= 300
}

func (c *client) isUrlIgnored(url u.URL) bool {
	for _, regex := range c.config.UrlBlacklist {
		if regex.MatchString(url.String()) {
			return true
		}
	}
	return false
}
