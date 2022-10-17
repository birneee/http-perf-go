package client

import (
	"bytes"
	"crypto/tls"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	log "github.com/sirupsen/logrus"
	"http-perf-go/common"
	"io"
	"net/http"
)

type ClientConfig struct {
	Url         string
	TLSCertFile string
}

func Run(config ClientConfig) error {
	certPool, err := common.NewCertPoolWithCert(config.TLSCertFile)
	if err != nil {
		return err
	}

	tlsConf := &tls.Config{
		RootCAs: certPool,
	}

	quicConf := &quic.Config{}

	roundTripper := &http3.RoundTripper{
		TLSClientConfig: tlsConf,
		QuicConfig:      quicConf,
	}
	defer roundTripper.Close()

	hclient := &http.Client{
		Transport: roundTripper,
	}

	rsp, err := hclient.Get(config.Url)
	if err != nil {
		return err
	}

	body := &bytes.Buffer{}

	_, err = io.Copy(body, rsp.Body)
	if err != nil {
		return err
	}

	log.Infof("Response Body: %s", body.Bytes())

	return nil
}
