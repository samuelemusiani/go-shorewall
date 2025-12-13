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
	ErrRuleAlreadyExists = errors.New("rule already exists")
	ErrRuleNotFound      = errors.New("rule not found")
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

func (r Rule) Compare(other Rule) int {
	r = r.fillEmpty()
	other = other.fillEmpty()
	if cmp := strings.Compare(r.Action, other.Action); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(r.Source, other.Source); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(r.Destination, other.Destination); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(r.Protocol, other.Protocol); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(r.Dport, other.Dport); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(r.Sport, other.Sport); cmp != 0 {
		return cmp
	}
	return strings.Compare(r.Origdest, other.Origdest)
}

func (r Rule) Equals(other Rule) bool {
	return r.Compare(other) == 0
}

func (r Rule) Format() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s", r.Action, r.Source, r.Destination, r.Protocol, r.Dport, r.Sport, r.Origdest)
}

func GetRules() ([]Rule, error) {
	buff, err := os.ReadFile(rulesFile)
	if err != nil {
		return nil, err
	}
	return getRulesBuff(buff)
}

func AddRule(rule Rule) error {
	return readWriteFile(rulesFile, addRuleBuff, rule)
}

func RemoveRule(rule Rule) error {
	return readWriteFile(rulesFile, removeRuleBuff, rule)
}

func getRulesBuff(buff []byte) ([]Rule, error) {
	return parseRules(buff), nil
}

func addRuleBuff(buff []byte, rule Rule) ([]byte, error) {
	rules, err := getRulesBuff(buff)
	if err != nil {
		return nil, err
	}

	rule = rule.fillEmpty()

	if slices.ContainsFunc(rules, func(r Rule) bool {
		return r.Equals(rule)
	}) {
		return nil, ErrRuleAlreadyExists
	}

	return fmt.Appendf(buff, "%s\n", rule.Format()), nil
}

func removeRuleBuff(buff []byte, rule Rule) ([]byte, error) {
	rules, err := getRulesBuff(buff)
	if err != nil {
		return nil, err
	}

	rule = rule.fillEmpty()

	index := slices.IndexFunc(rules, func(r Rule) bool {
		return r.Equals(rule)
	})
	if index == -1 {
		return nil, ErrRuleNotFound
	}

	rules = slices.Delete(rules, index, index+1)

	var b bytes.Buffer
	for _, r := range rules {
		b.WriteString(fmt.Sprintf("%s\n", r.Format()))
	}

	return b.Bytes(), nil
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
