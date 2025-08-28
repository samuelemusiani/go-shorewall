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
)

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

func GetZones() ([]Zone, error) {
	buff, err := os.ReadFile(zonesFile)
	if err != nil {
		return nil, err
	}

	return parseZones(buff), nil
}

func AddZone(zone Zone) error {
	zones, err := GetZones()
	if err != nil {
		return err
	}

	slices.SortFunc(zones, func(a, b Zone) int {
		return strings.Compare(a.Name, b.Name)
	})

	if slices.ContainsFunc(zones, func(z Zone) bool {
		return z.Name == zone.Name
	}) {
		return ErrZoneAlreadyExists
	}

	f, err := os.OpenFile(zonesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = fmt.Fprintf(f, "%s\t%s\n", zone.Name, zone.Type)
	if err != nil {
		return err
	}

	return nil
}

func RemoveZone(zoneName string) error {
	zones, err := GetZones()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(zones, func(z Zone) bool {
		return z.Name == zoneName
	})
	if index == -1 {
		return fmt.Errorf("zone %s not found", zoneName)
	}

	zones = append(zones[:index], zones[index+1:]...)

	var buff bytes.Buffer
	for _, z := range zones {
		fmt.Fprintf(&buff, "%s\t%s\n", z.Name, z.Type)
	}

	err = os.WriteFile(zonesFile, buff.Bytes(), 0644)
	if err != nil {
		return err
	}

	return nil
}
