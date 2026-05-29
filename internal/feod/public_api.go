package feod

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	importparser "github.com/feod-architecture/feod-analyzer/internal/imports"
	"github.com/feod-architecture/feod-analyzer/internal/report"
)

func (a *Analyzer) analyzePublicAPIs() error {
	for id, node := range a.nodes {
		if node.Kind == report.NodeLevel || node.Kind == report.NodeFile {
			continue
		}
		if node.Kind == report.NodeSubmodule {
			node.PublicAPI.Status = "internal"
		}

		indexPath, ok := a.findIndex(node.Path)
		if !ok {
			if node.Kind != report.NodeSubmodule {
				node.PublicAPI.Status = "missing"
			}
			a.nodes[id] = node
			continue
		}

		data, err := os.ReadFile(filepath.Join(a.srcAbs, indexPath))
		if err != nil {
			return fmt.Errorf("read public API %s: %w", indexPath, err)
		}

		node.PublicAPI.HasIndex = true
		node.PublicAPI.IndexPath = indexPath
		if node.PublicAPI.Status == "" {
			node.PublicAPI.Status = "explicit"
		}

		for _, stmt := range importparser.Extract(data) {
			if stmt.Kind == "export" {
				node.PublicAPI.Exports = append(node.PublicAPI.Exports, stmt.Path)
			}
		}
		for _, stmt := range importparser.ExtractStarExports(data) {
			node.PublicAPI.StarExports = append(node.PublicAPI.StarExports, stmt.Path)
		}
		if node.Kind == report.NodeModule {
			node.PublicAPI.ExposedSubmodules = a.detectExposedSubmodules(node.Path, data)
		}
		a.nodes[id] = node
	}

	return nil
}
func (a *Analyzer) analyzeReadmes() error {
	for id, node := range a.nodes {
		if node.Kind != report.NodeModule && node.Kind != report.NodeSubmodule {
			continue
		}

		readmePath := filepath.ToSlash(filepath.Join(node.Path, "README.md"))
		data, err := os.ReadFile(filepath.Join(a.srcAbs, readmePath))
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return fmt.Errorf("read README %s: %w", readmePath, err)
		}

		node.Readme = &report.ReadmeInfo{
			Path:    readmePath,
			Content: string(data),
		}
		a.nodes[id] = node
	}
	return nil
}
func (a *Analyzer) findIndex(entityPath string) (string, bool) {
	for _, name := range []string{"index.ts", "index.tsx", "index.js", "index.jsx", "index.mjs"} {
		rel := filepath.ToSlash(filepath.Join(entityPath, name))
		if _, err := os.Stat(filepath.Join(a.srcAbs, rel)); err == nil {
			return rel, true
		}
	}
	return "", false
}
func (a *Analyzer) detectExposedSubmodules(modulePath string, data []byte) []string {
	if !a.cfg.Submodules.Enabled || a.cfg.Submodules.MaxDepth <= 0 {
		return nil
	}

	exposed := []string{}
	seen := map[string]bool{}

	for _, stmt := range append(importparser.ExtractNamedExports(data), importparser.ExtractStarExports(data)...) {
		if !strings.HasPrefix(stmt.Path, "./") {
			continue
		}
		target := strings.TrimPrefix(stmt.Path, "./")
		target = strings.Trim(strings.TrimSuffix(target, "/index"), "/")
		targetParts := splitPath(target)
		if len(targetParts) == 0 || len(targetParts) > a.cfg.Submodules.MaxDepth {
			continue
		}

		validSubmodule := true
		for _, part := range targetParts {
			if a.segmentSet[part] || isIndexPart(part) {
				validSubmodule = false
				break
			}
		}
		if !validSubmodule {
			continue
		}

		submodule := strings.Join(targetParts, "/")
		if _, err := os.Stat(filepath.Join(a.srcAbs, modulePath, submodule)); err != nil {
			continue
		}
		if !seen[submodule] {
			seen[submodule] = true
			exposed = append(exposed, submodule)
		}
	}

	sort.Strings(exposed)
	return exposed
}
func (a *Analyzer) isPublicAPILeak(entityPath string, exportPath string) bool {
	if !strings.HasPrefix(exportPath, ".") {
		return false
	}
	target := normalizeImportPath(filepath.ToSlash(filepath.Join(entityPath, exportPath)))
	parts := splitPath(strings.TrimPrefix(strings.TrimPrefix(target, entityPath), "/"))
	if len(parts) == 0 {
		return false
	}
	if a.segmentSet[parts[0]] {
		return true
	}
	return strings.Contains(exportPath, "*")
}
