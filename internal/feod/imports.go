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

func (a *Analyzer) analyzeImports() []report.Violation {
	violations := []report.Violation{}

	filePaths := make([]string, 0, len(a.files))
	for rel := range a.files {
		filePaths = append(filePaths, rel)
	}
	sort.Strings(filePaths)

	for _, rel := range filePaths {
		file := a.files[rel]
		data, err := os.ReadFile(file.AbsPath)
		if err != nil {
			violations = append(violations, report.Violation{
				Rule:     "read-file",
				Severity: report.SeverityError,
				File:     file.RelPath,
				Message:  fmt.Sprintf("Не удалось прочитать файл: %v", err),
			})
			continue
		}

		for _, stmt := range importparser.Extract(data) {
			resolved, internal := a.resolveImport(file.RelPath, stmt.Path)
			usage := report.ImportUsage{
				File:         file.RelPath,
				Line:         stmt.Line,
				ImportPath:   stmt.Path,
				ResolvedPath: resolved,
				Kind:         stmt.Kind,
				TypeOnly:     stmt.TypeOnly,
			}
			file.Imports = append(file.Imports, usage)
			if !internal || !file.Class.IsFEOD {
				continue
			}

			target := a.classify(resolved)
			if !target.IsFEOD || target.NodeID == file.Class.NodeID {
				continue
			}

			a.ensureNode(target)
			edge := a.ensureEdge(file.Class.NodeID, target.NodeID)
			edge.Imports = append(edge.Imports, usage)

			ruleIDs, importViolations := a.evaluateImport(file.Class, target, resolved, usage)
			importViolations = a.filterIgnoredViolations(importViolations)
			ruleIDs = ruleIDsForViolations(importViolations, ruleIDs)
			if len(ruleIDs) > 0 {
				edge.RuleIDs = mergeStrings(edge.RuleIDs, ruleIDs)
			}
			edge.Status = maxStatus(edge.Status, statusForViolations(importViolations))
			violations = append(violations, importViolations...)
		}
	}

	return a.filterIgnoredViolations(violations)
}
func (a *Analyzer) evaluateImport(from classification, to classification, targetPath string, usage report.ImportUsage) ([]string, []report.Violation) {
	rules := []string{}
	violations := []report.Violation{}
	externalOwner := from.OwnerID != to.OwnerID

	if to.Level == "global" && from.Level != "global" {
		rules = append(rules, RuleDirectGlobalImport)
		violations = append(violations, report.Violation{
			Rule:       RuleDirectGlobalImport,
			Severity:   report.SeverityError,
			File:       usage.File,
			Line:       usage.Line,
			From:       from.NodeID,
			To:         to.NodeID,
			ImportPath: usage.ImportPath,
			Message:    "Прикладной FEOD-код не должен импортировать global напрямую.",
			Suggestion: "Подключите global через entrypoint, runtime bootstrap или конфигурацию сборщика.",
		})
		return rules, violations
	}

	if !allowedLevelDependency(from.Level, to.Level) {
		rules = append(rules, RuleForbiddenLevelDependency)
		violations = append(violations, report.Violation{
			Rule:       RuleForbiddenLevelDependency,
			Severity:   report.SeverityError,
			File:       usage.File,
			Line:       usage.Line,
			From:       from.NodeID,
			To:         to.NodeID,
			ImportPath: usage.ImportPath,
			Message:    fmt.Sprintf("Запрещённая зависимость FEOD: %s импортирует %s.", from.Level, to.Level),
			Suggestion: "Проверьте матрицу импортов FEOD и перенесите контракт на разрешённый уровень.",
		})
	}

	if to.Kind == report.NodeSubmodule && externalOwner {
		rules = append(rules, RuleExternalSubmoduleImport)
		violations = append(violations, report.Violation{
			Rule:       RuleExternalSubmoduleImport,
			Severity:   report.SeverityError,
			File:       usage.File,
			Line:       usage.Line,
			From:       from.NodeID,
			To:         to.NodeID,
			ImportPath: usage.ImportPath,
			Message:    "Внешний код импортирует подмодуль напрямую, обходя public API родительского модуля.",
			Suggestion: fmt.Sprintf("Экспортируйте нужный контракт из %s/index.ts и импортируйте из %q.", ownerPath(to), "@/"+ownerPath(to)),
		})
		return rules, violations
	}

	if to.Deep && externalOwner {
		rules = append(rules, RuleDeepImport)
		violations = append(violations, report.Violation{
			Rule:       RuleDeepImport,
			Severity:   report.SeverityError,
			File:       usage.File,
			Line:       usage.Line,
			From:       from.NodeID,
			To:         to.NodeID,
			ImportPath: usage.ImportPath,
			Message:    "Импорт обходит public API FEOD-сущности и зависит от её внутренней структуры.",
			Suggestion: fmt.Sprintf("Используйте public API: %q.", "@/"+publicPath(to)),
		})
	}

	return rules, violations
}
func (a *Analyzer) resolveImport(fromRel string, importPath string) (string, bool) {
	if strings.HasPrefix(importPath, ".") {
		resolved := filepath.ToSlash(filepath.Clean(filepath.Join(filepath.Dir(fromRel), importPath)))
		return normalizeImportPath(resolved), true
	}

	for alias, target := range a.cfg.Aliases {
		if importPath == alias || strings.HasPrefix(importPath, alias+"/") {
			suffix := strings.TrimPrefix(importPath, alias)
			suffix = strings.TrimPrefix(suffix, "/")
			resolved := filepath.ToSlash(filepath.Clean(filepath.Join(target, suffix)))
			return a.trimSrcPrefix(normalizeImportPath(resolved)), true
		}
	}

	first := importPath
	if slash := strings.Index(first, "/"); slash >= 0 {
		first = first[:slash]
	}
	if a.levelSet[first] {
		return normalizeImportPath(importPath), true
	}

	return "", false
}
func (a *Analyzer) trimSrcPrefix(path string) string {
	src := filepath.ToSlash(a.cfg.SrcDir)
	src = strings.TrimPrefix(src, "./")
	src = strings.TrimSuffix(src, "/")
	if src != "" && (path == src || strings.HasPrefix(path, src+"/")) {
		return strings.TrimPrefix(strings.TrimPrefix(path, src), "/")
	}
	return path
}
