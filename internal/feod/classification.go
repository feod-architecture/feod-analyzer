package feod

import (
	"path/filepath"
	"strings"

	"github.com/feod-architecture/feod-analyzer/internal/report"
)

func (a *Analyzer) classify(path string) classification {
	normalized := normalizeImportPath(path)
	parts := splitPath(normalized)
	if len(parts) == 0 || !a.levelSet[parts[0]] {
		return classification{}
	}

	level := parts[0]
	class := classification{
		IsFEOD:  true,
		Level:   level,
		Path:    level,
		NodeID:  "level:" + level,
		Kind:    report.NodeLevel,
		Name:    level,
		OwnerID: "level:" + level,
	}

	switch level {
	case "modules":
		if len(parts) < 2 {
			return class
		}
		module := parts[1]
		class.Module = module
		class.Name = module
		class.Path = "modules/" + module
		class.NodeID = "module:" + module
		class.Kind = report.NodeModule
		class.ParentID = "level:modules"
		class.OwnerID = class.NodeID

		if a.cfg.Submodules.Enabled && a.cfg.Submodules.MaxDepth > 0 && len(parts) >= 3 {
			submoduleParts := []string{}
			cursor := 2
			for cursor < len(parts) && len(submoduleParts) < a.cfg.Submodules.MaxDepth {
				part := parts[cursor]
				if a.segmentSet[part] || isIndexPart(part) {
					break
				}
				submoduleParts = append(submoduleParts, part)
				cursor++
			}
			if len(submoduleParts) == 0 {
				class.Deep = len(parts) > 2 && !isIndexPart(parts[2])
				return class
			}

			submodule := strings.Join(submoduleParts, "/")
			class.Submodule = submodule
			class.Name = submodule
			class.Path = "modules/" + module + "/" + submodule
			class.NodeID = "submodule:" + module + "/" + submodule
			class.Kind = report.NodeSubmodule
			class.ParentID = "module:" + module
			class.OwnerID = "module:" + module
			class.Deep = cursor < len(parts) && !isIndexPart(parts[cursor])
			return class
		}

		class.Deep = len(parts) > 2 && !isIndexPart(parts[2])
		return class
	case "pages":
		if len(parts) < 2 {
			return class
		}
		page := parts[1]
		class.Page = page
		class.Name = page
		class.Path = "pages/" + page
		class.NodeID = "page:" + page
		class.Kind = report.NodePage
		class.ParentID = "level:pages"
		class.OwnerID = class.NodeID
		class.Deep = len(parts) > 2 && !isIndexPart(parts[2])
		return class
	case "common":
		if len(parts) < 2 {
			return class
		}
		entity := parts[1]
		class.CommonEntity = entity
		class.Name = entity
		class.Path = "common/" + entity
		class.NodeID = "common:" + entity
		class.Kind = report.NodeCommonEntity
		class.ParentID = "level:common"
		class.OwnerID = class.NodeID
		class.Deep = len(parts) > 2 && !isIndexPart(parts[2])
		return class
	case "app", "global":
		class.Deep = len(parts) > 1
		return class
	default:
		return class
	}
}
func normalizeImportPath(path string) string {
	clean := filepath.ToSlash(filepath.Clean(strings.TrimSpace(path)))
	clean = strings.TrimPrefix(clean, "./")
	clean = strings.TrimPrefix(clean, "/")
	for _, ext := range []string{".d.ts", ".tsx", ".ts", ".jsx", ".js", ".vue", ".mjs", ".cjs"} {
		clean = strings.TrimSuffix(clean, ext)
	}
	clean = strings.TrimSuffix(clean, "/index")
	return clean
}
func splitPath(path string) []string {
	path = strings.Trim(path, "/")
	if path == "" || path == "." {
		return nil
	}
	return strings.Split(path, "/")
}
func isIndexPart(part string) bool {
	return part == "index" || strings.HasPrefix(part, "index.")
}
func ownerPath(class classification) string {
	if class.Level == "modules" && class.Module != "" {
		return "modules/" + class.Module
	}
	return publicPath(class)
}
func publicPath(class classification) string {
	switch class.Kind {
	case report.NodeSubmodule:
		return "modules/" + class.Module
	case report.NodeModule, report.NodeCommonEntity, report.NodePage:
		return class.Path
	case report.NodeLevel:
		return class.Level
	default:
		return class.Path
	}
}
