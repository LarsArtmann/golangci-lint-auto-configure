package constants

import (
	"testing"
)

func BenchmarkSettingsToMap_Simple(b *testing.B) {
	s := CyclopSettings{MaxComplexity: 12}

	b.ResetTimer()

	for range b.N {
		_ = s.ToMap()
	}
}

func BenchmarkSettingsToMap_WithSlices(b *testing.B) {
	s := IreturnSettings{
		Allow: []string{"error", "empty", "anon", "stdlib", "generic"},
	}

	b.ResetTimer()

	for range b.N {
		_ = s.ToMap()
	}
}

func BenchmarkSettingsToMap_WithNestedMaps(b *testing.B) {
	s := DepguardSettings{
		Rules: map[string]DepguardRule{
			"main": {Allow: []string{"$gostd", "$module"}},
		},
	}

	b.ResetTimer()

	for range b.N {
		_ = s.ToMap()
	}
}

func BenchmarkSettingsToMap_AllDefaults(b *testing.B) {
	b.ResetTimer()

	for range b.N {
		for _, s := range DefaultLinterSettings {
			_ = s.ToMap()
		}
	}
}

func FuzzSettingsToMap_StringField(f *testing.F) {
	f.Add("test")
	f.Add("")
	f.Add("very-long-string-with-special-chars-!@#$%^&*()")

	f.Fuzz(func(t *testing.T, name string) {
		s := ExhaustructSettings{Exclude: []string{name}}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("settingsToMap panicked with input %q: %v", name, r)
			}
		}()

		m := s.ToMap()
		if m == nil {
			t.Errorf("settingsToMap returned nil for input %q", name)
		}
	})
}

func FuzzSettingsToMap_IntField(f *testing.F) {
	f.Add(0)
	f.Add(120)
	f.Add(-1)

	f.Fuzz(func(t *testing.T, n int) {
		s := CyclopSettings{MaxComplexity: n}

		defer func() {
			if r := recover(); r != nil {
				t.Errorf("settingsToMap panicked with input %d: %v", n, r)
			}
		}()

		m := s.ToMap()

		if m == nil {
			t.Errorf("settingsToMap returned nil for input %d", n)
		}
	})
}
