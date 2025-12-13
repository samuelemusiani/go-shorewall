package goshorewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const snat01 = `
#ACTION		SOURCE		DEST		PROTO
MASQUERADE	172.16.0.0/12	eth0
MASQUERADE	10.0.0.0/8	eth0
MASQUERADE	192.168.0.0/16	eth0
`

func TestGetSnatBuff(t *testing.T) {
	snats, err := getSnatsBuff([]byte(snat01))
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 3, len(snats), "expected 3 snat rules")

	assert.Equal(t, "MASQUERADE", snats[0].Action)
	assert.Equal(t, "172.16.0.0/12", snats[0].Source)
	assert.Equal(t, "eth0", snats[0].Destination)

	assert.Equal(t, "MASQUERADE", snats[1].Action)
	assert.Equal(t, "10.0.0.0/8", snats[1].Source)
	assert.Equal(t, "eth0", snats[1].Destination)

	assert.Equal(t, "MASQUERADE", snats[2].Action)
	assert.Equal(t, "192.168.0.0/16", snats[2].Source)
	assert.Equal(t, "eth0", snats[2].Destination)
}

func TestAddSnatBuff(t *testing.T) {
	newSnat := Snat{Action: "MASQUERADE", Source: "100.90.0.0/23", Destination: "eth1"}
	buff, err := addSnatBuff([]byte(snat01), newSnat)
	assert.NoError(t, err, "expected no error")

	snats, err := getSnatsBuff(buff)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 4, len(snats), "expected 4 snat rules")

	assert.Equal(t, newSnat.Action, snats[3].Action)
	assert.Equal(t, newSnat.Source, snats[3].Source)
	assert.Equal(t, newSnat.Destination, snats[3].Destination)
}

func TestAddSnatBuff_AlreadyExists(t *testing.T) {
	newSnat := Snat{Action: "MASQUERADE", Source: "192.168.0.0/16", Destination: "eth0"}
	_, err := addSnatBuff([]byte(snat01), newSnat)
	assert.ErrorIs(t, err, ErrSnatAlreadyExists, "expected ErrSnatAlreadyExists")
}

func TestRemoveSnatBuff(t *testing.T) {
	snatToRemove := Snat{Action: "MASQUERADE", Source: "10.0.0.0/8", Destination: "eth0"}
	buff, err := removeSnatBuff([]byte(snat01), snatToRemove)
	assert.NoError(t, err, "expected no error")

	snats, err := getSnatsBuff(buff)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 2, len(snats), "expected 2 snat rules")

	assert.Equal(t, "MASQUERADE", snats[0].Action)
	assert.Equal(t, "172.16.0.0/12", snats[0].Source)
	assert.Equal(t, "eth0", snats[0].Destination)

	assert.Equal(t, "MASQUERADE", snats[1].Action)
	assert.Equal(t, "192.168.0.0/16", snats[1].Source)
	assert.Equal(t, "eth0", snats[1].Destination)
}

func TestRemoveSnatBuff_NotFound(t *testing.T) {
	snatToRemove := Snat{Action: "MASQUERADE", Source: "203.0.113.0/24", Destination: "eth2"}
	_, err := removeSnatBuff([]byte(snat01), snatToRemove)
	assert.ErrorIs(t, err, ErrSnatNotFound, "expected ErrSnatNotFound")

	snatToRemove2 := Snat{Action: "SNAT", Source: "192.168.0.0/16", Destination: "eth0"}
	_, err = removeSnatBuff([]byte(snat01), snatToRemove2)
	assert.ErrorIs(t, err, ErrSnatNotFound, "expected ErrSnatNotFound")
}

func TestParseSnat(t *testing.T) {
	data := []byte(`
#ACTION		SOURCE		DEST		PROTO
MASQUERADE	172.16.0.0/12	eth0
SNAT		10.0.0.0/8	eth1
`)
	snats := parseSnats(data)
	assert.Equal(t, 2, len(snats), "expected 2 snat rules")
	assert.Equal(t, "MASQUERADE", snats[0].Action)
	assert.Equal(t, "172.16.0.0/12", snats[0].Source)
	assert.Equal(t, "eth0", snats[0].Destination)
	assert.Equal(t, "SNAT", snats[1].Action)
	assert.Equal(t, "10.0.0.0/8", snats[1].Source)
	assert.Equal(t, "eth1", snats[1].Destination)
}
