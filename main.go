package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"http-perf-go/client"
	"http-perf-go/internal"
	"http-perf-go/server"
	u "net/url"
	"os"
)

const defaultTLSCertificateFile = "./server.crt"
const defaultTLSKeyFile = "./server.key"
const defaultServeDir = "./www"
const defaultServerAddr = "0.0.0.0:8080"

func main() {
	log.SetFormatter(internal.NewFormatter())
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
					&cli.BoolFlag{
						Name:    "page-requisites",
						Aliases: []string{"p"},
						Usage:   "This option causes Wget to download all the files that are necessary to properly display a given HTML page.  This includes such things as inlined images, sounds, and referenced stylesheets.",
						Value:   false,
					},
					&cli.UintFlag{
						Name:  "parallel",
						Usage: "Number of parallel requests to send",
						Value: 10,
					},
				},
				Action: func(c *cli.Context) error {
					if c.Args().Len() == 0 {
						return fmt.Errorf("missing URL")
					}
					var urls []*u.URL
					for _, urlStr := range c.Args().Slice() {
						url, err := u.ParseRequestURI(urlStr)
						if err != nil {
							return fmt.Errorf("invalid url %s: %v", urlStr, err)
						}
						urls = append(urls, url)
					}
					return client.Run(client.Config{
						Urls:             urls,
						TLSCertFile:      c.String("tls-cert"),
						Qlog:             c.Bool("qlog"),
						PageRequisites:   c.Bool("page-requisites"),
						ParallelRequests: c.Int("parallel"),
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
