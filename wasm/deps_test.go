package wasm

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestNoForbiddenDeps(t *testing.T) {
	cmd := exec.Command("go", "list", "-deps", "./...")

	cmd.Env = append(os.Environ(), "GOOS=js", "GOARCH=wasm")

	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%v\n%s", err, out)
	}

	for _, l := range strings.Split(string(out), "\n") {
		if strings.Contains(l, "elk") || strings.Contains(l, "dagre") || strings.Contains(l, "d2js/d2wasm") {
			t.Errorf("forbidden dep %s", l)
		}
	}
}
