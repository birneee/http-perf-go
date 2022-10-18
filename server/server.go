package server

import (
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/logging"
	log "github.com/sirupsen/logrus"
	"http-perf-go/common"
	"net"
	"net/http"
)
import "github.com/lucas-clemente/quic-go/http3"

// TODO chromium based browsers
// TODO HTTP/1
type Config struct {
	TlsCertFile string
	TlsKeyFile  string
	ServeDir    string
	Addr        string
	Qlog        bool
}

func Run(config Config) error {
	// Open the listeners
	udpAddr, err := net.ResolveUDPAddr("udp", config.Addr)
	if err != nil {
		return err
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return err
	}
	defer udpConn.Close()

	log.Infof("listening on %s, serving %s", udpAddr, config.ServeDir)

	tlsCert, err := tls.LoadX509KeyPair(config.TlsCertFile, config.TlsKeyFile)
	if err != nil {
		return err
	}
	tlsConf := http3.ConfigureTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})

	tracers := make([]logging.Tracer, 0)

	if config.Qlog {
		tracers = append(tracers, common.NewQlogTracer("server", func(filename string) {
			log.Infof("created qlog file: %s", filename)
		}))
	}

	quicConf := &quic.Config{
		Tracer: logging.NewMultiplexedTracer(tracers...),
	}

	server := http3.Server{
		Handler:    http.FileServer(http.Dir(config.ServeDir)),
		Addr:       config.Addr,
		QuicConfig: quicConf,
		TLSConfig:  tlsConf,
	}
	err = server.Serve(udpConn)
	if err != nil {
		fmt.Println(err)
	}
	return nil
}
