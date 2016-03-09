# [gentleman](https://github.com/h2non/gentleman)-consul [![Build Status](https://travis-ci.org/h2non/gentleman.png)](https://travis-ci.org/h2non/gentleman-consul) [![GitHub release](https://img.shields.io/badge/version-0.1.0-orange.svg?style=flat)](https://github.com/h2non/gentleman-consul/releases) [![GoDoc](https://godoc.org/github.com/h2non/gentleman-consul?status.svg)](https://godoc.org/github.com/h2non/gentleman-consul) [![Go Report Card](https://goreportcard.com/badge/github.com/h2non/gentleman-consul)](https://goreportcard.com/report/github.com/h2non/gentleman-consul)

<!--
[![Coverage Status](https://coveralls.io/repos/github/h2non/gentleman-consul/badge.svg?branch=master)](https://coveralls.io/github/h2non/gentleman-consul?branch=master)
-->

[gentleman](https://github.com/h2non/gentleman)'s plugin for easy service discovery and dynamic balancing using [Consul](https://www.consul.io).

Provides transparent retry/backoff support for resilient and [reactive](http://www.reactivemanifesto.org) HTTP client capabilities. Also it allows you to use a custom [retry strategy](https://github.com/h2non/gentleman-retry/blob/ce34094db8b9811b45e0395b64f9b1188cabb3ca/retry.go#L35-L38), such an [constant](https://godoc.org/github.com/eapache/go-resiliency/retrier#ConstantBackoff) or [exponential](https://godoc.org/github.com/eapache/go-resiliency/retrier#ExponentialBackoff) back off.

## Installation

```bash
go get -u gopkg.in/h2non/gentleman-consul.v0
```

## API

See [godoc reference](https://godoc.org/github.com/h2non/gentleman-consul) for detailed API documentation.

## Examples

See [examples](https://github.com/h2non/gentleman-consul/blob/master/_examples) directory for featured examples.

#### Simple request

```go
package main

import (
  "fmt"
  "gopkg.in/h2non/gentleman-consul.v0"
  "gopkg.in/h2non/gentleman.v0"
)

func main() {
  // Create a new client
  cli := gentleman.New()

  // Register Consul's plugin at client level
  cli.Use(consul.New(consul.NewConfig("demo.consul.io", "web")))

  // Create a new request based on the current client
  req := cli.Request()

  // Set a new header field
  req.SetHeader("Client", "gentleman")

  // Perform the request
  res, err := req.Send()
  if err != nil {
    fmt.Printf("Request error: %s\n", err)
    return
  }
  if !res.Ok {
    fmt.Printf("Invalid server response: %d\n", res.StatusCode)
    return
  }

  // Reads the whole body and returns it as string
  fmt.Printf("Server URL: %s\n", res.RawRequest.URL.String())
  fmt.Printf("Response status: %d\n", res.StatusCode)
  fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
```

## License 

MIT - Tomas Aparicio
