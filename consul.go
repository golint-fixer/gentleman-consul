package consul

import (
	"errors"
	"github.com/hashicorp/consul/api"
	"gopkg.in/h2non/gentleman-retry.v0"
	"gopkg.in/h2non/gentleman.v0/context"
	"gopkg.in/h2non/gentleman.v0/plugin"
	"strconv"
	"sync"
	"time"
)

// Scheme represents the URI scheme used by default.
var Scheme = "http"

// DefaultConfig provides a custom
var DefaultConfig = api.DefaultConfig

// DefaultRetrier stores the default retry strategy used by the plugin.
// By default will use a constant retry strategy with a maximum of 3 retry attempts.
var DefaultRetrier = retry.ConstantBackoff

// RefreshTTL stores the default Consul catalog refresh cycle TTL.
// Default to 2 minutes.
var RefreshTTL = time.Duration(2) * time.Minute

// Consul represents the Consul plugin adapter for gentleman,
// which encapsulates the official Consul client and plugin specific settings.
type Consul struct {
	updated time.Time
	mutex   *sync.Mutex
	cache   []*api.CatalogService
	Client  *api.Client
	Config  *Config
}

// New creates a new Consul plugin with the given config.
func New(config *Config) plugin.Plugin {
	client, _ := api.NewClient(config.Client)
	consul := &Consul{
		Config: config,
		Client: client,
		mutex:  &sync.Mutex{},
	}
	return consul.Plugin()
}

// Plugin returns the gentleman plugin to be plugged.
func (c *Consul) Plugin() plugin.Plugin {
	handlers := plugin.Handlers{
		"before dial": c.OnBeforeDial,
	}
	return &plugin.Layer{Handlers: handlers}
}

// IsUpdated returns true if the current list of catalog services is up-to-date,
// based on the refresh TTL.
func (c *Consul) IsUpdated() bool {
	return len(c.cache) > 0 && time.Duration((time.Now().Unix()-c.updated.Unix())) < (c.Config.RefreshTTL*time.Second)
}

// UpdateCache updates the list of catalog services.
func (c *Consul) UpdateCache(nodes []*api.CatalogService) {
	if !c.Config.Cache || len(nodes) != 0 {
		return
	}

	c.mutex.Lock()
	c.updated = time.Now()
	c.cache = nodes
	c.mutex.Unlock()
}

// GetNodes returns a list of nodes for the current service from Consul server
// or from cache (if enabled and not expired).
func (c *Consul) GetNodes() ([]*api.CatalogService, error) {
	if c.IsUpdated() {
		return c.cache, nil
	}

	nodes, _, err := c.Client.Catalog().Service(c.Config.Service, c.Config.Tag, c.Config.Query)
	if err != nil {
		c.UpdateCache(nodes)
	}

	return nodes, err
}

// SetServerURL sets the request URL fields based on the given Consul service instance.
func (c *Consul) SetServerURL(ctx *context.Context, node *api.CatalogService) {
	// Define server URL based on the best node
	ctx.Request.URL.Scheme = c.Config.Scheme
	ctx.Request.URL.Host = node.Address

	// Define URL port, if neccessary
	if node.ServicePort != 0 {
		ctx.Request.URL.Host += ":" + strconv.Itoa(node.ServicePort)
	}
}

// GetBestCandidateNode retrieves and returns the best service node candidate
// asking to Consul server catalog or reading catalog from cache.
func (c *Consul) GetBestCandidateNode(ctx *context.Context) (*api.CatalogService, error) {
	nodes, err := c.GetNodes()
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, errors.New("consul: missing servers for service: " + c.Config.Service)
	}

	index := 0
	if c.Config.Cache {
		index = ctx.Get("$consul.retries").(int)
	}

	if node := nodes[index]; node != nil {
		return node, nil
	}

	return nodes[0], nil
}

// SetBestCandidateNode sets the best service node URL in the given gentleman context.
func (c *Consul) SetBestCandidateNode(ctx *context.Context) error {
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

	// Get best node candidate
	err := c.SetBestCandidateNode(ctx)
	if err != nil {
		h.Error(ctx, err)
		return
	}

	// Wrap HTTP transport with Consul retrier, if enabled
	if c.Config.Retry && c.Config.Retrier != nil {
		retrier := &Retrier{Consul: c, Retry: c.Config.Retrier, Context: ctx}
		retry.InterceptTransport(ctx, retrier)
	}

	// Continue with the next middleware
	h.Next(ctx)
}
