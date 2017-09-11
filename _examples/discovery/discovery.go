package main

import (
	"fmt"

	"gopkg.in/h2non/gentleman-consul.v2"
	"gopkg.in/h2non/gentleman.v2"
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
