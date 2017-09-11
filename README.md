# [gentleman](https://github.com/h2non/gentleman)-consul [![Build Status](https://travis-ci.org/h2non/gentleman.png)](https://travis-ci.org/h2non/gentleman-consul) [![GoDoc](https://godoc.org/github.com/h2non/gentleman-consul?status.svg)](https://godoc.org/github.com/h2non/gentleman-consul) [![Coverage Status](https://coveralls.io/repos/github/h2non/gentleman-consul/badge.svg?branch=master)](https://coveralls.io/github/h2non/gentleman-consul?branch=master) [![Go Report Card](https://goreportcard.com/badge/github.com/h2non/gentleman-consul)](https://goreportcard.com/report/github.com/h2non/gentleman-consul)

[gentleman](https://github.com/h2non/gentleman)'s v2 plugin for easy service discovery using [Consul](https://www.consul.io).

Provides transparent retry/backoff support for resilient and [reactive](http://www.reactivemanifesto.org) HTTP client capabilities.  
It also allows you to use custom [retry strategies](#custom-retry-strategy), such as [constant](https://godoc.org/github.com/eapache/go-resiliency/retrier#ConstantBackoff) or [exponential](https://godoc.org/github.com/eapache/go-resiliency/retrier#ExponentialBackoff) retries.

## Installation

```bash
go get -u gopkg.in/h2non/gentleman-consul.v2
```

## Versions

- **[v1](https://github.com/h2non/gentleman-consul/tree/v1)** - First version, uses `gentleman@v1`.
- **[v2](https://github.com/h2non/gentleman-consul/tree/master)** - Latest version, uses `gentleman@v2`.

## API

See [godoc reference](https://godoc.org/github.com/h2non/gentleman-consul) for detailed API documentation.

## Examples

See [examples](https://github.com/h2non/gentleman-consul/blob/master/_examples) directory for featured examples.

#### Simple request

```go
package main

import (
  "fmt"

  "gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman-consul.v2"
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

  // Print response info
  fmt.Printf("Server URL: %s\n", res.RawRequest.URL.String())
  fmt.Printf("Response status: %d\n", res.StatusCode)
  fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
```

#### Custom retry strategy

```go
package main

import (
  "fmt"
	"time"

	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gentleman-consul.v1"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
)

func main() {
  // Create a new client
  cli := gentleman.New()

  // Configure Consul plugin
  config := consul.NewConfig("demo.consul.io", "web")

  // Use a custom retrier strategy with max 10 retry attempts
  config.Retrier = retrier.New(retrier.ConstantBackoff(10, time.Duration(25*time.Millisecond)), nil)

  // Register Consul's plugin at client level
  cli.Use(consul.New(config))

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

  // Print response info
  fmt.Printf("Server URL: %s\n", res.RawRequest.URL.String())
  fmt.Printf("Response status: %d\n", res.StatusCode)
  fmt.Printf("Server header: %s\n", res.Header.Get("Server"))
}
```

## License

MIT - Tomas Aparicio
