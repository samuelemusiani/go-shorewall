package goshorewall

import (
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSingleAppReadWriteFile(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "goshoreall_test-*")
	assert.NoError(t, err, "Creating tmp file")
	err = f.Close()
	assert.NoError(t, err, "Closing file")

	lockDirPath = "/tmp/.goshorewall"

	id := uuid.New()
	iface := Interface{
		Zone: "test",
		Name: "eth0",
	}

	interfaces, err := execGetWithLock("interfaces", id.String(), f.Name(), getInterfacesBuff)
	assert.NoError(t, err, "Getting interfaces")
	assert.Equal(t, 0, len(interfaces), "Expected 0 interfaces")

	err = execAddRemoveWithLock("interfaces", id.String(), f.Name(), addInterfaceBuff, iface)
	assert.NoError(t, err, "Adding interface")

	interfaces, err = execGetWithLock("interfaces", id.String(), f.Name(), getInterfacesBuff)
	assert.NoError(t, err, "Getting interfaces")
	assert.Equal(t, 1, len(interfaces), "Expected one interface")
	assert.Equal(t, iface, interfaces[0], "Expected interface to match")
}
