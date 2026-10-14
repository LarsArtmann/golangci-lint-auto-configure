package linter

import (
	"fmt"
	"testing"
)

func TestDebugUsesTech2(t *testing.T) {
	env := Env{Tech: map[string]bool{"clickhouse": false}}
	known, detected := env.Tech["clickhouse"]
	fmt.Printf("A: known=%v detected=%v\n", known, detected)
	fmt.Printf("B: UsesTech=%v\n", env.UsesTech("clickhouse"))
	known2, detected2 := env.Tech["clickhouse"]
	fmt.Printf("C: known=%v detected=%v\n", known2, detected2)
}
