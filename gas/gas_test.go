package gas

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func TestGetGasPrice(t *testing.T) {
	price, err := SafeLow()
	require.NotZero(t, price)
	require.Nil(t, err)
}
