package goshorewall

import (
	"bytes"
	"errors"
	"os"
	"slices"
)

type Interface struct {
	Zone string
	Name string
}

var (
	ErrInterfaceAlreadyExists = errors.New("interface already exists")
	ErrInterfaceNotFound      = errors.New("interface not found")
)

func parseInterfaces(data []byte) (interfaces []Interface) {
	iter := bytes.Lines(data)
	for z := range iter {
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

func GetInterfaces() ([]Interface, error) {
	buff, err := os.ReadFile(interfacesFile)
	if err != nil {
		return nil, err
	}

	return parseInterfaces(buff), nil
}

func AddInterface(iface Interface) error {
	interfaces, err := GetInterfaces()
	if err != nil {
		return err
	}

	for _, z := range interfaces {
		if z.Name == iface.Name {
			return ErrInterfaceAlreadyExists
		}
	}

	f, err := os.OpenFile(interfacesFile, os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(iface.Zone + "\t" + iface.Name + "\n"); err != nil {
		return err
	}

	return nil
}

func RemoveInterfaceByZone(zone string) error {
	interfaces, err := GetInterfaces()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(interfaces, func(z Interface) bool {
		return z.Zone == zone
	})
	if index == -1 {
		return ErrInterfaceNotFound
	}

	interfaces = append(interfaces[:index], interfaces[index+1:]...)

	var buffer bytes.Buffer
	for _, z := range interfaces {
		buffer.WriteString(z.Zone + "\t" + z.Name + "\n")
	}

	if err = os.WriteFile(interfacesFile, buffer.Bytes(), 0600); err != nil {
		return err
	}

	return nil
}
