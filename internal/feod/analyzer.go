package feod

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/feod-architecture/feod-analyzer/internal/config"
	importparser "github.com/feod-architecture/feod-analyzer/internal/imports"
	"github.com/feod-architecture/feod-analyzer/internal/report"
)

const Version = "0.1.0"

const (
	RuleForbiddenLevelDependency = "forbidden-level-dependency"
	RuleDeepImport               = "deep-import"
	RuleExternalSubmoduleImport  = "external-submodule-import"
	RuleMissingPublicAPI         = "missing-public-api"
	RulePublicAPILeak            = "public-api-leak"
	RuleDirectGlobalImport       = "direct-global-import"
	RuleCycle                    = "cycle"
)

type Analyzer struct {
	cfg        *config.Config
	srcAbs     string
	levelSet   map[string]bool
	segmentSet map[string]bool
	nodes      map[string]*report.Node
	edges      map[string]*report.Edge
	files      map[string]*projectFile
}

type projectFile struct {
	AbsPath string
	RelPath string
	Class   classification
	Imports []report.ImportUsage
}

type classification struct {
	IsFEOD       bool
	NodeID       string
	Kind         report.NodeKind
	Level        string
	Name         string
	Path         string
	ParentID     string
	OwnerID      string
	Module       string
	Submodule    string
	CommonEntity string
	Page         string
	Deep         bool
}

func Analyze(cfg *config.Config) (*report.Report, error) {
	start := time.Now()
	analyzer := &Analyzer{
		cfg:        cfg,
		srcAbs:     cfg.SrcAbs(),
		levelSet:   cfg.LevelSet(),
		segmentSet: cfg.SegmentSet(),
		nodes:      map[string]*report.Node{},
		edges:      map[string]*report.Edge{},
		files:      map[string]*projectFile{},
	}

	for _, level := range cfg.Levels {
		analyzer.ensureLevelNode(level)
	}

	if err := analyzer.scanFiles(); err != nil {
		return nil, err
	}
	if err := analyzer.analyzePublicAPIs(); err != nil {
		return nil, err
	}
	if err := analyzer.analyzeReadmes(); err != nil {
		return nil, err
	}

	violations := analyzer.analyzeImports()
	violations = append(violations, analyzer.publicAPIViolations()...)
	violations = append(violations, analyzer.detectCycles()...)

	result := &report.Report{
		Meta: report.Meta{
			Tool:       "feod-analyzer",
			Version:    Version,
			RootDir:    cfg.RootDir,
			SrcDir:     cfg.SrcAbs(),
			Generated:  time.Now().UTC(),
			Schema:     "feod-report/v1",
			DurationMS: time.Since(start).Milliseconds(),
		},
		Nodes:      analyzer.sortedNodes(),
		Edges:      analyzer.sortedEdges(),
		Violations: sortViolations(violations),
		Files:      analyzer.sortedFiles(),
	}
	result.Recount()

	return result, nil
}

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
			if len(ruleIDs) > 0 {
				edge.RuleIDs = mergeStrings(edge.RuleIDs, ruleIDs)
			}
			edge.Status = maxStatus(edge.Status, statusForViolations(importViolations))
			violations = append(violations, importViolations...)
		}
	}

	return violations
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
	return violations
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
			if stored := a.edges[edgeKey(edge.Source, edge.Target)]; stored != nil {
				stored.Status = report.EdgeError
				stored.RuleIDs = mergeStrings(stored.RuleIDs, []string{RuleCycle})
			}
			violations = append(violations, report.Violation{
				Rule:       RuleCycle,
				Severity:   report.SeverityError,
				From:       edge.Source,
				To:         edge.Target,
				Message:    fmt.Sprintf("Обнаружен цикл зависимостей между %s и %s.", edge.Source, edge.Target),
				Suggestion: "Разорвите цикл через public API, перенос общего кода в common или пересмотр ответственности модулей.",
			})
		}
	}
	return violations
}

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

		if a.cfg.Submodules.Enabled && len(parts) >= 3 && !a.segmentSet[parts[2]] && !isIndexPart(parts[2]) {
			submodule := parts[2]
			class.Submodule = submodule
			class.Name = submodule
			class.Path = "modules/" + module + "/" + submodule
			class.NodeID = "submodule:" + module + "/" + submodule
			class.Kind = report.NodeSubmodule
			class.ParentID = "module:" + module
			class.OwnerID = "module:" + module
			class.Deep = len(parts) > 3 && !isIndexPart(parts[3])
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

func (a *Analyzer) ensureLevelNode(level string) {
	id := "level:" + level
	if _, ok := a.nodes[id]; ok {
		return
	}
	a.nodes[id] = &report.Node{
		ID:    id,
		Kind:  report.NodeLevel,
		Name:  level,
		Level: level,
		Path:  level,
		PublicAPI: report.PublicAPIInfo{
			Status: "level",
		},
	}
}

func (a *Analyzer) ensureNode(class classification) {
	a.ensureLevelNode(class.Level)
	if _, ok := a.nodes[class.NodeID]; ok {
		a.nodes[class.NodeID].FileCount++
		return
	}
	a.nodes[class.NodeID] = &report.Node{
		ID:        class.NodeID,
		Kind:      class.Kind,
		Name:      class.Name,
		Level:     class.Level,
		Path:      class.Path,
		ParentID:  class.ParentID,
		FileCount: 1,
		PublicAPI: report.PublicAPIInfo{
			Status: "unknown",
		},
	}
}

func (a *Analyzer) ensureEdge(source string, target string) *report.Edge {
	key := edgeKey(source, target)
	if edge, ok := a.edges[key]; ok {
		return edge
	}
	edge := &report.Edge{
		ID:     key,
		Source: source,
		Target: target,
		Status: report.EdgeAllowed,
	}
	a.edges[key] = edge
	return edge
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
	exposed := []string{}
	seen := map[string]bool{}

	for _, stmt := range append(importparser.ExtractNamedExports(data), importparser.ExtractStarExports(data)...) {
		if !strings.HasPrefix(stmt.Path, "./") {
			continue
		}
		target := strings.TrimPrefix(stmt.Path, "./")
		target = strings.Trim(strings.TrimSuffix(target, "/index"), "/")
		if target == "" || strings.Contains(target, "/") || a.segmentSet[target] {
			continue
		}
		if _, err := os.Stat(filepath.Join(a.srcAbs, modulePath, target)); err != nil {
			continue
		}
		if !seen[target] {
			seen[target] = true
			exposed = append(exposed, target)
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

func (a *Analyzer) sortedNodes() []report.Node {
	nodes := make([]report.Node, 0, len(a.nodes))
	for _, node := range a.nodes {
		nodes = append(nodes, *node)
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].Level != nodes[j].Level {
			return levelRank(nodes[i].Level) < levelRank(nodes[j].Level)
		}
		if nodes[i].Kind != nodes[j].Kind {
			return nodes[i].Kind < nodes[j].Kind
		}
		return nodes[i].ID < nodes[j].ID
	})
	return nodes
}

func (a *Analyzer) sortedEdges() []report.Edge {
	edges := make([]report.Edge, 0, len(a.edges))
	for _, edge := range a.edges {
		sort.Slice(edge.Imports, func(i, j int) bool {
			if edge.Imports[i].File != edge.Imports[j].File {
				return edge.Imports[i].File < edge.Imports[j].File
			}
			return edge.Imports[i].Line < edge.Imports[j].Line
		})
		sort.Strings(edge.RuleIDs)
		edges = append(edges, *edge)
	}
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].Source != edges[j].Source {
			return edges[i].Source < edges[j].Source
		}
		return edges[i].Target < edges[j].Target
	})
	return edges
}

func (a *Analyzer) sortedFiles() []report.FileReport {
	paths := make([]string, 0, len(a.files))
	for path := range a.files {
		paths = append(paths, path)
	}
	sort.Strings(paths)

	files := make([]report.FileReport, 0, len(paths))
	for _, path := range paths {
		file := a.files[path]
		files = append(files, report.FileReport{
			Path:    file.RelPath,
			NodeID:  file.Class.NodeID,
			Imports: file.Imports,
		})
	}
	return files
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

func levelRank(level string) int {
	switch level {
	case "app":
		return 0
	case "pages":
		return 1
	case "modules":
		return 2
	case "common":
		return 3
	case "global":
		return 4
	default:
		return 100
	}
}

func mergeStrings(existing []string, next []string) []string {
	seen := map[string]bool{}
	for _, item := range existing {
		seen[item] = true
	}
	for _, item := range next {
		if !seen[item] {
			existing = append(existing, item)
			seen[item] = true
		}
	}
	return existing
}

func edgeKey(source string, target string) string {
	return source + "->" + target
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

func DebugJSON(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
