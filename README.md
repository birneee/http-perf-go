# http-perf-go

A performance measurement tool for HTTP/3.
Uses https://github.com/lucas-clemente/quic-go

## Example
```bash
$ http-perf-go server
INFO T0.000159 listening on 0.0.0.0:8080, serving ./www
```

```bash
$ http-perf-go client https://localhost:8080/
INFO T0.013947 GET https://localhost:8080/
INFO T0.019012 got https://localhost:8080/ HTTP/3.0 200, 5 byte, 0.003580 s
INFO T0.019016 total bytes received: 5 B
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

## Comparison with wget

```bash
$ wget -pH -e robots=off -U "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36" --delete-after https://www.google.com 2>&1 | grep -oP 'https?://[^\s]+' | sort

https://fonts.gstatic.com/s/i/productlogos/googleg/v6/24px.svg
https://ssl.gstatic.com/ui/v1/menu/checkmark2.png
https://www.google.com/
https://www.google.com/images/branding/googlelogo/1x/googlelogo_color_272x92dp.png
https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png
https://www.google.com/images/experiments/wavy-underline.png
https://www.google.com/images/searchbox/desktop_searchbox_sprites318_hr.webp
https://www.google.com/manifest?pwa=webhp
```

```bash
$ http-perf-go client -p -U "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/107.0.0.0 Safari/537.36" https://www.google.com 2>&1 | grep -oP '(?<=GET )(https?://[^\s]+)' | sort

https://fonts.gstatic.com/s/i/productlogos/googleg/v6/24px.svg
https://ssl.gstatic.com/ui/v1/menu/checkmark2.png
https://www.google.com
https://www.google.com/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png
https://www.google.com/images/experiments/wavy-underline.png
https://www.google.com/images/searchbox/desktop_searchbox_sprites318_hr.webp
https://www.google.com/manifest?pwa=webhp
```