package consul

import (
	"time"

	"github.com/hashicorp/consul/api"
	"gopkg.in/h2non/gentleman-retry.v2"
)

// Scheme represents the URI scheme used by default.
var Scheme = "http"

// DefaultConfig provides a custom
var DefaultConfig = api.DefaultConfig

// CacheTTL stores the default Consul catalog refresh cycle TTL.
// Default to 10 minutes.
var CacheTTL = 10 * time.Minute

// Config represents the plugin supported settings.
type Config struct {
	// Retry enables/disables HTTP request retry policy. Defaults to true.
	Retry bool

	// Cache enables/disables the Consul catalog internal cache
	// avoiding recurrent request to Consul server.
	Cache bool

	// Service stores the Consul's service name identifier. E.g: web.
	Service string

	// Tag stores the optional Consul's service tag to use when asking to Consul server.
	Tag string

	// Scheme stores the default HTTP URI scheme to be used when asking to Consul server.
	// Defaults to: http.
	Scheme string

	// Retrier stores the retry strategy to be used.
	// Defaults to: ContanstBackOff with max 3 retries.
	Retrier retry.Retrier

	// CacheTTL stores the max Consul catalog cache TTL.
	CacheTTL time.Duration

	// Client stores the official Consul client Config instance.
	Client *api.Config

	// Query stores the official Consul client query options when asking to Consul server.
	Query *api.QueryOptions
}

// NewConfig creates a new plugin with default settings and
// custom Consul server URL and service name.
func NewConfig(server, service string) *Config {
	config := api.DefaultConfig()
	config.Address = server
	return &Config{
		Retry:    true,
		Cache:    true,
		Service:  service,
		Client:   config,
		Scheme:   Scheme,
		CacheTTL: CacheTTL,
		Retrier:  DefaultRetrier,
	}
}

// NewBasicConfig creates a new basic default config with the given Consul server hostname.
func NewBasicConfig(server string) *Config {
	return NewConfig(server, "")
}
