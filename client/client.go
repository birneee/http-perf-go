package client

import "C"
import (
	"crypto/tls"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/logging"
	log "github.com/sirupsen/logrus"
	"http-perf-go/internal"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// TODO 0rtt
// TODO parallel downloads
// TODO page requisites
type Config struct {
	Urls           []string
	TLSCertFile    string
	Qlog           bool
	ProxyConfig    *quic.ProxyConfig
	PageRequisites bool
}

func Run(config Config) error {
	certPool, err := internal.NewCertPoolWithCert(config.TLSCertFile)
	if err != nil {
		return err
	}

	tlsConf := &tls.Config{
		RootCAs: certPool,
	}

	tracers := make([]logging.Tracer, 0)

	tracers = append(tracers, internal.NewMigrationTracer(func(addr net.Addr) {
		log.Infof("migrated to %s", addr)
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

	//TODO support parallel downloads
	for _, url := range config.Urls {
		log.Infof("GET %s", url)
		start := time.Now()
		rsp, err := hclient.Get(url)
		if err != nil {
			return err
		}

		contentType := strings.ToLower(strings.Split(rsp.Header.Get("Content-Type"), ";")[0])

		var received int64
		switch contentType {
		case "text/html":
			html, err := io.ReadAll(rsp.Body)
			if err != nil {
				return err
			}
			received = int64(len(html))
		case "text/css":
			css, err := io.ReadAll(rsp.Body)
			if err != nil {
				return err
			}
			received = int64(len(css))
		default:
			received, err = io.Copy(internal.DiscardWriter{}, rsp.Body)
			if err != nil {
				return err
			}
		}

		stop := time.Now()
		log.Infof("got %s %s %d, %d byte, %f s", url, rsp.Proto, rsp.StatusCode, received, stop.Sub(start).Seconds())
	}

	return nil
}
