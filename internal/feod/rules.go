package feod

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/feod-architecture/feod-analyzer/internal/report"
)

func (a *Analyzer) publicAPIViolations() []report.Violation {
	violations := []report.Violation{}
	for _, node := range a.sortedNodes() {
		if node.Kind == report.NodeModule || node.Kind == report.NodeCommonEntity || node.Kind == report.NodePage {
			if !node.PublicAPI.HasIndex {
				violations = append(violations, report.Violation{
					Rule:       RuleMissingPublicAPI,
					Severity:   report.SeverityWarning,
					File:       node.Path,
					From:       node.ID,
					Message:    fmt.Sprintf("%s не имеет явного root index.ts public API.", node.Path),
					Suggestion: "Добавьте root index.ts и экспортируйте только поддерживаемый контракт.",
				})
			}
			for _, exportPath := range node.PublicAPI.StarExports {
				if a.isPublicAPILeak(node.Path, exportPath) {
					violations = append(violations, report.Violation{
						Rule:       RulePublicAPILeak,
						Severity:   report.SeverityWarning,
						File:       node.PublicAPI.IndexPath,
						From:       node.ID,
						ImportPath: exportPath,
						Message:    "Public API использует export * из внутренней структуры.",
						Suggestion: "Замените export * явным списком поддерживаемых экспортов.",
					})
				}
			}
		}
	}
	return a.filterIgnoredViolations(violations)
}
func (a *Analyzer) detectCycles() []report.Violation {
	adjacency := map[string][]string{}
	for _, edge := range a.edges {
		adjacency[edge.Source] = append(adjacency[edge.Source], edge.Target)
	}
	for source := range adjacency {
		sort.Strings(adjacency[source])
	}

	violations := []report.Violation{}
	seen := map[string]bool{}
	for _, edge := range a.sortedEdges() {
		if edge.Source == edge.Target {
			continue
		}
		if hasPath(adjacency, edge.Target, edge.Source, map[string]bool{}) {
			key := edge.Source + "->" + edge.Target
			if seen[key] {
				continue
			}
			seen[key] = true
			violation := report.Violation{
				Rule:       RuleCycle,
				Severity:   report.SeverityError,
				From:       edge.Source,
				To:         edge.Target,
				Message:    fmt.Sprintf("Обнаружен цикл зависимостей между %s и %s.", edge.Source, edge.Target),
				Suggestion: "Разорвите цикл через public API, перенос общего кода в common или пересмотр ответственности модулей.",
			}
			if a.isIgnoredViolation(violation) {
				continue
			}
			if stored := a.edges[edgeKey(edge.Source, edge.Target)]; stored != nil {
				stored.Status = report.EdgeError
				stored.RuleIDs = mergeStrings(stored.RuleIDs, []string{RuleCycle})
			}
			violations = append(violations, violation)
		}
	}
	return violations
}
func allowedLevelDependency(from string, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "app":
		return to == "pages" || to == "modules" || to == "common"
	case "pages":
		return to == "modules" || to == "common"
	case "modules":
		return to == "modules" || to == "common"
	case "common":
		return to == "common"
	case "global":
		return false
	default:
		return true
	}
}
func statusForViolations(violations []report.Violation) report.EdgeStatus {
	status := report.EdgeAllowed
	for _, violation := range violations {
		if violation.Severity == report.SeverityError {
			return report.EdgeError
		}
		if violation.Severity == report.SeverityWarning {
			status = report.EdgeWarning
		}
	}
	return status
}
func (a *Analyzer) filterIgnoredViolations(violations []report.Violation) []report.Violation {
	if len(a.cfg.IgnoreRules) == 0 || len(violations) == 0 {
		return violations
	}
	filtered := violations[:0]
	for _, violation := range violations {
		if !a.isIgnoredViolation(violation) {
			filtered = append(filtered, violation)
		}
	}
	return filtered
}
func (a *Analyzer) isIgnoredViolation(violation report.Violation) bool {
	for _, rule := range a.cfg.IgnoreRules {
		if rule.Rule == "" && rule.Path == "" && rule.ImportPath == "" {
			continue
		}
		if rule.Rule != "" && rule.Rule != violation.Rule {
			continue
		}
		if rule.Path != "" && !matchIgnorePattern(rule.Path, violation.File) {
			continue
		}
		if rule.ImportPath != "" && !matchIgnorePattern(rule.ImportPath, violation.ImportPath) {
			continue
		}
		return true
	}
	return false
}
func matchIgnorePattern(pattern string, value string) bool {
	pattern = filepath.ToSlash(strings.TrimSpace(pattern))
	value = filepath.ToSlash(strings.TrimSpace(value))
	if pattern == "" {
		return true
	}
	if value == "" {
		return false
	}
	if pattern == value {
		return true
	}
	if ok, err := filepath.Match(pattern, value); err == nil && ok {
		return true
	}
	return strings.HasSuffix(pattern, "/") && strings.HasPrefix(value, pattern)
}
func ruleIDsForViolations(violations []report.Violation, fallback []string) []string {
	if len(violations) == 0 {
		return nil
	}
	rules := make([]string, 0, len(violations))
	for _, violation := range violations {
		rules = mergeStrings(rules, []string{violation.Rule})
	}
	if len(rules) == 0 {
		return fallback
	}
	return rules
}
func maxStatus(current report.EdgeStatus, next report.EdgeStatus) report.EdgeStatus {
	if current == report.EdgeError || next == report.EdgeError {
		return report.EdgeError
	}
	if current == report.EdgeWarning || next == report.EdgeWarning {
		return report.EdgeWarning
	}
	return report.EdgeAllowed
}
func hasPath(adjacency map[string][]string, from string, to string, visited map[string]bool) bool {
	if from == to {
		return true
	}
	if visited[from] {
		return false
	}
	visited[from] = true
	for _, next := range adjacency[from] {
		if hasPath(adjacency, next, to, visited) {
			return true
		}
	}
	return false
}
func sortViolations(violations []report.Violation) []report.Violation {
	sort.Slice(violations, func(i, j int) bool {
		if violations[i].Severity != violations[j].Severity {
			return severityRank(violations[i].Severity) < severityRank(violations[j].Severity)
		}
		if violations[i].File != violations[j].File {
			return violations[i].File < violations[j].File
		}
		if violations[i].Line != violations[j].Line {
			return violations[i].Line < violations[j].Line
		}
		return violations[i].Rule < violations[j].Rule
	})
	return violations
}
func severityRank(severity report.Severity) int {
	switch severity {
	case report.SeverityError:
		return 0
	case report.SeverityWarning:
		return 1
	default:
		return 2
	}
}
