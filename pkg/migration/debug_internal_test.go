package migration

import (
	"fmt"
	"testing"
)

func TestDebugGocritic(t *testing.T) {
	settings := map[string]any{
		"gocritic": map[string]any{
			"settings": map[string]any{
				"ruleguard": map[string]any{
					"rules": "rules.go,${configDir}/more-rules.go",
				},
			},
		},
	}
	fixes := migrateLinterSettings(settings, DefaultRules())
	fmt.Printf("fixes=%d rules=%q\n", fixes, settings["gocritic"].(map[string]any)["settings"].(map[string]any)["rules"])
}
