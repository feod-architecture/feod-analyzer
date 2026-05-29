package report

import "time"

type Severity string

const (
	SeverityInfo    Severity = "info"
	SeverityWarning Severity = "warning"
	SeverityError   Severity = "error"
)

type EdgeStatus string

const (
	EdgeAllowed EdgeStatus = "allowed"
	EdgeWarning EdgeStatus = "warning"
	EdgeError   EdgeStatus = "error"
)

type NodeKind string

const (
	NodeLevel        NodeKind = "level"
	NodeModule       NodeKind = "module"
	NodeSubmodule    NodeKind = "submodule"
	NodeCommonEntity NodeKind = "commonEntity"
	NodePage         NodeKind = "page"
	NodeFile         NodeKind = "file"
)

type PublicAPIInfo struct {
	HasIndex          bool     `json:"hasIndex"`
	IndexPath         string   `json:"indexPath,omitempty"`
	Status            string   `json:"status"`
	Exports           []string `json:"exports,omitempty"`
	StarExports       []string `json:"starExports,omitempty"`
	ExposedSubmodules []string `json:"exposedSubmodules,omitempty"`
}

type ReadmeInfo struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Node struct {
	ID        string        `json:"id"`
	Kind      NodeKind      `json:"kind"`
	Name      string        `json:"name"`
	Level     string        `json:"level"`
	Path      string        `json:"path"`
	ParentID  string        `json:"parentId,omitempty"`
	PublicAPI PublicAPIInfo `json:"publicApi"`
	FileCount int           `json:"fileCount"`
	Readme    *ReadmeInfo   `json:"readme,omitempty"`
}

type ImportUsage struct {
	File         string `json:"file"`
	Line         int    `json:"line"`
	ImportPath   string `json:"importPath"`
	ResolvedPath string `json:"resolvedPath,omitempty"`
	Kind         string `json:"kind"`
	TypeOnly     bool   `json:"typeOnly"`
}

type Edge struct {
	ID      string        `json:"id"`
	Source  string        `json:"source"`
	Target  string        `json:"target"`
	Imports []ImportUsage `json:"imports"`
	Status  EdgeStatus    `json:"status"`
	RuleIDs []string      `json:"ruleIds,omitempty"`
}

type Violation struct {
	Rule       string   `json:"rule"`
	Severity   Severity `json:"severity"`
	File       string   `json:"file,omitempty"`
	Line       int      `json:"line,omitempty"`
	From       string   `json:"from,omitempty"`
	To         string   `json:"to,omitempty"`
	ImportPath string   `json:"importPath,omitempty"`
	Message    string   `json:"message"`
	Suggestion string   `json:"suggestion,omitempty"`
}

type FileReport struct {
	Path    string        `json:"path"`
	NodeID  string        `json:"nodeId,omitempty"`
	Imports []ImportUsage `json:"imports,omitempty"`
}

type Summary struct {
	Files       int `json:"files"`
	Nodes       int `json:"nodes"`
	Edges       int `json:"edges"`
	Errors      int `json:"errors"`
	Warnings    int `json:"warnings"`
	Infos       int `json:"infos"`
	Violations  int `json:"violations"`
	Modules     int `json:"modules"`
	Submodules  int `json:"submodules"`
	Pages       int `json:"pages"`
	CommonItems int `json:"commonItems"`
}

type Meta struct {
	Tool       string    `json:"tool"`
	Version    string    `json:"version"`
	RootDir    string    `json:"rootDir"`
	SrcDir     string    `json:"srcDir"`
	Generated  time.Time `json:"generated"`
	Schema     string    `json:"schema"`
	DurationMS int64     `json:"durationMs"`
}

type Report struct {
	Meta       Meta         `json:"meta"`
	Summary    Summary      `json:"summary"`
	Nodes      []Node       `json:"nodes"`
	Edges      []Edge       `json:"edges"`
	Violations []Violation  `json:"violations"`
	Files      []FileReport `json:"files"`
}

func (r *Report) Recount() {
	r.Summary.Nodes = len(r.Nodes)
	r.Summary.Edges = len(r.Edges)
	r.Summary.Files = len(r.Files)
	r.Summary.Violations = len(r.Violations)

	r.Summary.Errors = 0
	r.Summary.Warnings = 0
	r.Summary.Infos = 0
	r.Summary.Modules = 0
	r.Summary.Submodules = 0
	r.Summary.Pages = 0
	r.Summary.CommonItems = 0

	for _, violation := range r.Violations {
		switch violation.Severity {
		case SeverityError:
			r.Summary.Errors++
		case SeverityWarning:
			r.Summary.Warnings++
		default:
			r.Summary.Infos++
		}
	}

	for _, node := range r.Nodes {
		switch node.Kind {
		case NodeModule:
			r.Summary.Modules++
		case NodeSubmodule:
			r.Summary.Submodules++
		case NodePage:
			r.Summary.Pages++
		case NodeCommonEntity:
			r.Summary.CommonItems++
		}
	}
}
