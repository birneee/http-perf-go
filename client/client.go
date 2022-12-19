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
	"strings"
	"sync/atomic"
	"time"
)

// TODO 0rtt
type Config struct {
	Urls             []*u.URL
	TLSCertFile      string
	Qlog             bool
	PageRequisites   bool
	ParallelRequests int
	ProxyConfig      *quic.ProxyConfig
	UserAgent        string
}

type client struct {
	config     *Config
	httpClient *http.Client
}

// Run blocks until everything is downloaded
func Run(config Config) error {
	certPool, err := internal.SystemCertPoolWithAdditionalCert(config.TLSCertFile)
	if err != nil {
		return err
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
		},
		ClosedConnection: func(odcid logging.ConnectionID, err error) {
			log.Infof("closed QUIC connection %s", odcid.String())
		},
	}))

	if config.Qlog {
		tracers = append(tracers, internal.NewQlogTracer("client", func(filename string) {
			log.Infof("created qlog file: %s", filename)
		}))
	}

	quicConf := &quic.Config{
		Tracer:                logging.NewMultiplexedTracer(tracers...),
		ProxyConf:             config.ProxyConfig,
		EnableActiveMigration: true,
	}

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConf,
		QuicConfig:      quicConf,
	}
	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}

	client := &client{
		config:     &config,
		httpClient: hclient,
	}

	urlQueue := internal.NewDistinctChannel[u.URL](1024)

	pendingRequests := internal.NewCondHelper(0)
	var totalReceivedBytes int64 = 0 /* atomic */

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
				receivedBytes, err := client.download(&url, func(url *u.URL) {
					distinct := urlQueue.Add(*url)
					if distinct {
						pendingRequests.UpdateState(func(s int) int { return s + 1 })
					}
				})
				atomic.AddInt64(&totalReceivedBytes, receivedBytes)
				pendingRequests.UpdateState(func(s int) int { return s - 1 })
				if err != nil {
					log.Errorf("failed to download %s: %v", url.String(), err)
				}
			}
		}()
	}

	pendingRequests.Wait(func(s int) bool { return s == 0 })

	log.Infof("total bytes received: %d B, time: %.3f s", atomic.LoadInt64(&totalReceivedBytes), time.Now().Sub(firstRequestTime).Seconds())

	return nil
}

// return received bytes
func (c *client) download(url *u.URL, onFindRequisite func(*u.URL)) (int64, error) {
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

	log.Infof("got %s %s %d, %d byte, %f s", url, rsp.Proto, rsp.StatusCode, received, stop.Sub(start).Seconds())

	for _, requisite := range requisites {
		absolute := url.ResolveReference(requisite)
		onFindRequisite(absolute)
	}

	return received, nil
}
