package client

import (
	"crypto/tls"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/logging"
	log "github.com/sirupsen/logrus"
	"http-perf-go/common"
	"io"
	"net/http"
	"time"
)

// TODO 0rtt
// TODO parallel downloads
// TODO page requisites
type Config struct {
	Urls        []string
	TLSCertFile string
	Qlog        bool
}

func Run(config Config) error {
	certPool, err := common.NewCertPoolWithCert(config.TLSCertFile)
	if err != nil {
		return err
	}

	tlsConf := &tls.Config{
		RootCAs: certPool,
	}

	tracers := make([]logging.Tracer, 0)

	if config.Qlog {
		tracers = append(tracers, common.NewQlogTracer("client", func(filename string) {
			log.Infof("created qlog file: %s", filename)
		}))
	}

	quicConf := &quic.Config{
		Tracer: logging.NewMultiplexedTracer(tracers...),
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

		received, err := io.Copy(common.DiscardWriter{}, rsp.Body)
		if err != nil {
			return err
		}
		stop := time.Now()
		log.Infof("got %s %s %d, %d byte, %f s", url, rsp.Proto, rsp.StatusCode, received, stop.Sub(start).Seconds())
	}

	return nil
}
