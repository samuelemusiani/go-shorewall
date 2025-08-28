package goshorewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseZones(t *testing.T) {
	data := []byte(`
#ZONE		TYPE
zone1		ip
zone2		ip
`)
	zones := parseZones(data)
	assert.Equal(t, 2, len(zones), "expected 2 zones")
	assert.Equal(t, "zone1", zones[0].Name)
	assert.Equal(t, "ip", zones[0].Type)
	assert.Equal(t, "zone2", zones[1].Name)
	assert.Equal(t, "ip", zones[1].Type)
}
