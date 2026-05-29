package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type SubmodulesConfig struct {
	Enabled  bool `yaml:"enabled" json:"enabled"`
	MaxDepth int  `yaml:"maxDepth" json:"maxDepth"`
}

type IgnoreRule struct {
	Rule       string `yaml:"rule" json:"rule"`
	Path       string `yaml:"path" json:"path"`
	ImportPath string `yaml:"importPath" json:"importPath"`
	Reason     string `yaml:"reason" json:"reason"`
}

type Config struct {
	RootDir       string            `yaml:"-" json:"-"`
	SrcDir        string            `yaml:"srcDir" json:"srcDir"`
	OutputDir     string            `yaml:"outputDir" json:"outputDir"`
	OutputFormats []string          `yaml:"outputFormats" json:"outputFormats"`
	ExcludeDirs   []string          `yaml:"excludeDirs" json:"excludeDirs"`
	Aliases       map[string]string `yaml:"aliases" json:"aliases"`
	Tsconfig      string            `yaml:"tsconfig" json:"tsconfig"`
	Levels        []string          `yaml:"levels" json:"levels"`
	Segments      []string          `yaml:"segments" json:"segments"`
	Submodules    SubmodulesConfig  `yaml:"submodules" json:"submodules"`
	IgnoreRules   []IgnoreRule      `yaml:"ignoreRules" json:"ignoreRules"`
}

var Default = Config{
	SrcDir:        "src",
	OutputDir:     "dist/feod",
	OutputFormats: []string{"html", "json"},
	ExcludeDirs:   []string{"node_modules", ".git", "dist", "build", "coverage", ".next", ".nuxt", ".vite"},
	Aliases: map[string]string{
		"@": "src",
	},
	Tsconfig: "tsconfig.json",
	Levels:   []string{"app", "pages", "modules", "common", "global"},
	Segments: []string{"ui", "model", "api", "lib", "config", "types", "test", "tests"},
	Submodules: SubmodulesConfig{
		Enabled:  true,
		MaxDepth: 2,
	},
	IgnoreRules: []IgnoreRule{},
}

var configNames = []string{
	"feod-analyzer.yml",
	"feod-analyzer.yaml",
	".feod-analyzer.yml",
	".feod-analyzer.yaml",
}

func Load(rootDir string, explicitPath string) (*Config, error) {
	absRoot, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, fmt.Errorf("resolve root dir: %w", err)
	}

	cfg := cloneDefault()
	cfg.RootDir = absRoot

	configPath := explicitPath
	if configPath == "" {
		configPath, err = FindConfig(absRoot)
		if err != nil {
			return nil, err
		}
	}

	if configPath != "" {
		if !filepath.IsAbs(configPath) {
			configPath = filepath.Join(absRoot, configPath)
		}
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("read config %s: %w", configPath, err)
		}
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			return nil, fmt.Errorf("parse config %s: %w", configPath, err)
		}
		cfg.RootDir = absRoot
	}

	cfg.normalize()
	if err := cfg.LoadTsconfigAliases(); err != nil && explicitPath != "" {
		return nil, err
	}

	return &cfg, nil
}

func cloneDefault() Config {
	cfg := Default
	cfg.OutputFormats = append([]string(nil), Default.OutputFormats...)
	cfg.ExcludeDirs = append([]string(nil), Default.ExcludeDirs...)
	cfg.Aliases = cloneStringMap(Default.Aliases)
	cfg.Levels = append([]string(nil), Default.Levels...)
	cfg.Segments = append([]string(nil), Default.Segments...)
	cfg.IgnoreRules = append([]IgnoreRule(nil), Default.IgnoreRules...)
	return cfg
}

func cloneStringMap(source map[string]string) map[string]string {
	target := make(map[string]string, len(source))
	for key, value := range source {
		target[key] = value
	}
	return target
}

func FindConfig(startDir string) (string, error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", fmt.Errorf("resolve config search dir: %w", err)
	}

	for {
		for _, name := range configNames {
			candidate := filepath.Join(dir, name)
			if _, err := os.Stat(candidate); err == nil {
				return candidate, nil
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", nil
		}
		dir = parent
	}
}

func (cfg *Config) SrcAbs() string {
	if filepath.IsAbs(cfg.SrcDir) {
		return filepath.Clean(cfg.SrcDir)
	}
	return filepath.Clean(filepath.Join(cfg.RootDir, cfg.SrcDir))
}

func (cfg *Config) OutputAbs() string {
	if filepath.IsAbs(cfg.OutputDir) {
		return filepath.Clean(cfg.OutputDir)
	}
	return filepath.Clean(filepath.Join(cfg.RootDir, cfg.OutputDir))
}

func (cfg *Config) LevelSet() map[string]bool {
	result := make(map[string]bool, len(cfg.Levels))
	for _, level := range cfg.Levels {
		result[level] = true
	}
	return result
}

func (cfg *Config) SegmentSet() map[string]bool {
	result := make(map[string]bool, len(cfg.Segments))
	for _, segment := range cfg.Segments {
		result[segment] = true
	}
	return result
}

func (cfg *Config) normalize() {
	if cfg.SrcDir == "" {
		cfg.SrcDir = Default.SrcDir
	}
	if cfg.OutputDir == "" {
		cfg.OutputDir = Default.OutputDir
	}
	if len(cfg.OutputFormats) == 0 {
		cfg.OutputFormats = Default.OutputFormats
	}
	if len(cfg.ExcludeDirs) == 0 {
		cfg.ExcludeDirs = Default.ExcludeDirs
	}
	if cfg.Aliases == nil {
		cfg.Aliases = map[string]string{}
	}
	if cfg.Tsconfig == "" {
		cfg.Tsconfig = Default.Tsconfig
	}
	if len(cfg.Levels) == 0 {
		cfg.Levels = Default.Levels
	}
	if len(cfg.Segments) == 0 {
		cfg.Segments = Default.Segments
	}
	if cfg.Submodules.MaxDepth == 0 {
		cfg.Submodules.MaxDepth = Default.Submodules.MaxDepth
	}
}

type tsConfigFile struct {
	CompilerOptions struct {
		BaseURL string              `json:"baseUrl"`
		Paths   map[string][]string `json:"paths"`
	} `json:"compilerOptions"`
}

func (cfg *Config) LoadTsconfigAliases() error {
	tsconfigPath := cfg.Tsconfig
	if tsconfigPath == "" {
		return nil
	}
	if !filepath.IsAbs(tsconfigPath) {
		tsconfigPath = filepath.Join(cfg.RootDir, tsconfigPath)
	}

	data, err := os.ReadFile(tsconfigPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("read tsconfig aliases: %w", err)
	}

	var parsed tsConfigFile
	if err := json.Unmarshal(stripJSONComments(data), &parsed); err != nil {
		return fmt.Errorf("parse tsconfig aliases: %w", err)
	}

	baseURL := parsed.CompilerOptions.BaseURL
	for pattern, targets := range parsed.CompilerOptions.Paths {
		if len(targets) == 0 {
			continue
		}

		alias := strings.TrimSuffix(pattern, "/*")
		target := strings.TrimSuffix(targets[0], "/*")
		if baseURL != "" && !strings.HasPrefix(target, baseURL) && !filepath.IsAbs(target) {
			target = filepath.Join(baseURL, target)
		}
		if alias != "" && target != "" {
			cfg.Aliases[alias] = filepath.ToSlash(target)
		}
	}

	return nil
}

func stripJSONComments(data []byte) []byte {
	text := string(data)
	var out strings.Builder
	inString := false
	escaped := false

	for i := 0; i < len(text); i++ {
		ch := text[i]
		next := byte(0)
		if i+1 < len(text) {
			next = text[i+1]
		}

		if inString {
			out.WriteByte(ch)
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}

		if ch == '"' {
			inString = true
			out.WriteByte(ch)
			continue
		}

		if ch == '/' && next == '/' {
			for i < len(text) && text[i] != '\n' {
				i++
			}
			if i < len(text) {
				out.WriteByte('\n')
			}
			continue
		}

		if ch == '/' && next == '*' {
			i += 2
			for i+1 < len(text) && !(text[i] == '*' && text[i+1] == '/') {
				if text[i] == '\n' {
					out.WriteByte('\n')
				}
				i++
			}
			i++
			continue
		}

		out.WriteByte(ch)
	}

	return []byte(out.String())
}
