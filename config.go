package consul

import (
	"github.com/hashicorp/consul/api"
	"gopkg.in/h2non/gentleman-retry.v0"
	"time"
)

// Scheme represents the URI scheme used by default.
var Scheme = "http"

// DefaultConfig provides a custom
var DefaultConfig = api.DefaultConfig

// RefreshTTL stores the default Consul catalog refresh cycle TTL.
// Default to 2 minutes.
var RefreshTTL = time.Duration(2) * time.Minute

// Config represents the plugin supported settings.
type Config struct {
	Retry      bool
	Cache      bool
	Service    string
	Tag        string
	Scheme     string
	Retrier    retry.Retrier
	RefreshTTL time.Duration
	Client     *api.Config
	Query      *api.QueryOptions
}

// NewConfig creates a new plugin with default settings and
// custom Consul server URL and service name.
func NewConfig(addr, service string) *Config {
	config := api.DefaultConfig()
	config.Address = addr
	return &Config{
		Retry:      true,
		Cache:      true,
		Service:    service,
		Client:     config,
		Scheme:     Scheme,
		RefreshTTL: RefreshTTL,
		Retrier:    DefaultRetrier,
	}
}
