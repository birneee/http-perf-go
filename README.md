# http-perf-go

A performance measurement tool for HTTP/3

## Build

```bash
go build
```

## Run Server

```bash
http-perf-go server
```

## Run Client

```bash
http-perf-go client
```

It is recommended to increase the maximum buffer size by running (See https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size for details):

```bash
sysctl -w net.core.rmem_max=2500000
```


## Generate Self-signed certificate

```bash
openssl req -x509 -nodes -days 358000 -out server.crt -keyout server.key -config server.req
```