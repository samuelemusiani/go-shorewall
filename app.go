package goshorewall

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/gofrs/flock"
	"github.com/google/uuid"
)

// var lockDirPath = "/etc/shorewall/.goshorewall"
var lockDirPath = "/tmp/tsh/.goshorewall"

// App represents a single application that can manage Shorewall configurations.
// This solves the problem of having multiple independent application managing Shorewall
// in the same system. App use a random generated identifier to isolate its configuration
// from other applications. This identifier must be saved by the application to be able
// to manage its Shorewall configuration across restarts. App also handles concurrent
// access to Shorewall configuration files.
type App struct {
	basePath   string
	identifier uuid.UUID
}

// NewApp creates a new App with a random generated identifier.
// This must be called only once per application installation.
// After saving the identifier with App.ID(), the application should use
// AppFromID to retrieve the App instance on subsequent runs.
func NewApp() (*App, error) {
	return NewAppWithBasePath(shorewallConfigPath)
}

func NewAppWithBasePath(basePath string) (*App, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return nil, fmt.Errorf("failed to generate application identifier: %w", err)
	}
	return &App{
		basePath:   basePath,
		identifier: id,
	}, nil
}

// AppFromID creates an App instance from a previously saved identifier.
func AppFromID(id string) (*App, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("failed to parse application identifier: %w", err)
	}
	return &App{
		identifier: parsedID,
	}, nil
}

// ID returns the unique identifier of the App instance.
func (a *App) ID() string {
	return a.identifier.String()
}

// BasePath returns the Shorewall configuration base path used by the App instance.
func (a *App) BasePath() string {
	return a.basePath
}

// ZonesFilePath returns the full path to the zones file used by the App instance.
func (a *App) ZonesFilePath() string {
	return path.Join(a.basePath, zonesFile)
}

// InterfaceFilePath returns the full path to the interfaces file used by the App instance.
func (a *App) InterfaceFilePath() string {
	return path.Join(a.basePath, interfacesFile)
}

// PolicyFilePath returns the full path to the policy file used by the App instance.
func (a *App) PolicyFilePath() string {
	return path.Join(a.basePath, policyFile)
}

// RulesFilePath returns the full path to the rules file used by the App instance.
func (a *App) RulesFilePath() string {
	return path.Join(a.basePath, rulesFile)
}

// SnatFilePath returns the full path to the snat file used by the App instance.
func (a *App) SnatFilePath() string {
	return path.Join(a.basePath, snatFile)
}

// Reload reloads Shorewall configuration.
func (a *App) Reload() error {
	return execWithLock("reload", Reload)
}

// Version returns the Shorewall version.
func (a *App) Version() (string, error) {
	return Version()
}

// Interfaces returns the list of interfaces managed by the App instance.
func (a *App) Interfaces() ([]Interface, error) {
	return execGetWithLock("interfaces", a.ID(), interfacesFile, getInterfacesBuff)
}

// AddInterface adds a new interface to the Shorewall configuration managed by the App instance.
func (a *App) AddInterface(iface Interface) error {
	return execAddRemoveWithLock("interfaces", a.ID(), interfacesFile, addInterfaceBuff, iface)
}

// RemoveInterfaceByZone removes all interfaces associated with the specified zone
func (a *App) RemoveInterfaceByZone(zone string) error {
	return execAddRemoveWithLock("interfaces", a.ID(), interfacesFile, removeInterfaceByZoneBuff, zone)
}

// Policies returns the list of policies managed by the App instance.
func (a *App) Policies() ([]Policy, error) {
	return execGetWithLock("policies", a.ID(), policyFile, getPoliciesBuff)
}

// AddPolicy adds a new policy to the Shorewall configuration managed by the App instance.
func (a *App) AddPolicy(policy Policy) error {
	return execAddRemoveWithLock("policies", a.ID(), policyFile, addPolicyBuff, policy)
}

// RemovePolicy removes a policy from the Shorewall configuration managed by the App instance.
func (a *App) RemovePolicy(policy Policy) error {
	return execAddRemoveWithLock("policies", a.ID(), policyFile, removePolicyBuff, policy)
}

// Rules returns the list of rules managed by the App instance.
func (a *App) Rules() ([]Rule, error) {
	return execGetWithLock("rules", a.ID(), rulesFile, getRulesBuff)
}

// AddRule adds a new rule to the Shorewall configuration managed by the App instance.
func (a *App) AddRule(rule Rule) error {
	return execAddRemoveWithLock("rules", a.ID(), rulesFile, addRuleBuff, rule)
}

// RemoveRule removes a rule from the Shorewall configuration managed by the App instance.
func (a *App) RemoveRule(rule Rule) error {
	return execAddRemoveWithLock("rules", a.ID(), rulesFile, removeRuleBuff, rule)
}

// Snats returns the list of SNATs managed by the App instance.
func (a *App) Snats() ([]Snat, error) {
	return execGetWithLock("snats", a.ID(), snatFile, getSnatsBuff)
}

// AddSnat adds a new SNAT to the Shorewall configuration managed by the App instance.
func (a *App) AddSnat(snat Snat) error {
	return execAddRemoveWithLock("snats", a.ID(), snatFile, addSnatBuff, snat)
}

// RemoveSnat removes a SNAT from the Shorewall configuration managed by the App instance.
func (a *App) RemoveSnat(snat Snat) error {
	return execAddRemoveWithLock("snats", a.ID(), snatFile, removeSnatBuff, snat)
}

// Zones returns the list of zones managed by the App instance.
func (a *App) Zones() ([]Zone, error) {
	return execGetWithLock("zones", a.ID(), zonesFile, getZonesBuff)
}

// AddZone adds a new zone to the Shorewall configuration managed by the App instance.
func (a *App) AddZone(zone Zone) error {
	return execAddRemoveWithLock("zones", a.ID(), zonesFile, addZoneBuff, zone)
}

// RemoveZone removes a zone from the Shorewall configuration managed by the App instance.
func (a *App) RemoveZone(zoneName string) error {
	return execAddRemoveWithLock("zones", a.ID(), zonesFile, removeZoneBuff, zoneName)
}

func execWithLock(component string, fn func() error) error {
	flock, err := takeLock(component)
	if err != nil {
		return fmt.Errorf("failed to take lock for component %s: %w", component, err)
	}
	defer flock.Unlock()
	err = flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock for component %s: %w", component, err)
	}
	return fn()
}

func execGetWithLock[S any](component, id, path string, fn func([]byte) ([]S, error)) ([]S, error) {
	flock, err := takeLock(component)
	if err != nil {
		return nil, fmt.Errorf("failed to take lock for component %s: %w", component, err)
	}
	defer flock.Unlock()
	err = flock.Lock()
	if err != nil {
		return nil, fmt.Errorf("failed to acquire lock for component %s: %w", component, err)
	}
	return appReadFile(id, path, fn)
}

func execAddRemoveWithLock[S any](component, id, path string, fn func([]byte, S) ([]byte, error), item S) error {
	flock, err := takeLock(component)
	if err != nil {
		return fmt.Errorf("failed to take lock for component %s: %w", component, err)
	}
	defer flock.Unlock()
	err = flock.Lock()
	if err != nil {
		return fmt.Errorf("failed to acquire lock for component %s: %w", component, err)
	}
	return appReadWriteFile(id, path, fn, item)
}

func takeLock(component string) (*flock.Flock, error) {
	err := os.Mkdir(lockDirPath, 0o755)
	if err != nil && !os.IsExist(err) {
		return nil, fmt.Errorf("failed to create lock directory: %w", err)
	}
	lockFilePath := path.Join(lockDirPath, component+".lock")
	return flock.New(lockFilePath), nil
}

func commentIdentifierLineStart(id string) []byte {
	return fmt.Appendf([]byte{}, "### Managed by goshorewall app ID: %s ###\n", id)
}

func commentIdentifierEnd(id string) []byte {
	return fmt.Appendf([]byte{}, "### End of goshorewall app ID: %s ###\n", id)
}

func extractApplicationSubsetBufferIndexes(id string, buff []byte) (uint, uint, bool, error) {
	startMarker := bytes.TrimRight(commentIdentifierLineStart(id), "\n")
	endMarker := bytes.TrimRight(commentIdentifierEnd(id), "\n")

	var startIdx, endIdx int = -1, -1
	var counter = 0
	for line := range bytes.SplitSeq(buff, []byte("\n")) {
		counter += len(line) + 1
		line = bytes.TrimRight(line, "\n")
		if bytes.Contains(line, startMarker) {
			startIdx = counter
		} else if bytes.Contains(line, endMarker) {
			if startIdx == -1 {
				return 0, 0, false, fmt.Errorf("malformed configuration: end marker found without start marker")
			}
			endIdx = counter - len(line) - 1
			break
		}
	}

	// First time using this app on this file
	if startIdx == -1 && endIdx == -1 {
		l := uint(len(buff))
		return l, l, false, nil
	}

	if startIdx == -1 {
		return 0, 0, false, fmt.Errorf("malformed configuration: start marker not found")
	} else if endIdx == -1 {
		return 0, 0, false, fmt.Errorf("malformed configuration: end marker not found")
	}

	return uint(startIdx), uint(endIdx), true, nil
}

func wrapBuffWithAppIdentifier(buff []byte, id string) []byte {
	startMarker := commentIdentifierLineStart(id)
	endMarker := commentIdentifierEnd(id)

	b := make([]byte, 0, len(startMarker)+len(buff)+len(endMarker))

	b = append(b, startMarker...)
	b = append(b, buff...)
	b = append(b, endMarker...)

	return b
}

func appReadFile[S any](id, path string, fn func([]byte) ([]S, error)) ([]S, error) {
	buff, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	is, ie, _, err := extractApplicationSubsetBufferIndexes(id, buff)
	if err != nil {
		return nil, err
	}

	return fn(buff[is:ie])
}

func appReadWriteFile[S any](id, path string, fn func([]byte, S) ([]byte, error), i S) error {
	file, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	defer file.Close()

	buff, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	is, ie, found, err := extractApplicationSubsetBufferIndexes(id, buff)
	if err != nil {
		return err
	}

	tmpBuff := make([]byte, ie-is)
	copy(tmpBuff, buff[is:ie])

	buff2, err := fn(tmpBuff, i)
	if err != nil {
		return err
	}

	if ie == is && !found {
		buff2 = wrapBuffWithAppIdentifier(buff2, id)
	}

	newBuff := make([]byte, 0, len(buff2)+len(buff[ie:]))
	newBuff = append(newBuff, buff2...)
	newBuff = append(newBuff, buff[ie:]...)

	n, err := file.WriteAt(newBuff, int64(is))
	if err != nil {
		return err
	}
	if n < len(buff2) {
		return fmt.Errorf("failed to write complete data to zones file")
	}
	return file.Truncate(int64(int(is) + n))
}
