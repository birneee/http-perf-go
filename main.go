package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"http-perf-go/client"
	"http-perf-go/server"
	"os"
)

const defaultTLSCertificateFile = "./server.crt"
const defaultTLSKeyFile = "./server.key"
const defaultServeDir = "./www"
const defaultServerAddr = "0.0.0.0:8080"

const timeFormatRFC3339Micro = "2006-01-02T15:04:05.999Z07:00"

func main() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: timeFormatRFC3339Micro,
	})
	app := &cli.App{
		Name:  "http-perf-go",
		Usage: "A performance measurement tool for HTTP/3",
		Commands: []*cli.Command{
			{
				Name:      "client",
				Usage:     "run in client mode",
				ArgsUsage: "[URL...]",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "tls-cert",
						Usage: "TLS certificate file to use",
						Value: defaultTLSCertificateFile,
					},
					&cli.BoolFlag{
						Name:  "qlog",
						Usage: "create qlog file",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					if c.Args().Len() == 0 {
						return fmt.Errorf("missing URL")
					}
					return client.Run(client.Config{
						Urls:        c.Args().Slice(),
						TLSCertFile: c.String("tls-cert"),
						Qlog:        c.Bool("qlog"),
					})
				},
			},
			{
				Name:  "server",
				Usage: "run in server mode",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "addr",
						Usage: "IP address and UDP port to listen on",
						Value: defaultServerAddr,
					},
					&cli.StringFlag{
						Name:  "dir",
						Usage: "directory to serve",
						Value: defaultServeDir,
					},
					&cli.StringFlag{
						Name:  "tls-cert",
						Usage: "TLS certificate file to use",
						Value: defaultTLSCertificateFile,
					},
					&cli.StringFlag{
						Name:  "tls-key",
						Usage: "TLS key file to use",
						Value: defaultTLSKeyFile,
					},
					&cli.BoolFlag{
						Name:  "qlog",
						Usage: "create qlog file",
						Value: false,
					},
				},
				Action: func(c *cli.Context) error {
					return server.Run(server.Config{
						Addr:        c.String("addr"),
						ServeDir:    c.String("dir"),
						TlsCertFile: c.String("tls-cert"),
						TlsKeyFile:  c.String("tls-key"),
						Qlog:        c.Bool("qlog"),
					})
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Errorf("%v", err)
		os.Exit(1)
	}
}
