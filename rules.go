package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Rule struct {
	Action      string
	Source      string
	Destination string
	Protocol    string
	Dport       string
	Sport       string
	Origdest    string
}

var (
	ErrRuleAlreadyExists = errors.New("rule already exists")
	ErrRuleNotFound      = errors.New("rule not found")
)

func parseRules(data []byte) (rules []Rule) {
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
		rule := Rule{
			Action:      string(parts[0]),
			Source:      string(parts[1]),
			Destination: string(parts[2]),
		}
		if len(parts) > 3 {
			rule.Protocol = string(parts[3])
		}
		if len(parts) > 4 {
			rule.Dport = string(parts[4])
		}
		if len(parts) > 5 {
			rule.Sport = string(parts[5])
		}
		if len(parts) > 6 {
			rule.Origdest = string(parts[6])
		}
		rules = append(rules, rule)
	}
	return
}

func GetRules() ([]Rule, error) {
	buff, err := os.ReadFile(rulesFile)
	if err != nil {
		return nil, err
	}

	return parseRules(buff), nil
}

func AddRule(rule Rule) error {
	rules, err := GetRules()
	if err != nil {
		return err
	}

	rule = rule.fillEmpty()

	for _, r := range rules {
		if r.Equals(rule) {
			return ErrRuleAlreadyExists
		}
	}

	f, err := os.OpenFile(rulesFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", rule.Action, rule.Source, rule.Destination, rule.Protocol, rule.Dport, rule.Sport, rule.Origdest)

	if _, err := f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func RemoveRule(rule Rule) error {
	rules, err := GetRules()
	if err != nil {
		return err
	}

	rule = rule.fillEmpty()

	index := slices.IndexFunc(rules, func(r Rule) bool {
		return r.Equals(rule)
	})
	if index == -1 {
		return ErrRuleNotFound
	}

	rules = slices.Delete(rules, index, index+1)

	var buff bytes.Buffer
	for _, r := range rules {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", r.Action, r.Source, r.Destination, r.Protocol, r.Dport, r.Sport, r.Origdest)
		buff.WriteString(line)
	}
	if err := os.WriteFile(rulesFile, buff.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}

func (r Rule) Compare(other Rule) int {
	r = r.fillEmpty()
	other = other.fillEmpty()

	if r.Action != other.Action {
		return strings.Compare(r.Action, other.Action)
	}
	if r.Source != other.Source {
		return strings.Compare(r.Source, other.Source)
	}
	if r.Destination != other.Destination {
		return strings.Compare(r.Destination, other.Destination)
	}
	if r.Protocol != other.Protocol {
		return strings.Compare(r.Protocol, other.Protocol)
	}
	if r.Dport != other.Dport {
		return strings.Compare(r.Dport, other.Dport)
	}
	if r.Sport != other.Sport {
		return strings.Compare(r.Sport, other.Sport)
	}
	if r.Origdest != other.Origdest {
		return strings.Compare(r.Origdest, other.Origdest)
	}
	return 0
}

func (r Rule) Equals(other Rule) bool {
	return r.Compare(other) == 0
}

// fillEmpty fills empty fields with "-" where necessary
func (r Rule) fillEmpty() Rule {
	if r.Dport == "" && r.Sport != "" {
		r.Dport = "-"
	}

	if r.Sport == "" && r.Origdest != "" {
		r.Sport = "-"
	}

	return r
}
