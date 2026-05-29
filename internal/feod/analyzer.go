package feod

import (
	"encoding/json"
	"time"

	"github.com/feod-architecture/feod-analyzer/internal/config"
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

func DebugJSON(v any) string {
	data, _ := json.MarshalIndent(v, "", "  ")
	return string(data)
}
