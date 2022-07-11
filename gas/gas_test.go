package gas

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetGasPrice(t *testing.T) {
	g := New(8)
	defer close(g.Ch)
	price := g.GetSafeLow()
	require.NotZero(t, price)
}
