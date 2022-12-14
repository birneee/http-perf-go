package server

import (
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/logging"
	log "github.com/sirupsen/logrus"
	"http-perf-go/internal"
	"net"
	"net/http"
	"strings"
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
	QlogPrefix  string
	MultiDomain bool
	// serve files with query strings in its filenames.
	// e.g. wget does put them in the filename
	QueryStringInFilename bool
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

	tcpAddr, err := net.ResolveTCPAddr("tcp", config.Addr)
	if err != nil {
		return err
	}
	tcpConn, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return err
	}
	defer tcpConn.Close()

	log.Infof("listening on %s, serving %s", udpAddr, config.ServeDir)

	tlsCert, err := tls.LoadX509KeyPair(config.TlsCertFile, config.TlsKeyFile)
	if err != nil {
		return err
	}

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
	}

	tlsConn := tls.NewListener(tcpConn, tlsConf)
	defer tlsConn.Close()

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
		tracers = append(tracers, internal.NewQlogTracer(config.QlogPrefix, func(filename string) {
			log.Infof("created qlog file: %s", filename)
		}))
	}

	quicConf := &quic.Config{
		Tracer:                logging.NewMultiplexedTracer(tracers...),
		EnableActiveMigration: true,
		ExtraStreamEncryption: quic.PreferExtraStreamEncryption,
	}

	fileServerConfig := internal.FileServerConfig{
		QueryStringAsPartOfFile: config.QueryStringInFilename,
	}

	var handler http.Handler
	if config.MultiDomain {
		handler, err = internal.NewHostnameDirectoryMultiplexHandler(config.ServeDir, fileServerConfig)
		if err != nil {
			return fmt.Errorf("failed to create handler: %v", err)
		}
		log.Infof("hostnames: %s", strings.Join(handler.(internal.HostnameDirectoryMultiplexHandler).Hostnames(), ", "))
	} else {
		handler = internal.NewFileServer(http.Dir(config.ServeDir), fileServerConfig)
	}

	// HTTP/3 server
	quicServer := http3.Server{
		Handler:    handler,
		Addr:       config.Addr,
		QuicConfig: quicConf,
		TLSConfig:  tlsConf,
	}

	// HTTP/1.1 server
	tcpServer := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			quicServer.SetQuicHeaders(w.Header())
			quicServer.Handler.ServeHTTP(w, r)
		}),
	}

	tErr := make(chan error)
	qErr := make(chan error)
	go func() {
		tErr <- tcpServer.Serve(tlsConn)
	}()
	go func() {
		qErr <- quicServer.Serve(udpConn)
	}()

	select {
	case err := <-tErr:
		quicServer.Close()
		return err
	case err := <-qErr:
		tcpServer.Close()
		return err
	}
}
