package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
)

type Policy struct {
	Source      string
	Destination string
	Policy      string
	Log         string
}

var (
	ErrPolicyAlreadyExists = errors.New("policy already exists")
	ErrPolicyNotFound      = errors.New("policy not found")
)

func parsePolicies(data []byte) (policies []Policy) {
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
		policy := Policy{
			Source:      string(parts[0]),
			Destination: string(parts[1]),
			Policy:      string(parts[2]),
		}
		if len(parts) > 3 {
			policy.Log = string(parts[3])
		}
		policies = append(policies, policy)
	}
	return
}

func GetPolicies() ([]Policy, error) {
	buff, err := os.ReadFile(policyFile)
	if err != nil {
		return nil, err
	}

	return parsePolicies(buff), nil
}

func AddPolicy(policy Policy) error {
	policies, err := GetPolicies()
	if err != nil {
		return err
	}

	slices.SortFunc(policies, func(a, b Policy) int {
		if cmp := strings.Compare(a.Source, b.Source); cmp != 0 {
			return cmp
		}
		if cmp := strings.Compare(a.Destination, b.Destination); cmp != 0 {
			return cmp
		}
		return strings.Compare(a.Policy, b.Policy)
	})

	for _, p := range policies {
		if p.Source == policy.Source && p.Destination == policy.Destination && p.Policy == policy.Policy {
			return ErrPolicyAlreadyExists
		}
	}

	f, err := os.OpenFile(policyFile, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	line := fmt.Sprintf("%s\t%s\t%s\t%s\n", policy.Source, policy.Destination, policy.Policy, policy.Log)

	if _, err := f.WriteString(line); err != nil {
		return err
	}

	return nil
}

func RemovePolicy(policy Policy) error {
	policies, err := GetPolicies()
	if err != nil {
		return err
	}

	index := slices.IndexFunc(policies, func(p Policy) bool {
		return p.Source == policy.Source && p.Destination == policy.Destination && p.Policy == policy.Policy
	})
	if index == -1 {
		return ErrPolicyNotFound
	}

	policies = append(policies[:index], policies[index+1:]...)

	var buff bytes.Buffer
	for _, p := range policies {
		line := fmt.Sprintf("%s\t%s\t%s\t%s\n", p.Source, p.Destination, p.Policy, p.Log)
		buff.WriteString(line)
	}
	if err := os.WriteFile(policyFile, buff.Bytes(), 0644); err != nil {
		return err
	}

	return nil
}
