package feod

import (
	"sort"

	"github.com/feod-architecture/feod-analyzer/internal/report"
)

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
