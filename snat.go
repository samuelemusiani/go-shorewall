package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
)

type Snat struct {
	Action      string
	Source      string
	Destination string
}

var (
	ErrSnatAlreadyExists = errors.New("snat rule already exists")
	ErrSnatNotFound      = errors.New("snat rule not found")
)

func parseSnats(data []byte) (snats []Snat) {
	iter := bytes.Lines(data)
	for z := range iter {
		z = bytes.TrimSpace(z)
		if len(z) == 0 || z[0] == '#' {
			continue
		}
		z = bytes.ReplaceAll(z, []byte("\t"), []byte(" "))
		parts := bytes.Fields(z)
		if len(parts) < 3 {
			continue
		}
		snat := Snat{
			Action:      string(parts[0]),
			Source:      string(parts[1]),
			Destination: string(parts[2]),
		}
		snats = append(snats, snat)
	}
	return
}

func GetSnats() ([]Snat, error) {
	buff, err := os.ReadFile(snatFile)
	if err != nil {
		return nil, err
	}

	return parseSnats(buff), nil
}

func AddSnat(snat Snat) error {
	snats, err := GetSnats()
	if err != nil {
		return err
	}

	for _, s := range snats {
		if s.Action == snat.Action && s.Source == snat.Source && s.Destination == snat.Destination {
			return ErrSnatAlreadyExists
		}
	}

	f, err := os.OpenFile(snatFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s\t%s\t%s\n", snat.Action, snat.Source, snat.Destination)

	if _, err := f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func RemoveSnat(snat Snat) error {
	snats, err := GetSnats()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(snats, func(s Snat) bool {
		return s.Action == snat.Action && s.Source == snat.Source && s.Destination == snat.Destination
	})
	if index == -1 {
		return ErrSnatNotFound
	}

	snats = append(snats[:index], snats[index+1:]...)

	var buff bytes.Buffer
	for _, s := range snats {
		line := fmt.Sprintf("%s\t%s\t%s\n", s.Action, s.Source, s.Destination)
		buff.WriteString(line)
	}
	if err := os.WriteFile(snatFile, buff.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

