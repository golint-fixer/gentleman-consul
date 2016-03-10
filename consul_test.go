package consul

import (
	"github.com/nbio/st"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
	"gopkg.in/h2non/gentleman-mock.v0"
	"gopkg.in/h2non/gentleman.v0"
	"gopkg.in/h2non/gock.v0"
	"testing"
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

func TestConsulClient(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/catalog/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(200).
		BodyString("hello world")

	cli := gentleman.New()
	cli.Use(mock.Plugin)
	cli.Use(consul)

	res, err := cli.Request().Send()
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)
	st.Expect(t, res.String(), "hello world")
}

func TestConsulRetry(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/catalog/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Times(3).
		Reply(503)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(200).
		BodyString("hello world")

	cli := gentleman.New()
	cli.Use(mock.Plugin)
	cli.Use(consul)

	res, err := cli.Request().Send()
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)
	st.Expect(t, res.String(), "hello world")
	st.Expect(t, gock.IsPending(), false)
}

func TestConsulRetryCustomStrategy(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	config.Retrier = retrier.New(retrier.ConstantBackoff(10, time.Duration(25*time.Millisecond)), nil)
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/catalog/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Times(10).
		Reply(503)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(200).
		BodyString("hello world")

	cli := gentleman.New()
	cli.Use(mock.Plugin)
	cli.Use(consul)

	res, err := cli.Request().Send()
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)
	st.Expect(t, res.String(), "hello world")
	st.Expect(t, gock.IsPending(), false)
}

func TestConsulDisableCache(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	config.Cache = false
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/catalog/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Times(3).
		Reply(503)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(200).
		BodyString("hello world")

	cli := gentleman.New()
	cli.Use(mock.Plugin)
	cli.Use(consul)

	res, err := cli.Request().Send()
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 200)
	st.Expect(t, res.String(), "hello world")
	st.Expect(t, gock.IsPending(), false)
}
