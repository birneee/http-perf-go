# http-perf-go

A performance measurement tool for HTTP/3.
Uses https://github.com/lucas-clemente/quic-go

## Example
```bash
$ http-perf-go server
INFO[2022-10-18T22:57:04.587+02:00] listening on 0.0.0.0:8080, serving ./www
```

```bash
$ http-perf-go client https://localhost:8080/
INFO[2022-10-18T22:57:07.564+02:00] GET https://localhost:8080/
INFO[2022-10-18T22:57:07.571+02:00] got https://localhost:8080/ HTTP/3.0 200, 5 byte, 0.003580 s
```

## Build

```bash
go build
```

## Setup

It is recommended to increase the maximum buffer size by running (See https://github.com/lucas-clemente/quic-go/wiki/UDP-Receive-Buffer-Size for details):

```bash
sysctl -w net.core.rmem_max=2500000
```


## Generate Self-signed certificate

```bash
openssl req -x509 -nodes -days 358000 -out server.crt -keyout server.key -config server.req
```