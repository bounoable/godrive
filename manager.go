package godrive

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var (
	// ErrNoDefaultDisk is returned on Disk operations when no default Disk is set.
	ErrNoDefaultDisk = errors.New("no default disk specified")
)

// Manager is a container for multiple Disks and is itself also a Disk.
// Disk operations are delegated to the configured default Disk.
// Manager is thread-safe (but the Disk implementations may not).
type Manager struct {
	mux         sync.RWMutex
	disks       map[string]Disk
	defaultDisk string
}

// New returns a new disk manager. The disk manager is a container for multiple storage disks
// and also implements the Disk interface, so it can be directly used to access the configured default disk.
//
// Normally you don't instantiate the manager with New() but through the AutoWire config.
func New() *Manager {
	return &Manager{
		disks: make(map[string]Disk),
	}
}

// ConfigureOption is a Disk configuration option.
type ConfigureOption func(*configureConfig)

type configureConfig struct {
	replace   bool
	asDefault bool
}

// Replace will replace the previously configured Disk with the same name.
func Replace() ConfigureOption {
	return func(cfg *configureConfig) {
		cfg.replace = true
	}
}

// Default makes the Disk the default Disk.
func Default() ConfigureOption {
	return func(cfg *configureConfig) {
		cfg.asDefault = true
	}
}

// Configure adds a Disk to the Manager.
// If the name is already in use, it returns a DuplicateNameError unless the Replace option is used.
// The first Disk will automatically be made the default Disk, even if the Default option is not used.
func (m *Manager) Configure(name string, disk Disk, options ...ConfigureOption) error {
	var cfg configureConfig
	for _, opt := range options {
		opt(&cfg)
	}

	m.mux.Lock()
	defer m.mux.Unlock()

	defer func() {
		if cfg.asDefault || len(m.disks) == 1 {
			m.defaultDisk = name
		}
	}()

	if _, ok := m.disks[name]; ok {
		if !cfg.replace {
			return DuplicateNameError{Name: name}
		}
	}

	m.disks[name] = disk

	return nil
}

// DuplicateNameError is returned when a Disk is added to a Manager with a name that was already used.
type DuplicateNameError struct {
	Name string
}

func (err DuplicateNameError) Error() string {
	return fmt.Sprintf("duplicate disk name: %s", err.Name)
}

// RemoveDisk removes the Disk with the configured name from the Manager.
func (m *Manager) RemoveDisk(name string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.disks, name)
}

// Disk returns the Disk with the configured name.
// If no Disk with the name is configured, it returns an UnconfiguredDiskError.
func (m *Manager) Disk(name string) (Disk, error) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	disk, ok := m.disks[name]
	if !ok {
		return nil, UnconfiguredDiskError{Name: name}
	}

	return disk, nil
}

// UnconfiguredDiskError is returned when no Disk can be found for a name.
type UnconfiguredDiskError struct {
	Name string
}

func (err UnconfiguredDiskError) Error() string {
	return fmt.Sprintf("unconfigured disk: %s", err.Name)
}

// Put writes b to the file at the given path on the default Disk.
// If no default Disk is set, it returns ErrNoDefaultDisk.
func (m *Manager) Put(ctx context.Context, path string, b []byte) error {
	disk, err := m.Disk(m.defaultDisk)
	if err != nil {
		if errors.As(err, &UnconfiguredDiskError{}) {
			return ErrNoDefaultDisk
		}

		return err
	}

	return disk.Put(ctx, path, b)
}

// Get retrieves the file at the given path.
// If no default Disk is set, it returns ErrNoDefaultDisk.
func (m *Manager) Get(ctx context.Context, path string) ([]byte, error) {
	disk, err := m.Disk(m.defaultDisk)
	if err != nil {
		if errors.As(err, &UnconfiguredDiskError{}) {
			return nil, ErrNoDefaultDisk
		}

		return nil, err
	}

	return disk.Get(ctx, path)
}

// Delete deletes the file at the given path.
// If no default Disk is set, it returns ErrNoDefaultDisk.
func (m *Manager) Delete(ctx context.Context, path string) error {
	disk, err := m.Disk(m.defaultDisk)
	if err != nil {
		if errors.As(err, &UnconfiguredDiskError{}) {
			return ErrNoDefaultDisk
		}

		return err
	}

	return disk.Delete(ctx, path)
}

// GetURL returns the public URL for the file at the given path.
// If no default Disk is set, it returns ErrNoDefaultDisk.
// If the default Disk does not implement URLProvider, it returns an UnimplementedError.
func (m *Manager) GetURL(ctx context.Context, path string) (string, error) {
	disk, err := m.Disk(m.defaultDisk)
	if err != nil {
		if errors.As(err, &UnconfiguredDiskError{}) {
			return "", ErrNoDefaultDisk
		}

		return "", err
	}

	urldisk, ok := disk.(URLProvider)
	if !ok {
		return "", UnimplementedError{
			DiskName:  m.defaultDisk,
			Interface: new(URLProvider),
		}
	}

	return urldisk.GetURL(ctx, path)
}

// UnimplementedError means a Disk does not implement a specific feature.
type UnimplementedError struct {
	DiskName  string
	Interface interface{}
}

func (err UnimplementedError) Error() string {
	return fmt.Sprintf("disk '%s' does not implement '%T'", err.DiskName, err.Interface)
}
