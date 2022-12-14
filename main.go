package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"github.com/lucas-clemente/quic-go"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"http-perf-go/client"
	"http-perf-go/internal"
	"http-perf-go/server"
	u "net/url"
	"os"
	"regexp"
)

const (
	defaultTLSCertificateFile = "./server.crt"
	defaultTLSKeyFile         = "./server.key"
	defaultServeDir           = "./www"
	defaultServerAddr         = "0.0.0.0:8080"
	defaultUserAgent          = "http-perf-go"
)

// TODO add xse option
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
					&cli.StringFlag{
						Name:  "proxy",
						Usage: "the proxy to use, in the form \"host:port\", default port 18081 if not specified",
					},
					&cli.BoolFlag{
						Name:  "proxy-0rtt",
						Usage: "gather 0-RTT information to the proxy beforehand",
						Value: false,
					},
					&cli.BoolFlag{
						Name:  "early-handover",
						Usage: "allow creating H-QUIC state earlier, when handshake is completed but not yet confirmed. Optimistic approach! Success is not guaranteed due to race conditions.",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "tls-proxy-cert",
						Usage: "certificate file to trust the proxy",
					},
					&cli.StringFlag{
						Name:    "user-agent",
						Aliases: []string{"U"},
						Usage:   "Identification of client to the HTTP server",
						Value:   defaultUserAgent,
					},
					&cli.StringFlag{
						Name:  "url-blacklist",
						Usage: "file containing regular expressions for urls that will not be requested",
					},
					&cli.BoolFlag{
						Name:  "xse",
						Usage: "use XSE-QUIC extension; handshake will fail if not supported by server",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "qlog-prefix",
						Usage: "the prefix of the qlog file name",
						Value: "client",
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

					var proxyConf *quic.ProxyConfig
					if c.IsSet("proxy") {
						proxyConf = &quic.ProxyConfig{}
						proxyAddr, err := internal.ParseResolveHost(c.String("proxy"), quic.DefaultHQUICProxyControlPort)
						if err != nil {
							return fmt.Errorf("failed to resolve proxy address: %w", err)
						}
						proxyConf.Addr = proxyAddr.String()
						proxyConf.Config = &quic.Config{
							TokenStore: quic.NewLRUTokenStore(1, 1),
						}
						if c.String("tls-proxy-cert") != "" {
							proxyConf.TlsConf = &tls.Config{
								NextProtos:         []string{quic.HQUICProxyALPN},
								ClientSessionCache: tls.NewLRUClientSessionCache(1),
							}
							certPool, err := internal.SystemCertPoolWithAdditionalCert(c.String("tls-proxy-cert"))
							if err != nil {
								return fmt.Errorf("failed to load proxy certificate: %w", err)
							}
							proxyConf.TlsConf.RootCAs = certPool
						}

						if c.IsSet("proxy-0rtt") {
							err := internal.PingToGatherSessionTicketAndToken(proxyConf.Addr, proxyConf.TlsConf, proxyConf.Config)
							if err != nil {
								panic(fmt.Errorf("failed to prepare 0-RTT to proxy: %w", err))
							}
							log.Infof("stored session ticket and address token of proxy for 0-RTT")
						}
					}

					urlBlacklist := make([]*regexp.Regexp, 0)
					if c.IsSet("url-blacklist") {
						file, err := os.Open(c.String("url-blacklist"))
						if err != nil {
							return fmt.Errorf("failed to open url blacklist: %v", err)
						}
						defer file.Close()
						scanner := bufio.NewScanner(file)
						for scanner.Scan() {
							expr, err := regexp.Compile(scanner.Text())
							if err != nil {
								return fmt.Errorf("failed to compile url blacklist regexp: %v", err)
							}
							urlBlacklist = append(urlBlacklist, expr)
						}
					}

					return client.Run(client.Config{
						Urls:                  urls,
						TLSCertFile:           c.String("tls-cert"),
						Qlog:                  c.Bool("qlog"),
						QlogPrefix:            c.String("qlog-prefix"),
						PageRequisites:        c.Bool("page-requisites"),
						ParallelRequests:      c.Int("parallel"),
						ProxyConfig:           proxyConf,
						AllowEarlyHandover:    c.Bool("early-handover"),
						ExtraStreamEncryption: c.Bool("xse"),
						UserAgent:             c.String("user-agent"),
						UrlBlacklist:          urlBlacklist,
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
					&cli.BoolFlag{
						Name:  "multi-domain",
						Usage: "interpret first level in directory as the domain name",
						Value: false,
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
					&cli.BoolFlag{
						Name:  "query-in-filename",
						Usage: "serve files with query string in filename",
						Value: false,
					},
					&cli.StringFlag{
						Name:  "qlog-prefix",
						Usage: "the prefix of the qlog file name",
						Value: "server",
					},
				},
				Action: func(c *cli.Context) error {
					return server.Run(server.Config{
						Addr:                  c.String("addr"),
						ServeDir:              c.String("dir"),
						TlsCertFile:           c.String("tls-cert"),
						TlsKeyFile:            c.String("tls-key"),
						Qlog:                  c.Bool("qlog"),
						QlogPrefix:            c.String("qlog-prefix"),
						MultiDomain:           c.Bool("multi-domain"),
						QueryStringInFilename: c.Bool("query-in-filename"),
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
