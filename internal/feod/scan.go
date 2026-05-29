package feod

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (a *Analyzer) scanFiles() error {
	if _, err := os.Stat(a.srcAbs); err != nil {
		return fmt.Errorf("source directory %s: %w", a.srcAbs, err)
	}

	return filepath.WalkDir(a.srcAbs, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if entry.IsDir() {
			if path != a.srcAbs && a.isExcluded(entry.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		if !isSourceFile(path) {
			return nil
		}

		rel, err := filepath.Rel(a.srcAbs, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		class := a.classify(rel)
		if class.IsFEOD {
			a.ensureNode(class)
		}
		a.files[rel] = &projectFile{AbsPath: path, RelPath: rel, Class: class}
		return nil
	})
}
func (a *Analyzer) isExcluded(name string) bool {
	if strings.HasPrefix(name, ".") && name != "." {
		for _, allowed := range []string{".storybook"} {
			if name == allowed {
				return false
			}
		}
		return true
	}
	for _, excluded := range a.cfg.ExcludeDirs {
		if name == excluded {
			return true
		}
	}
	return false
}
func isSourceFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".ts", ".tsx", ".js", ".jsx", ".vue", ".mjs", ".cjs":
		return true
	default:
		return false
	}
}
