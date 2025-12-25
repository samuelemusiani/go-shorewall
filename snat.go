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
	ErrSnatAlreadyExists = errors.New("snat rule already exists")
	ErrSnatNotFound      = errors.New("snat rule not found")
)

type Snat struct {
	Action      string
	Source      string
	Destination string
}

func (s Snat) Compare(other Snat) int {
	if cmp := strings.Compare(s.Action, other.Action); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(s.Source, other.Source); cmp != 0 {
		return cmp
	}
	return strings.Compare(s.Destination, other.Destination)
}

func (s Snat) Equals(other Snat) bool {
	return s.Action == other.Action && s.Source == other.Source && s.Destination == other.Destination
}

func (s Snat) Format() string {
	return fmt.Sprintf("%s\t%s\t%s", s.Action, s.Source, s.Destination)
}

func Snats() ([]Snat, error) {
	buff, err := os.ReadFile(fullSnatFile)
	if err != nil {
		return nil, err
	}
	return getSnatsBuff(buff)
}

func AddSnat(snat Snat) error {
	return readWriteFile(fullSnatFile, addSnatBuff, snat)
}

func RemoveSnat(snat Snat) error {
	return readWriteFile(fullSnatFile, removeSnatBuff, snat)
}

func getSnatsBuff(buff []byte) ([]Snat, error) {
	return parseSnats(buff), nil
}

func addSnatBuff(buff []byte, snat Snat) ([]byte, error) {
	snats, err := getSnatsBuff(buff)
	if err != nil {
		return nil, err
	}

	if slices.ContainsFunc(snats, func(s Snat) bool {
		return s.Equals(snat)
	}) {
		return nil, ErrSnatAlreadyExists
	}

	return fmt.Appendf(buff, "%s\n", snat.Format()), nil
}

func removeSnatBuff(buff []byte, snat Snat) ([]byte, error) {
	snats, err := getSnatsBuff(buff)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(snats, func(s Snat) bool {
		return s.Equals(snat)
	})
	if index == -1 {
		return nil, ErrSnatNotFound
	}

	snats = slices.Delete(snats, index, index+1)

	var b bytes.Buffer
	for _, s := range snats {
		line := fmt.Sprintf("%s\n", s.Format())
		b.WriteString(line)
	}
	return b.Bytes(), nil
}

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
