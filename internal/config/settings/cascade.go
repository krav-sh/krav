package settings

import (
	"errors"

	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/v2"

	goyaml "gopkg.in/yaml.v3"
)

type Source string

const (
	SourceDefaults           Source = "defaults"
	SourceManagedRecommended Source = "managed-recommended"
	SourceSystem             Source = "system"
	SourceUser               Source = "user"
	SourceProject            Source = "project"
	SourceProjectLocal       Source = "project-local"
	SourceEnvironment        Source = "environment"
	SourceCLI                Source = "cli"
	SourceCustom             Source = "custom"
	SourceManagedRequired    Source = "managed-required"
)

// SourceOrder defines the cascade order from lowest to highest precedence.
var SourceOrder = []Source{
	SourceDefaults,
	SourceManagedRecommended,
	SourceSystem,
	SourceUser,
	SourceProject,
	SourceProjectLocal,
	SourceEnvironment,
	SourceCLI,
	SourceCustom,
	SourceManagedRequired,
}

// Precedence returns the precedence level of this source in the cascade.
// Higher values indicate higher precedence (override lower values).
// Returns -1 if the source is not in the standard cascade order.
func (s Source) Precedence() int {
	for i, src := range SourceOrder {
		if src == s {
			return i
		}
	}
	return -1
}

type Cascade interface {
	All() map[string]any
	AllFromSource(Source) map[string]any
	AllWithSources() map[string]Entry
	Copy() Cascade
	Exists(string) bool
	ExistsInSource(string, Source) bool
	Explain(string) (Entry, bool)
	Get(string) any
	GetFromSource(string, Source) any
	GetOrDefault(string, any) any
	GetOrDefaultFromSource(string, any, Source) any
	KeyMap() map[string][]string
	Keys() []string
	MarshalYAML() ([]byte, error)
	Merge(Cascade) error
	Unmarshal(path string, o any) error
}

// Layer represents a single configuration source within a cascade level.
type Layer struct {
	cfg  *koanf.Koanf
	Path string // file path, empty for non-file sources
	Desc string // human-readable description for non-file sources
}

// Entry represents a configuration value with its provenance information.
type Entry struct {
	Key    string
	Value  any
	Source Source
	Path   string // file path, empty for env/cli/defaults
}

// DefaultCascade implements Cascade with layered configuration sources.
type DefaultCascade struct {
	layers map[Source][]Layer
	merged *koanf.Koanf
}

// NewDefaultCascade creates a new DefaultCascade with properly initialized maps.
func NewDefaultCascade() *DefaultCascade {
	return &DefaultCascade{
		layers: make(map[Source][]Layer),
		merged: koanf.NewWithConf(koanf.Conf{
			Delim:       ".",
			StrictMerge: false,
		}),
	}
}

// All returns all configuration values from the merged cascade.
func (cfg *DefaultCascade) All() map[string]any {
	return cfg.merged.All()
}

// AllFromSource returns all configuration values from a specific source.
func (cfg *DefaultCascade) AllFromSource(source Source) map[string]any {
	result := make(map[string]any)
	for _, layer := range cfg.layers[source] {
		for k, v := range layer.cfg.All() {
			result[k] = v
		}
	}
	return result
}

// AllWithSources returns all configuration values with their provenance information.
func (cfg *DefaultCascade) AllWithSources() map[string]Entry {
	result := make(map[string]Entry)

	for _, source := range SourceOrder {
		for _, layer := range cfg.layers[source] {
			for _, key := range layer.cfg.Keys() {
				result[key] = Entry{
					Key:    key,
					Value:  layer.cfg.Get(key),
					Source: source,
					Path:   layer.Path,
				}
			}
		}
	}

	return result
}

// Copy creates an independent copy of the cascade that doesn't share state.
func (cfg *DefaultCascade) Copy() Cascade {
	newCfg := NewDefaultCascade()

	for source, layers := range cfg.layers {
		for _, layer := range layers {
			k := koanf.New(".")
			_ = k.Load(confmap.Provider(layer.cfg.All(), "."), nil)
			newCfg.setLayer(k, source, layer.Path, layer.Desc)
		}
	}

	return newCfg
}

// Exists returns true if the key exists in the merged cascade.
func (cfg *DefaultCascade) Exists(key string) bool {
	return cfg.merged.Exists(key)
}

// ExistsInSource returns true if the key exists in a specific source.
func (cfg *DefaultCascade) ExistsInSource(key string, source Source) bool {
	for _, layer := range cfg.layers[source] {
		if layer.cfg.Exists(key) {
			return true
		}
	}
	return false
}

// Explain returns provenance information for a key, showing which source provided it.
func (cfg *DefaultCascade) Explain(key string) (Entry, bool) {
	for i := len(SourceOrder) - 1; i >= 0; i-- {
		source := SourceOrder[i]
		layers := cfg.layers[source]
		for j := len(layers) - 1; j >= 0; j-- {
			layer := layers[j]
			if layer.cfg.Exists(key) {
				return Entry{
					Key:    key,
					Value:  layer.cfg.Get(key),
					Source: source,
					Path:   layer.Path,
				}, true
			}
		}
	}
	return Entry{}, false
}

// Get returns the value for a key from the merged cascade.
func (cfg *DefaultCascade) Get(key string) any {
	return cfg.merged.Get(key)
}

// GetFromSource returns the value for a key from a specific source.
func (cfg *DefaultCascade) GetFromSource(key string, source Source) any {
	layers := cfg.layers[source]
	for i := len(layers) - 1; i >= 0; i-- {
		if layers[i].cfg.Exists(key) {
			return layers[i].cfg.Get(key)
		}
	}
	return nil
}

// GetOrDefault returns the value for a key, or the default if the key doesn't exist.
func (cfg *DefaultCascade) GetOrDefault(key string, defaultVal any) any {
	if cfg.merged.Exists(key) {
		return cfg.merged.Get(key)
	}
	return defaultVal
}

// GetOrDefaultFromSource returns the value for a key from a specific source,
// or the default if the key doesn't exist in that source.
func (cfg *DefaultCascade) GetOrDefaultFromSource(key string, defaultVal any, source Source) any {
	if val := cfg.GetFromSource(key, source); val != nil {
		return val
	}
	return defaultVal
}

// KeyMap returns a map of all keys to their path components.
func (cfg *DefaultCascade) KeyMap() map[string][]string {
	return cfg.merged.KeyMap()
}

// Keys returns all keys in the merged cascade.
func (cfg *DefaultCascade) Keys() []string {
	return cfg.merged.Keys()
}

// MarshalYAML marshals the merged cascade to YAML.
func (cfg *DefaultCascade) MarshalYAML() ([]byte, error) {
	return goyaml.Marshal(cfg.merged.All())
}

// Merge merges another cascade into this one, preserving source information.
func (cfg *DefaultCascade) Merge(other Cascade) error {
	otherCfg, ok := other.(*DefaultCascade)
	if !ok {
		return errors.New("can only merge DefaultCascade instances")
	}

	for source, layers := range otherCfg.layers {
		for _, layer := range layers {
			k := koanf.New(".")
			_ = k.Load(confmap.Provider(layer.cfg.All(), "."), nil)
			cfg.setLayer(k, source, layer.Path, layer.Desc)
		}
	}

	return nil
}

// Unmarshal unmarshals a path in the cascade to a struct.
func (cfg *DefaultCascade) Unmarshal(path string, o any) error {
	return cfg.merged.UnmarshalWithConf(path, o, koanf.UnmarshalConf{
		Tag: "config",
	})
}

// setLayer adds a layer to the cascade for a specific source.
func (cfg *DefaultCascade) setLayer(k *koanf.Koanf, source Source, path, desc string) {
	if cfg.layers[source] == nil {
		cfg.layers[source] = []Layer{}
	}

	cfg.layers[source] = append(cfg.layers[source], Layer{
		cfg:  k,
		Path: path,
		Desc: desc,
	})

	_ = cfg.merged.Merge(k)
}

var _ Cascade = (*DefaultCascade)(nil)
