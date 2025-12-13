package goshorewall

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const zones01 = `
#ZONE		TYPE
a		ip
b 		ip
c		ip
`

const zones02 = `
#ZONE		TYPE
	akdslfasodf ip
		23eficnwiudc9e firewall
	asczxy ip
`

func TestGetZonesBuff(t *testing.T) {
	zones, err := getZonesBuff([]byte(zones01))
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 3, len(zones), "expected 3 zones")
	assert.Equal(t, "a", zones[0].Name)
	assert.Equal(t, "ip", zones[0].Type)
	assert.Equal(t, "b", zones[1].Name)
	assert.Equal(t, "ip", zones[1].Type)
	assert.Equal(t, "c", zones[2].Name)
	assert.Equal(t, "ip", zones[2].Type)
}

func TestAddZoneBuff(t *testing.T) {
	buff, err := addZoneBuff([]byte(zones01), Zone{Name: "d", Type: "ip"})
	assert.NoError(t, err, "expected no error")
	zones, err := getZonesBuff(buff)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 4, len(zones), "expected 4 zones")
	assert.Equal(t, "d", zones[3].Name)
	assert.Equal(t, "ip", zones[3].Type)
}

func TestAddZoneBuff_AlreadyExists(t *testing.T) {
	_, err := addZoneBuff([]byte(zones01), Zone{Name: "b", Type: "ip"})
	assert.ErrorIs(t, err, ErrZoneAlreadyExists, "expected ErrZoneAlreadyExists")

	_, err = addZoneBuff([]byte(zones02), Zone{Name: "23eficnwiudc9e", Type: "firewall"})
	assert.ErrorIs(t, err, ErrZoneAlreadyExists, "expected ErrZoneAlreadyExists")
}

func TestRemoveZoneBuff(t *testing.T) {
	buff, err := removeZoneBuff([]byte(zones01), "b")
	assert.NoError(t, err, "expected no error")
	zones, err := getZonesBuff(buff)
	assert.NoError(t, err, "expected no error")
	assert.Equal(t, 2, len(zones), "expected 2 zones")
	assert.Equal(t, "a", zones[0].Name)
	assert.Equal(t, "ip", zones[0].Type)
	assert.Equal(t, "c", zones[1].Name)
	assert.Equal(t, "ip", zones[1].Type)
}

func TestRemoveZoneBuff_NotFound(t *testing.T) {
	_, err := removeZoneBuff([]byte(zones01), "d")
	assert.ErrorIs(t, err, ErrZoneNotFound, "expected ErrZoneNotFound")

	_, err = removeZoneBuff([]byte(zones02), "nonexistentzone")
	assert.ErrorIs(t, err, ErrZoneNotFound, "expected ErrZoneNotFound")
}

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
