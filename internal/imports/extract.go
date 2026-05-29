package imports

import (
	"regexp"
	"sort"
	"strings"
)

type Statement struct {
	Path     string
	Kind     string
	TypeOnly bool
	Line     int
	Offset   int
}

type pattern struct {
	kind      string
	re        *regexp.Regexp
	pathIndex int
	typeIndex int
	typeOnly  bool
}

var patterns = []pattern{
	{
		kind:      "import",
		re:        regexp.MustCompile(`(?ms)\bimport\s+(type\s+)?(?:[^'";]|\n)*?\s+from\s*["']([^"']+)["']`),
		pathIndex: 2,
		typeIndex: 1,
	},
	{
		kind:      "side-effect-import",
		re:        regexp.MustCompile(`(?m)\bimport\s*["']([^"']+)["']`),
		pathIndex: 1,
	},
	{
		kind:      "dynamic-import",
		re:        regexp.MustCompile(`(?m)\bimport\s*\(\s*["']([^"']+)["']\s*\)`),
		pathIndex: 1,
	},
	{
		kind:      "require",
		re:        regexp.MustCompile(`(?m)\brequire\s*\(\s*["']([^"']+)["']\s*\)`),
		pathIndex: 1,
	},
	{
		kind:      "export",
		re:        regexp.MustCompile(`(?ms)\bexport\s+(type\s+)?(?:[^'";]|\n)*?\s+from\s*["']([^"']+)["']`),
		pathIndex: 2,
		typeIndex: 1,
	},
}

func Extract(source []byte) []Statement {
	text := string(source)
	ignored := ignoredOffsets(text)
	statements := []Statement{}
	seen := map[string]bool{}

	for _, p := range patterns {
		matches := p.re.FindAllStringSubmatchIndex(text, -1)
		for _, match := range matches {
			if ignored[match[0]] {
				continue
			}
			pathStart := match[p.pathIndex*2]
			pathEnd := match[p.pathIndex*2+1]
			if pathStart < 0 || pathEnd < 0 {
				continue
			}

			typeOnly := p.typeOnly
			if p.typeIndex > 0 {
				typeStart := match[p.typeIndex*2]
				typeEnd := match[p.typeIndex*2+1]
				typeOnly = typeStart >= 0 && typeEnd > typeStart
			}

			importPath := text[pathStart:pathEnd]
			key := p.kind + ":" + importPath + ":" + string(rune(match[0]))
			if seen[key] {
				continue
			}
			seen[key] = true

			statements = append(statements, Statement{
				Path:     importPath,
				Kind:     p.kind,
				TypeOnly: typeOnly,
				Line:     lineAtOffset(text, match[0]),
				Offset:   match[0],
			})
		}
	}

	sort.Slice(statements, func(i, j int) bool {
		return statements[i].Offset < statements[j].Offset
	})

	return statements
}

func ExtractStarExports(source []byte) []Statement {
	text := string(source)
	ignored := ignoredOffsets(text)
	re := regexp.MustCompile(`(?ms)\bexport\s+(type\s+)?\*\s+from\s*["']([^"']+)["']`)
	matches := re.FindAllStringSubmatchIndex(text, -1)
	result := []Statement{}

	for _, match := range matches {
		if ignored[match[0]] {
			continue
		}
		pathStart := match[4]
		pathEnd := match[5]
		if pathStart < 0 || pathEnd < 0 {
			continue
		}
		typeOnly := match[2] >= 0 && match[3] > match[2]
		result = append(result, Statement{
			Path:     text[pathStart:pathEnd],
			Kind:     "star-export",
			TypeOnly: typeOnly,
			Line:     lineAtOffset(text, match[0]),
			Offset:   match[0],
		})
	}

	return result
}

func ExtractNamedExports(source []byte) []Statement {
	text := string(source)
	ignored := ignoredOffsets(text)
	re := regexp.MustCompile(`(?ms)\bexport\s+(type\s+)?(?:\{[^}]*\}|[A-Za-z0-9_$]+\s*\{[^}]*\})\s+from\s*["']([^"']+)["']`)
	matches := re.FindAllStringSubmatchIndex(text, -1)
	result := []Statement{}

	for _, match := range matches {
		if ignored[match[0]] {
			continue
		}
		pathStart := match[4]
		pathEnd := match[5]
		if pathStart < 0 || pathEnd < 0 {
			continue
		}
		typeOnly := match[2] >= 0 && match[3] > match[2]
		result = append(result, Statement{
			Path:     text[pathStart:pathEnd],
			Kind:     "named-export",
			TypeOnly: typeOnly,
			Line:     lineAtOffset(text, match[0]),
			Offset:   match[0],
		})
	}

	return result
}

func lineAtOffset(text string, offset int) int {
	if offset <= 0 {
		return 1
	}
	return strings.Count(text[:offset], "\n") + 1
}

func ignoredOffsets(text string) []bool {
	ignored := make([]bool, len(text))
	for i := 0; i < len(text); {
		ch := text[i]
		next := byte(0)
		if i+1 < len(text) {
			next = text[i+1]
		}

		if ch == '/' && next == '/' {
			for i < len(text) && text[i] != '\n' {
				ignored[i] = true
				i++
			}
			continue
		}

		if ch == '/' && next == '*' {
			ignored[i] = true
			ignored[i+1] = true
			i += 2
			for i < len(text) {
				ignored[i] = true
				if i+1 < len(text) && text[i] == '*' && text[i+1] == '/' {
					ignored[i+1] = true
					i += 2
					break
				}
				i++
			}
			continue
		}

		if ch == '"' || ch == '\'' || ch == '`' {
			quote := ch
			escaped := false
			ignored[i] = true
			i++
			for i < len(text) {
				ignored[i] = true
				if escaped {
					escaped = false
					i++
					continue
				}
				if text[i] == '\\' {
					escaped = true
					i++
					continue
				}
				if text[i] == quote {
					i++
					break
				}
				i++
			}
			continue
		}

		i++
	}
	return ignored
}
