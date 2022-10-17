package server

import (
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	log "github.com/sirupsen/logrus"
	"net"
	"net/http"
)
import "github.com/lucas-clemente/quic-go/http3"

type ServerConfig struct {
	TlsCertFile string
	TlsKeyFile  string
	ServeDir    string
	Addr        string
}

func Run(config ServerConfig) error {
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

	log.Infof("server listening on %s, serving %s", udpAddr, config.ServeDir)

	tlsCert, err := tls.LoadX509KeyPair(config.TlsCertFile, config.TlsKeyFile)
	if err != nil {
		return err
	}
	tlsConf := http3.ConfigureTLSConfig(&tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	})
	quicConf := &quic.Config{}
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
