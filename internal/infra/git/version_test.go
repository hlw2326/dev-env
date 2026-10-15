package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestVersionDoesNotEditDependency(t *testing.T) {
	for _, input := range []struct{ name, content string }{
		{"package.json", `{"dependencies":{"version":"9.0.0"},"version":"1.0.0"}`},
		{"pom.xml", `<project><parent><version>9.0.0</version></parent><dependencies><dependency><version>8.0.0</version></dependency></dependencies><version>1.0.0</version></project>`},
	} {
		start, end, value, ok := findVersionValue(input.name, []byte(input.content))
		if !ok || value != "1.0.0" || input.content[start:end] != value {
			t.Fatalf("wrong project version: %s %q", input.name, value)
		}
	}
	_, _, _, ok := findVersionValue("pom.xml", []byte(`<project><dependencies><dependency><version>1.0.0</version></dependency></dependencies></project>`))
	if ok {
		t.Fatal("must not infer project version from dependency")
	}
}

func TestNextVersionSequence(t *testing.T) {
	tests := []struct {
		current  string
		expected string
	}{
		{current: "1.0.0", expected: "1.0.01"},
		{current: "1.0.01", expected: "1.0.02"},
		{current: "1.0.20", expected: "1.0.21"},
		{current: "1.0.99", expected: "1.1.0"},
		{current: "1.99.99", expected: "2.0.0"},
		{current: "v2.4.09", expected: "v2.4.10"},
	}
	for _, item := range tests {
		t.Run(item.current, func(t *testing.T) {
			actual, err := nextVersion(item.current)
			if err != nil {
				t.Fatal(err)
			}
			if actual != item.expected {
				t.Fatalf("nextVersion(%q) = %q, want %q", item.current, actual, item.expected)
			}
		})
	}
}

func TestPackageVersionBumpPreservesManifest(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "package.json")
	content := []byte("{\n  \"name\": \"sample\",\n  \"version\": \"1.0.0\",\n  \"scripts\": {\"dev\": \"vite\"}\n}\n")
	if err := os.WriteFile(path, content, 0o600); err != nil {
		t.Fatal(err)
	}
	location, next, message, err := readVersion(root, "node")
	if err != nil {
		t.Fatal(err)
	}
	if message != "" || next != "1.0.01" || location.value != "1.0.0" {
		t.Fatalf("unexpected version metadata: location=%+v next=%q message=%q", location, next, message)
	}
	if _, err := writeNextVersion(root, location, next); err != nil {
		t.Fatal(err)
	}
	updated, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(updated) != "{\n  \"name\": \"sample\",\n  \"version\": \"1.0.01\",\n  \"scripts\": {\"dev\": \"vite\"}\n}\n" {
		t.Fatalf("manifest changed unexpectedly:\n%s", updated)
	}
}

func TestCargoVersionUsesPackageSection(t *testing.T) {
	root := t.TempDir()
	content := "[workspace.package]\nversion = \"9.9.9\"\n\n[package]\nname = \"sample\"\nversion = \"1.0.99\"\n"
	if err := os.WriteFile(filepath.Join(root, "Cargo.toml"), []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	location, next, message, err := readVersion(root, "rust")
	if err != nil {
		t.Fatal(err)
	}
	if message != "" || location.value != "1.0.99" || next != "1.1.0" {
		t.Fatalf("unexpected Cargo version metadata: location=%+v next=%q message=%q", location, next, message)
	}
}
