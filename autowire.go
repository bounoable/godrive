package godrive

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// AutoWireConfig contains the configuration for the disk autowire.
type AutoWireConfig struct {
	Disks    map[string]DiskCreatorConfig
	Creators map[string]DiskCreator
}

// DiskCreatorConfig is the configuration for the creation of a single storage disk.
type DiskCreatorConfig struct {
	Provider string
	Config   map[string]interface{}
}

// DiskCreator creates storage disks.
type DiskCreator interface {
	CreateDisk(ctx context.Context, cfg map[string]interface{}) (Disk, error)
}

// DiskCreatorFunc creates storage disks.
type DiskCreatorFunc func(context.Context, map[string]interface{}) (Disk, error)

// CreateDisk creates a storage disk.
func (fn DiskCreatorFunc) CreateDisk(ctx context.Context, cfg map[string]interface{}) (Disk, error) {
	return fn(ctx, cfg)
}

// NewAutoWire returns a new autowire configuration.
func NewAutoWire() AutoWireConfig {
	return AutoWireConfig{
		Disks:    make(map[string]DiskCreatorConfig),
		Creators: make(map[string]DiskCreator),
	}
}

// RegisterProvider registers a storage disk creator.
func (cfg AutoWireConfig) RegisterProvider(name string, creator DiskCreator) {
	cfg.Creators[name] = creator
}

// Configure adds a disk to the configuration.
func (cfg AutoWireConfig) Configure(diskname, provider string, config map[string]interface{}) {
	if config == nil {
		config = make(map[string]interface{})
	}

	cfg.Disks[diskname] = DiskCreatorConfig{
		Provider: provider,
		Config:   config,
	}
}

// NewManager creates a new Manager with the initialized storage disks.
func (cfg AutoWireConfig) NewManager(ctx context.Context) (*Manager, error) {
	m := New()

	for diskname, diskcfg := range cfg.Disks {
		creator, ok := cfg.Creators[diskname]
		if !ok {
			return nil, UnregisteredProviderError{Provider: diskcfg.Provider}
		}

		disk, err := creator.CreateDisk(ctx, diskcfg.Config)
		if err != nil {
			return nil, err
		}

		if err := m.Configure(diskname, disk, Replace()); err != nil {
			return nil, err
		}
	}

	return m, nil
}

// UnregisteredProviderError means the configuration contains a disk with an unregistered provider.
type UnregisteredProviderError struct {
	Provider string
}

func (err UnregisteredProviderError) Error() string {
	return fmt.Sprintf("unregistered storage provider '%s'", err.Provider)
}

// Load loads the disk configuration from a file.
// It checks against provided file extensions and
// returns an error if the filetype is unsupported.
func (cfg AutoWireConfig) Load(path string) error {
	ext := filepath.Ext(path)

	switch ext {
	case ".yml":
		fallthrough
	case ".yaml":
		return cfg.LoadYAML(path)
	default:
		return fmt.Errorf("unknown file extension for disk configuration '%s'", ext)
	}
}

// LoadYAML loads the disk configuration from a YAML file.
func (cfg AutoWireConfig) LoadYAML(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return cfg.LoadYAMLReader(f)
}

// LoadYAMLReader loads the disk configuration from the YAML in r.
func (cfg AutoWireConfig) LoadYAMLReader(r io.Reader) error {
	var yamlcfg autowireYamlConfig
	if err := yaml.NewDecoder(r).Decode(&yamlcfg); err != nil {
		return err
	}

	if err := yamlcfg.apply(cfg); err != nil {
		return err
	}

	return nil
}

// map[DISKNAME]map[CONFIGKEY]interface{}
type autowireYamlConfig map[string]map[string]interface{}

func (cfg autowireYamlConfig) apply(config AutoWireConfig) error {
	disks := make(map[string]DiskCreatorConfig)

	for diskname, diskcfg := range cfg {
		if _, ok := disks[diskname]; ok {
			return DuplicateDiskConfigError{DiskName: diskname}
		}

		provider, ok := diskcfg["provider"].(string)
		if !ok {
			return InvalidConfigValueError{
				DiskName:  diskname,
				ConfigKey: "provider",
				Expected:  "",
				Provided:  provider,
			}
		}

		varcfg := make(map[string]interface{})

		if ivarcfg, ok := diskcfg["config"]; ok {
			tcfg, ok := ivarcfg.(map[string]interface{})
			if !ok {
				return InvalidConfigValueError{
					DiskName:  diskname,
					ConfigKey: "config",
					Expected:  new(map[string]interface{}),
					Provided:  tcfg,
				}
			}
			varcfg = tcfg
		}

		disks[diskname] = DiskCreatorConfig{
			Provider: provider,
			Config:   varcfg,
		}
	}

	for diskname, creatorcfg := range disks {
		config.Configure(diskname, creatorcfg.Provider, creatorcfg.Config)
	}

	return nil
}

// DuplicateDiskConfigError means the YAML configuration contains multiple configurations for a disk name.
type DuplicateDiskConfigError struct {
	DiskName string
}

func (err DuplicateDiskConfigError) Error() string {
	return fmt.Sprintf("duplicate configuration for disk '%s'", err.DiskName)
}

// InvalidConfigValueError means a configuration value for a disk has a wrong type.
type InvalidConfigValueError struct {
	DiskName  string
	ConfigKey string
	Expected  interface{}
	Provided  interface{}
}

func (err InvalidConfigValueError) Error() string {
	return fmt.Sprintf("invalid config value for disk '%s': '%s' must be a '%T' but is a '%T'", err.DiskName, err.ConfigKey, err.Expected, err.Provided)
}
