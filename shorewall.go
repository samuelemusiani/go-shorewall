package goshorewall

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

const (
	shorewallConfigPath = "/etc/shorewall"
	zonesFile           = shorewallConfigPath + "/zones"
	interfacesFile      = shorewallConfigPath + "/interfaces"
	policyFile          = shorewallConfigPath + "/policy"
	rulesFile           = shorewallConfigPath + "/rules"
	hostsFile           = shorewallConfigPath + "/hosts"
	snatFile            = shorewallConfigPath + "/snat"
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

func GetVersion() (string, error) {
	stdout, stderr, err := executeCommand("/usr/sbin/shorewall", "version")
	if err != nil {
		err = errors.Join(fmt.Errorf("failed to execute shorewall version command: %w", err), errors.New(stderr))
		return "", err
	}

	return stdout, nil
}
