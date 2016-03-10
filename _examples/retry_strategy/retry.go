package main

import (
	"fmt"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
	"gopkg.in/h2non/gentleman-consul.v0"
	"gopkg.in/h2non/gentleman-mock.v0"
	"gopkg.in/h2non/gentleman.v0"
	"gopkg.in/h2non/gock.v0"
	"time"
)

const consulValidResponse = `
[
  {
    "Node":"consul-client-nyc3-1",
    "Address":"127.0.0.1",
    "ServiceID":"web",
    "ServiceName":"web",
    "ServiceTags":[],
    "ServiceAddress":"",
    "ServicePort":80,
    "ServiceEnableTagOverride":false,
    "CreateIndex":17,
    "ModifyIndex":17
  }
]`

func main() {
	defer gock.Off()

	// Mock consul server
	gock.New("http://demo.consul.io").
		Get("/v1/catalog/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	// Configure failure responses
	gock.New("http://127.0.0.1:80").
		Get("/").
		Times(10).
		Reply(503)

	// Final valid response
	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(200).
		SetHeader("Server", "gock").
		BodyString("hello world")

	// Create a new client
	cli := gentleman.New()

	// Configure Consul plugin
	config := consul.NewConfig("demo.consul.io", "web")

	// Use a custom retrier strategy with max 10 retry attempts
	config.Retrier = retrier.New(retrier.ConstantBackoff(10, time.Duration(25*time.Millisecond)), nil)

	// Intercept HTTP transport via gock to simulate the failures
	gock.InterceptClient(config.Client.HttpClient)
	cli.Use(mock.Plugin)

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
