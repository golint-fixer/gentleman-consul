package consul

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/hashicorp/consul/api"
	"gopkg.in/h2non/gentleman-retry.v2"
	"gopkg.in/h2non/gentleman.v2/context"
	"gopkg.in/h2non/gentleman.v2/plugin"
)

// Consul represents the Consul plugin adapter for gentleman,
// which encapsulates the official Consul client and plugin specific settings.
type Consul struct {
	// mutex is used internallt to avoid race coditions for multithread scenarios.
	sync.Mutex

	// updated stores the last Consul update date.
	updated time.Time

	// cache is used to store and cache the servers.
	cache []string

	// Config stores the Consul's plugin specific settings.
	Config *Config

	// Client stores the official Consul client.
	Client *api.Client
}

// New creates a new Consul client with the given config
// and returns the gentleman plugin.
func New(config *Config) plugin.Plugin {
	return NewClient(config).Plugin()
}

// NewClient creates a new Consul high-level client.
func NewClient(config *Config) *Consul {
	client, _ := api.NewClient(config.Client)
	return &Consul{Config: config, Client: client}
}

// Plugin returns the gentleman plugin to be plugged.
func (c *Consul) Plugin() plugin.Plugin {
	handlers := plugin.Handlers{"before dial": c.OnBeforeDial}
	return &plugin.Layer{Handlers: handlers}
}

// IsUpdated returns true if the current list of catalog services is up-to-date,
// based on the cache TTL.
func (c *Consul) IsUpdated() bool {
	return len(c.cache) > 0 && time.Duration((time.Now().UnixNano()-c.updated.UnixNano())) < c.Config.CacheTTL
}

// UpdateCache updates the list of catalog services.
func (c *Consul) UpdateCache(nodes []string) {
	if !c.Config.Cache || len(nodes) == 0 {
		return
	}

	c.updated = time.Now()
	c.cache = nodes
}

// GetNodes returns a list of nodes for the current service from Consul server
// or from cache (if enabled and not expired).
func (c *Consul) GetNodes() ([]string, error) {
	c.Lock()
	defer c.Unlock()

	if c.IsUpdated() {
		return c.cache, nil
	}

	entries, _, err := c.Client.Health().Service(c.Config.Service, c.Config.Tag, true, c.Config.Query)
	if err != nil {
		return nil, err
	}

	nodes := makeInstances(entries)
	c.UpdateCache(nodes)

	return nodes, nil
}

func makeInstances(entries []*api.ServiceEntry) []string {
	instances := make([]string, len(entries))

	for i, entry := range entries {
		addr := entry.Node.Address
		if entry.Service.Address != "" {
			addr = entry.Service.Address
		}
		instances[i] = fmt.Sprintf("%s:%d", addr, entry.Service.Port)
	}

	return instances
}

// SetServerURL sets the request URL fields based on the given Consul service instance.
func (c *Consul) SetServerURL(ctx *context.Context, host string) {
	// Define server URL based on the best node
	ctx.Request.URL.Scheme = c.Config.Scheme
	ctx.Request.URL.Host = host
}

// GetBestCandidateNode retrieves and returns the best service node candidate
// asking to Consul server catalog or reading catalog from cache.
func (c *Consul) GetBestCandidateNode(ctx *context.Context) (string, error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return "", err
	}
	if len(nodes) == 0 {
		return "", errors.New("consul: missing servers for service: " + c.Config.Service)
	}

	index := 0
	if retries, ok := ctx.Get("$consul.retries").(int); ok {
		index = retries
	}

	if index < len(nodes) {
		return nodes[index], nil
	}

	return nodes[0], nil
}

// UseBestCandidateNode sets the best service node URL in the given gentleman context.
func (c *Consul) UseBestCandidateNode(ctx *context.Context) error {
	node, err := c.GetBestCandidateNode(ctx)
	if err != nil {
		return err
	}

	// Define the proper URL in the outgoing request
	c.SetServerURL(ctx, node)
	return nil
}

// OnBeforeDial is a middleware function handler that replaces
// the outgoing request URL and provides a new http.RoundTripper if necessary
// in order to handle request failures and retry it accordingly.
func (c *Consul) OnBeforeDial(ctx *context.Context, h context.Handler) {
	// Define the server retries
	ctx.Set("$consul.retries", 0)

	// Use best service node candidate
	err := c.UseBestCandidateNode(ctx)
	if err != nil {
		h.Error(ctx, err)
		return
	}

	// Always continue with the next middleware
	defer h.Next(ctx)

	// If retry is disable, just continue
	if !c.Config.Retry {
		return
	}

	// Wrap HTTP transport with Consul retrier, if enabled
	retrier := NewRetrier(c, ctx)
	if c.Config.Retrier != nil {
		retrier.Retry = c.Config.Retrier
	}
	retry.InterceptTransport(ctx, retrier)
}
