package coveragedemo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShippingCost_RegularSmallParcel(t *testing.T) {
	assert.Equal(t, 300, ShippingCost(5, false))
}
