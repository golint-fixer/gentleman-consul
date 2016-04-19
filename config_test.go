package consul

import (
	"testing"

	"github.com/nbio/st"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig("server", "foo")
	st.Expect(t, config.Client.Address, "server")
	st.Expect(t, config.Service, "foo")
	st.Expect(t, config.Scheme, Scheme)
	st.Expect(t, config.Cache, true)
	st.Expect(t, config.Retry, true)
}

func TestNewBasicConfig(t *testing.T) {
	config := NewBasicConfig("server")
	st.Expect(t, config.Client.Address, "server")
	st.Expect(t, config.Service, "")
	st.Expect(t, config.Scheme, Scheme)
	st.Expect(t, config.Cache, true)
	st.Expect(t, config.Retry, true)
}
