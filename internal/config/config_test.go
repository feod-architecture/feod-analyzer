package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTsconfigAliases(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["src/*"],
      "~/*": ["app/*"]
    }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load(dir, "")
	if err != nil {
		t.Fatal(err)
	}

	if cfg.Aliases["@"] != "src" {
		t.Fatalf("expected @ alias to src, got %q", cfg.Aliases["@"])
	}
	if cfg.Aliases["~"] != "app" {
		t.Fatalf("expected ~ alias to app, got %q", cfg.Aliases["~"])
	}
}

func TestFindConfigWalksParents(t *testing.T) {
	dir := t.TempDir()
	project := filepath.Join(dir, "project")
	nested := filepath.Join(project, "src", "modules")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".feod-analyzer.yml"), []byte("srcDir: src\n"), 0644); err != nil {
		t.Fatal(err)
	}

	path, err := FindConfig(nested)
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != ".feod-analyzer.yml" {
		t.Fatalf("unexpected config path %s", path)
	}
}
