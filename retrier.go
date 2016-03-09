package consul

import (
	"gopkg.in/h2non/gentleman-retry.v0"
	"gopkg.in/h2non/gentleman.v0/context"
)

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

// Run runs the given function multiple times, acting like a proxy
// to user defined retry strategy.
func (r *Retrier) Run(fn func() error) error {
	return r.Retry.Run(func() error {
		retries := r.Context.Get("$consul.retries").(int)

		// Call the function directly for the first attempt
		if retries == 0 {
			return fn()
		}

		r.Context.Set("$consul.retries", retries+1)
		err := r.Consul.SetBestCandidateNode(r.Context)
		if err != nil {
			return err
		}

		return fn()
	})
}
