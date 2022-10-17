package main

import (
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
const defaultClientAddr = "https://127.0.0.1:8080"
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
				Name:  "client",
				Usage: "run in client mode",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "url",
						Usage: "file to download",
						Value: defaultClientAddr,
					},
					&cli.StringFlag{
						Name:  "tls-cert",
						Usage: "TLS certificate file to use",
						Value: defaultTLSCertificateFile,
					},
				},
				Action: func(c *cli.Context) error {
					return client.Run(client.ClientConfig{
						Url:         c.String("url"),
						TLSCertFile: c.String("tls-cert"),
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
				},
				Action: func(c *cli.Context) error {
					return server.Run(server.ServerConfig{
						Addr:        c.String("addr"),
						ServeDir:    c.String("dir"),
						TlsCertFile: c.String("tls-cert"),
						TlsKeyFile:  c.String("tls-key"),
					})
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		panic(err)
	}
}
