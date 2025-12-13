package goshorewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const interfaces01 = `
#ZONE	    INTERFACE
out	    eth0
management  eth2.100
h	    eth2.90     
h4	    eth2.99     
g       eth2.80     
l	    eth3.50
li	    eth3.55
pve	    eth3.78
vpn         wg0
relay       eth5
`

func TestGetNamesBuff(t *testing.T) {
	interfaces, err := getInterfacesBuff([]byte(interfaces01))
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 10, len(interfaces), "expected 9 interfaces")
	assert.Equal(t, "out", interfaces[0].Zone)
	assert.Equal(t, "eth0", interfaces[0].Name)

	assert.Equal(t, "management", interfaces[1].Zone)
	assert.Equal(t, "eth2.100", interfaces[1].Name)

	assert.Equal(t, "h", interfaces[2].Zone)
	assert.Equal(t, "eth2.90", interfaces[2].Name)

	assert.Equal(t, "h4", interfaces[3].Zone)
	assert.Equal(t, "eth2.99", interfaces[3].Name)

	assert.Equal(t, "g", interfaces[4].Zone)
	assert.Equal(t, "eth2.80", interfaces[4].Name)

	assert.Equal(t, "l", interfaces[5].Zone)
	assert.Equal(t, "eth3.50", interfaces[5].Name)

	assert.Equal(t, "li", interfaces[6].Zone)
	assert.Equal(t, "eth3.55", interfaces[6].Name)

	assert.Equal(t, "pve", interfaces[7].Zone)
	assert.Equal(t, "eth3.78", interfaces[7].Name)

	assert.Equal(t, "vpn", interfaces[8].Zone)
	assert.Equal(t, "wg0", interfaces[8].Name)

	assert.Equal(t, "relay", interfaces[9].Zone)
	assert.Equal(t, "eth5", interfaces[9].Name)
}

func TestAddInterfaceBuff(t *testing.T) {
	buff, err := addInterfaceBuff([]byte(interfaces01), Interface{Zone: "dmz", Name: "eth4"})
	assert.NoError(t, err, "expected no error")
	interfaces, err := getInterfacesBuff(buff)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 11, len(interfaces), "expected 11 interfaces")
	assert.Equal(t, "dmz", interfaces[10].Zone)
	assert.Equal(t, "eth4", interfaces[10].Name)
}

func TestRemoveInterfaceByZoneBuff(t *testing.T) {
	buff, err := removeInterfaceByZoneBuff([]byte(interfaces01), "h4")
	assert.NoError(t, err, "expected no error")
	interfaces, err := getInterfacesBuff(buff)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 9, len(interfaces), "expected 9 interfaces")

	for _, iface := range interfaces {
		assert.NotEqual(t, "h4", iface.Zone, "expected 'h4' interface to be removed")
	}
}

func TestAddInterfaceBuff_AlreadyExists(t *testing.T) {
	_, err := addInterfaceBuff([]byte(interfaces01), Interface{Zone: "out", Name: "eth0"})
	assert.ErrorIs(t, err, ErrInterfaceAlreadyExists, "expected ErrInterfaceAlreadyExists")
}

func TestRemoveInterfaceByZoneBuff_NotFound(t *testing.T) {
	_, err := removeInterfaceByZoneBuff([]byte(interfaces01), "dmz")
	assert.ErrorIs(t, err, ErrInterfaceNotFound, "expected ErrInterfaceNotFound")
}
