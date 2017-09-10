package consul

import (
	"testing"
	"time"

	"github.com/nbio/st"
	"gopkg.in/eapache/go-resiliency.v1/retrier"
	"gopkg.in/h2non/gentleman-mock.v2"
	"gopkg.in/h2non/gentleman.v2"
	"gopkg.in/h2non/gock.v1"
)

const consulValidResponse = `
[
  {
    "Node":{
      "Node":"consul-client-nyc3-1",
      "Address":"127.0.0.1",
      "TaggedAddresses":{
        "wan":"127.0.0.1"
      },
      "CreateIndex":7,
      "ModifyIndex":375588
    },
    "Service":{
      "ID":"web",
      "Service":"web",
      "Tags":null,
      "Address":"",
      "Port":80,
      "EnableTagOverride":false,
      "CreateIndex":13,
      "ModifyIndex":13
    }
  }
]`

func TestConsulClient(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/health/service/web").
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
		Get("/v1/health/service/web").
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
		Get("/v1/health/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Times(9).
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
		Get("/v1/health/service/web").
		Times(4).
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

func TestConsulRetryDisabled(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	config.Retry = false
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/health/service/web").
		Reply(200).
		Type("json").
		BodyString(consulValidResponse)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(503)

	cli := gentleman.New()
	cli.Use(mock.Plugin)
	cli.Use(consul)

	res, err := cli.Request().Send()
	st.Expect(t, err, nil)
	st.Expect(t, res.StatusCode, 503)
	st.Expect(t, res.String(), "")
	st.Expect(t, gock.IsPending(), false)
}

const consulMultipleServers = `
[
  {
    "Node":{
      "Node":"consul-client-nyc3-1",
      "Address":"127.0.0.1",
      "TaggedAddresses":{
        "wan":"127.0.0.1"
      },
      "CreateIndex":7,
      "ModifyIndex":375588
    },
    "Service":{
      "ID":"web",
      "Service":"web",
      "Tags":null,
      "Address":"",
      "Port":80,
      "EnableTagOverride":false,
      "CreateIndex":13,
      "ModifyIndex":13
    }
  },
  {
    "Node":{
      "Node":"consul-client-nyc3-1",
      "Address":"127.0.0.2",
      "TaggedAddresses":{
        "wan":"127.0.0.2"
      },
      "CreateIndex":7,
      "ModifyIndex":375588
    },
    "Service":{
      "ID":"web",
      "Service":"web",
      "Tags":null,
      "Address":"",
      "Port":80,
      "EnableTagOverride":false,
      "CreateIndex":13,
      "ModifyIndex":13
    }
  },
  {
    "Node":{
      "Node":"consul-client-nyc3-1",
      "Address":"127.0.0.3",
      "TaggedAddresses":{
        "wan":"127.0.0.3"
      },
      "CreateIndex":7,
      "ModifyIndex":375588
    },
    "Service":{
      "ID":"web",
      "Service":"web",
      "Tags":null,
      "Address":"",
      "Port":80,
      "EnableTagOverride":false,
      "CreateIndex":13,
      "ModifyIndex":13
    }
  }
]`

func TestConsulNextServerFallback(t *testing.T) {
	defer gock.Off()

	config := NewConfig("demo.consul.io", "web")
	consul := New(config)
	gock.InterceptClient(config.Client.HttpClient)

	gock.New("http://demo.consul.io").
		Get("/v1/health/service/web").
		Reply(200).
		Type("json").
		BodyString(consulMultipleServers)

	gock.New("http://127.0.0.1:80").
		Get("/").
		Reply(503)

	gock.New("http://127.0.0.2:80").
		Get("/").
		Reply(503)

	gock.New("http://127.0.0.3:80").
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
