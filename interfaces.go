package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

var (
	ErrInterfaceAlreadyExists = errors.New("interface already exists")
	ErrInterfaceNotFound      = errors.New("interface not found")
)

type Interface struct {
	Zone string
	Name string
}

func (i Interface) Compare(other Interface) int {
	if cmp := strings.Compare(i.Zone, other.Zone); cmp != 0 {
		return cmp
	}
	// This should not be needed, as interface names are unique within zones
	return strings.Compare(i.Name, other.Name)
}

func (i Interface) Equals(other Interface) bool {
	return i.Zone == other.Zone && i.Name == other.Name
}

func (i Interface) Format() string {
	return fmt.Sprintf("%s\t%s", i.Zone, i.Name)
}

func Interfaces() ([]Interface, error) {
	buff, err := os.ReadFile(fullInterfacesFile)
	if err != nil {
		return nil, err
	}
	return getInterfacesBuff(buff)
}

func AddInterface(iface Interface) error {
	return readWriteFile(fullInterfacesFile, addInterfaceBuff, iface)
}

func RemoveInterfaceByZone(zone string) error {
	return readWriteFile(fullInterfacesFile, removeInterfaceByZoneBuff, zone)
}

func getInterfacesBuff(buff []byte) ([]Interface, error) {
	return parseInterfaces(buff), nil
}

func addInterfaceBuff(buff []byte, iface Interface) ([]byte, error) {
	interfaces, err := getInterfacesBuff(buff)
	if err != nil {
		return nil, err
	}

	if slices.ContainsFunc(interfaces, func(z Interface) bool {
		return z.Name == iface.Name
	}) {
		return nil, ErrInterfaceAlreadyExists
	}

	return fmt.Appendf(buff, "%s\n", iface.Format()), nil
}

func removeInterfaceByZoneBuff(buff []byte, zone string) ([]byte, error) {
	interfaces, err := getInterfacesBuff(buff)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(interfaces, func(z Interface) bool {
		return z.Zone == zone
	})
	if index == -1 {
		return nil, ErrInterfaceNotFound
	}

	interfaces = slices.Delete(interfaces, index, index+1)

	var b bytes.Buffer
	for _, z := range interfaces {
		b.WriteString(fmt.Sprintf("%s\n", z.Format()))
	}

	return b.Bytes(), nil
}

func parseInterfaces(data []byte) (interfaces []Interface) {
	for z := range bytes.Lines(data) {
		z = bytes.TrimSpace(z)
		if len(z) == 0 || z[0] == '#' {
			continue
		}
		z = bytes.ReplaceAll(z, []byte("\t"), []byte(" "))
		parts := bytes.Fields(z)
		if len(parts) < 2 {
			continue
		}
		iface := Interface{
			Zone: string(parts[0]),
			Name: string(parts[1]),
		}
		interfaces = append(interfaces, iface)
	}
	return
}
