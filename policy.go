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
	ErrPolicyAlreadyExists = errors.New("policy already exists")
	ErrPolicyNotFound      = errors.New("policy not found")
)

type Policy struct {
	Source      string
	Destination string
	Policy      string
	Log         string
}

func (p Policy) Compare(other Policy) int {
	if cmp := strings.Compare(p.Source, other.Source); cmp != 0 {
		return cmp
	}
	if cmp := strings.Compare(p.Destination, other.Destination); cmp != 0 {
		return cmp
	}
	return strings.Compare(p.Policy, other.Policy)
}

func (p Policy) Equals(other Policy) bool {
	return p.Source == other.Source && p.Destination == other.Destination && p.Policy == other.Policy
}

func (p Policy) Format() string {
	return fmt.Sprintf("%s\t%s\t%s\t%s", p.Source, p.Destination, p.Policy, p.Log)
}

func Policies() ([]Policy, error) {
	buff, err := os.ReadFile(fullPolicyFile)
	if err != nil {
		return nil, err
	}
	return getPoliciesBuff(buff)
}

func AddPolicy(policy Policy) error {
	return readWriteFile(fullPolicyFile, addPolicyBuff, policy)
}

func getPoliciesBuff(buff []byte) ([]Policy, error) {
	return parsePolicies(buff), nil
}

func RemovePolicy(policy Policy) error {
	return readWriteFile(fullPolicyFile, removePolicyBuff, policy)
}

func addPolicyBuff(buff []byte, policy Policy) ([]byte, error) {
	policies, err := getPoliciesBuff(buff)
	if err != nil {
		return nil, err
	}

	if slices.ContainsFunc(policies, func(p Policy) bool {
		return p.Equals(policy)
	}) {
		return nil, ErrPolicyAlreadyExists
	}

	return fmt.Appendf(buff, "%s\n", policy.Format()), nil
}

func removePolicyBuff(buff []byte, policy Policy) ([]byte, error) {
	policies, err := getPoliciesBuff(buff)
	if err != nil {
		return nil, err
	}

	index := slices.IndexFunc(policies, func(p Policy) bool {
		return p.Equals(policy)
	})
	if index == -1 {
		return nil, ErrPolicyNotFound
	}

	policies = slices.Delete(policies, index, index+1)

	var b bytes.Buffer
	for _, p := range policies {
		b.WriteString(fmt.Sprintf("%s\n", p.Format()))
	}

	return b.Bytes(), nil
}

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
