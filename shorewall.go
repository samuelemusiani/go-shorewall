package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path"
)

const (
	shorewallConfigPath = "/etc/shorewall"

	zonesFile      = "zones"
	interfacesFile = "interfaces"
	policyFile     = "policy"
	rulesFile      = "rules"
	snatFile       = "snat"
)

var (
	fullZonesFile      = path.Join(shorewallConfigPath, zonesFile)
	fullInterfacesFile = path.Join(shorewallConfigPath, interfacesFile)
	fullPolicyFile     = path.Join(shorewallConfigPath, policyFile)
	fullRulesFile      = path.Join(shorewallConfigPath, rulesFile)
	fullSnatFile       = path.Join(shorewallConfigPath, snatFile)
)

func executeCommand(command string, args ...string) (string, string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(command, args...)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func Version() (string, error) {
	stdout, stderr, err := executeCommand("/usr/sbin/shorewall", "version")
	if err != nil {
		err = errors.Join(fmt.Errorf("failed to execute shorewall version command: %w", err), errors.New(stderr))
		return "", err
	}

	return stdout, nil
}

func Reload() error {
	_, stderr, err := executeCommand("/usr/sbin/shorewall", "reload")
	if err != nil {
		err = errors.Join(fmt.Errorf("failed to reload Shorewall: %w", err), errors.New(stderr))
		return err
	}
	return nil
}
