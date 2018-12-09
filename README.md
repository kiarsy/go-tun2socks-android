# go-tun2socks-android

Demo for building and using `go-tun2socks` on Android.

# Including V2Ray features

Refer [here](https://github.com/eycorsican/go-tun2socks/tree/master/proxy/v2ray) for details, specifically, features to include are defining in these [two](https://github.com/eycorsican/go-tun2socks/blob/master/proxy/v2ray/features.go) [files](https://github.com/eycorsican/go-tun2socks/blob/master/proxy/v2ray/features_other.go).

## Prerequisites

- make
- Go >= 1.11
- A C compiler (e.g.: clang, gcc)
- gomobile (https://github.com/golang/go/wiki/Mobile)
- Other common utilities (e.g.: git)

## Build
```bash
go get -d ./...
make
```
