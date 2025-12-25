package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Zone struct {
	Name string
	Type string
}

var (
	ErrZoneAlreadyExists = errors.New("zone already exists")
	ErrZoneNotFound      = errors.New("zone not found")
)

func Zones() ([]Zone, error) {
	buff, err := os.ReadFile(fullZonesFile)
	if err != nil {
		return nil, err
	}
	return getZonesBuff(buff)
}

func AddZone(zone Zone) error {
	return readWriteFile(fullZonesFile, addZoneBuff, zone)
}

func RemoveZone(zoneName string) error {
	return readWriteFile(fullZonesFile, removeZoneBuff, zoneName)
}

func getZonesBuff(buff []byte) ([]Zone, error) {
	return parseZones(buff), nil
}

func addZoneBuff(buff []byte, zone Zone) ([]byte, error) {
	zones, err := getZonesBuff(buff)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(zones, func(a, b Zone) int {
		return strings.Compare(a.Name, b.Name)
	})

	if slices.ContainsFunc(zones, func(z Zone) bool {
		return z.Name == zone.Name
	}) {
		return nil, ErrZoneAlreadyExists
	}

	return fmt.Appendf(buff, "%s\t%s\n", zone.Name, zone.Type), nil
}

func removeZoneBuff(buff []byte, zoneName string) ([]byte, error) {
	zones, err := getZonesBuff(buff)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(zones, func(z Zone) bool {
		return z.Name == zoneName
	})
	if index == -1 {
		return nil, ErrZoneNotFound
	}

	zones = slices.Delete(zones, index, index+1)

	var b bytes.Buffer
	for _, z := range zones {
		b.WriteString(z.Name + "\t" + z.Type + "\n")
	}

	return b.Bytes(), nil
}

func parseZones(data []byte) (zones []Zone) {
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
		zone := Zone{
			Name: string(parts[0]),
			Type: string(parts[1]),
		}
		zones = append(zones, zone)
	}
	return
}
