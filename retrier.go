package consul

import (
	"gopkg.in/h2non/gentleman-retry.v2"
	"gopkg.in/h2non/gentleman.v2/context"
)

// DefaultRetrier stores the default retry strategy used by the plugin.
// By default will use a constant retry strategy with a maximum of 3 retry attempts.
var DefaultRetrier = retry.ConstantBackoff

// Retrier provides a retry.Retrier capable interface that
// encapsulates Consul client and user defined strategy.
type Retrier struct {
	// Consul stores the Consul client wrapper instance.
	Consul *Consul

	// Context stores the HTTP current gentleman context.
	Context *context.Context

	// Retry stores the retry strategy to be used.
	Retry retry.Retrier
}

// NewRetrier creates a default retrier for the given Consul client and context.
func NewRetrier(c *Consul, ctx *context.Context) *Retrier {
	return &Retrier{Consul: c, Context: ctx, Retry: DefaultRetrier}
}

// Run runs the given function multiple times, acting like a proxy
// to user defined retry strategy.
func (r *Retrier) Run(fn func() error) error {
	return r.Retry.Run(func() error {
		retries := 0
		if val, ok := r.Context.Get("$consul.retries").(int); ok {
			retries = val
		}

		// Expose number of retries
		defer r.Context.Set("$consul.retries", retries+1)

		// Call the function directly for the first attempt
		if retries == 0 {
			return fn()
		}

		// Set best server candidate
		err := r.Consul.UseBestCandidateNode(r.Context)
		if err != nil {
			return err
		}

		return fn()
	})
}
