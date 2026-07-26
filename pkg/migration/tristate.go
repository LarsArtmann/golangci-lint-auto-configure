package migration

import "go.yaml.in/yaml/v3"

// TriState represents a three-valued boolean: unspecified, enabled, or disabled.
// It replaces *bool fields in v1 migration config types where nil/true/false
// are distinct states (nil = not specified, true = enabled, false = disabled).
type TriState int

const (
	TriStateUnspecified TriState = iota
	TriStateEnabled
	TriStateDisabled
)

// UnmarshalYAML decodes a YAML boolean into a TriState.
// Absent keys leave the zero value (TriStateUnspecified).
//
//nolint:wrapcheck // yaml.Node.Decode is a trusted library call
func (t *TriState) UnmarshalYAML(value *yaml.Node) error {
	var b bool

	err := value.Decode(&b)
	if err != nil {
		return err
	}

	if b {
		*t = TriStateEnabled
	} else {
		*t = TriStateDisabled
	}

	return nil
}
