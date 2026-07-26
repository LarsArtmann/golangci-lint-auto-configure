package types

// SettingsMap wraps map[string]any for linter/formatter settings access.
// Centralizes type assertions that were previously scattered across clone,
// merge, validate, and prune code paths.
type SettingsMap map[string]any

// AsSettingsMap converts a raw any value to SettingsMap.
// Returns false if the value is not a map[string]any.
func AsSettingsMap(v any) (SettingsMap, bool) {
	m, ok := v.(map[string]any)
	if !ok {
		return nil, false
	}

	return SettingsMap(m), true
}

// IsEmpty reports whether the settings map has no entries.
func (s SettingsMap) IsEmpty() bool {
	return len(s) == 0
}

// GetMap returns a nested SettingsMap at the given key.
func (s SettingsMap) GetMap(key string) (SettingsMap, bool) {
	v, ok := s[key]
	if !ok {
		return nil, false
	}

	return AsSettingsMap(v)
}

// Clone returns a deep copy of the settings map, recursively cloning
// nested maps and slices.
func (s SettingsMap) Clone() SettingsMap {
	if s == nil {
		return nil
	}

	result := make(SettingsMap, len(s))
	for k, v := range s {
		result[k] = deepCloneAny(v)
	}

	return result
}
