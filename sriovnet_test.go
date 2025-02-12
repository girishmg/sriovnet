package sriovnet

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	utilfs "github.com/Mellanox/sriovnet/pkg/utils/filesystem"
)

const (
	fakeFsRoot = "/tmp/sriovnet-tests"
)

func setupFakeFs(t *testing.T) func() {
	var err error
	var teardown func()
	utilfs.Fs, teardown, err = utilfs.NewFakeFs(fakeFsRoot)
	if err != nil {
		t.Errorf("setupFakeFs: Failed to create fake FS %v", err)
	}
	return teardown
}

func setupGetNetDevicesFromPciEnv(t *testing.T, pciAddress string, deviceNames []string) func() {
	var err error
	teardown := setupFakeFs(t)
	if len(deviceNames) > 0 {
		pciNetDir := filepath.Join(PciSysDir, pciAddress, "net")
		err = utilfs.Fs.MkdirAll(pciNetDir, os.FileMode(0755))
		defer func() {
			if err != nil {
				teardown()
				t.Errorf("setupGetNetDevicesFromPciEnv, got %v", err)
			}
		}()
		for _, deviceName := range deviceNames {
			deviceNamePath := filepath.Join(pciNetDir, deviceName)
			err = utilfs.Fs.MkdirAll(deviceNamePath, os.FileMode(0755))
		}
	} else {
		pciNetDir := filepath.Join(PciSysDir, pciAddress)
		err = utilfs.Fs.MkdirAll(pciNetDir, os.FileMode(0755))
		if err != nil {
			teardown()
			t.Errorf("setupGetNetDevicesFromPciEnv, got %v", err)
		}
	}
	return teardown
}

func TestGetNetDevicesFromPciSuccess(t *testing.T) {
	pciAddress := "0000:02:00.0"
	deviceNames := []string{"enp0s0f0", "enp0s0f1", "enp0s0f2"}
	teardown := setupGetNetDevicesFromPciEnv(t, pciAddress, deviceNames)
	defer teardown()
	devNames, err := GetNetDevicesFromPci(pciAddress)
	assert.NoError(t, err)
	assert.Equal(t, deviceNames, devNames)
}

func TestGetNetDevicesFromPciErrorNoPCI(t *testing.T) {
	teardown := setupFakeFs(t)
	defer teardown()
	pciAddress := "0000:02:00.0"
	devNames, err := GetNetDevicesFromPci(pciAddress)
	assert.Error(t, err)
	assert.Equal(t, []string(nil), devNames)
}

func TestGetNetDevicesFromPciErrorNoDevices(t *testing.T) {
	var err error
	pciAddress := "0000:02:00.0"
	deviceNames := []string{}
	teardown := setupGetNetDevicesFromPciEnv(t, pciAddress, deviceNames)
	defer teardown()
	devNames, err := GetNetDevicesFromPci(pciAddress)
	assert.Error(t, err)
	assert.Equal(t, []string(nil), devNames)
}

func TestGetPfPciFromVfPci(t *testing.T) {
	teardown := setupFakeFs(t)
	defer teardown()
	// Create PCI sysfs layout with FakfeFs
	pfPciAddr := "0000:02:00.0"
	pfPciPath := filepath.Join(PciSysDir, pfPciAddr)
	vfPciAddr := "0000:02:00.6"
	vfPciPath := filepath.Join(PciSysDir, vfPciAddr)

	// PF PCI path
	_ = utilfs.Fs.MkdirAll(pfPciPath, os.FileMode(0755))
	// VF PCI path and physfn link
	_ = utilfs.Fs.MkdirAll(vfPciPath, os.FileMode(0755))
	_ = utilfs.Fs.Symlink(pfPciPath, filepath.Join(vfPciPath, "physfn"))

	pf, err := GetPfPciFromVfPci(vfPciAddr)
	assert.NoError(t, err)
	assert.Equal(t, pfPciAddr, pf)
}

func TestGetPfPciFromVfPciError(t *testing.T) {
	teardown := setupFakeFs(t)
	defer teardown()
	// Create PCI sysfs layout with FakfeFs
	pciAddr := "0000:02:00.0"
	pciPath := filepath.Join(PciSysDir, pciAddr)

	// PCI path
	_ = utilfs.Fs.MkdirAll(pciPath, os.FileMode(0755))

	pf, err := GetPfPciFromVfPci(pciAddr)
	assert.Error(t, err)
	assert.Equal(t, "", pf)
}
