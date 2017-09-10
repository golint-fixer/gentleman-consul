package consul

import (
	"testing"

	"github.com/nbio/st"
	"gopkg.in/h2non/gentleman.v2/context"
)

func TestRetrier(t *testing.T) {
	consul := NewClient(NewConfig("consul.server", "foo"))
	retrier := &Retrier{Consul: consul, Context: context.New(), Retry: DefaultRetrier}

	calls := 0
	retrier.Run(func() error {
		calls++
		return nil
	})

	st.Expect(t, calls, 1)
}

func TestNewRetrier(t *testing.T) {
	consul := NewClient(NewConfig("consul.server", "foo"))
	retrier := NewRetrier(consul, context.New())

	calls := 0
	retrier.Run(func() error {
		calls++
		return nil
	})

	st.Expect(t, calls, 1)
}
