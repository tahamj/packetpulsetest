package docsparity

import (
	"go/ast"
	"html"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// ── the reflector's /healthz, as the guides print it ─────────────────────

var (
	// healthzSample is a whole /healthz answer as a guide prints it: the
	// reflector's own keys, from the first in order. A release's guide is a
	// template, its port and version {{PLACEHOLDERS}} until the build fills
	// them in.
	healthzSample = regexp.MustCompile(`\{"active_sessions"(?:[^{}]|\{\{[A-Z_]+\}\})*\}`)
	jsonKey       = regexp.MustCompile(`"([a-z_]+)":`)
)

// healthzKeys are the keys the reflector's /healthz writes, read from its
// handler: a Go map, which JSON writes in the keys' order.
func healthzKeys(t *testing.T, root string) []string {
	t.Helper()
	file := parseGo(t, filepath.Join(root, "packetpulsego/cmd/packetpulsereflector/main.go"))
	var keys []string
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}
		if path, ok := call.Args[0].(*ast.BasicLit); !ok || path.Value != `"/healthz"` {
			return true
		}
		ast.Inspect(call.Args[1], func(inner ast.Node) bool {
			literal, ok := inner.(*ast.CompositeLit)
			if !ok {
				return true
			}
			if _, isMap := literal.Type.(*ast.MapType); !isMap {
				return true
			}
			for _, element := range literal.Elts {
				key, _ := element.(*ast.KeyValueExpr).Key.(*ast.BasicLit)
				if key == nil {
					t.Fatalf("a /healthz key that is not a literal: %#v", element)
				}
				name, _ := strconv.Unquote(key.Value)
				keys = append(keys, name)
			}
			return false
		})
		return false
	})
	if len(keys) == 0 {
		t.Fatal("no /healthz answer in the reflector's main.go: the reader no longer finds it")
	}
	sort.Strings(keys)
	return keys
}

// Every /healthz answer a guide prints is the reflector's: the same keys, in
// the order it writes them (19 November 2026). An installer compares what
// their reflector says with the guide's sample, and the reflector of that day
// added speed_media - whether it sends a test's download as a video stream -
// while every guide still printed the sample without it.
func TestEveryHealthzSampleShowsWhatTheReflectorSays(t *testing.T) {
	root := repoRoot(t)
	want := strings.Join(healthzKeys(t, root), ",")
	var files []string
	for _, pattern := range []string{"docs/*.html", "packetpulserelease/reflector/docs/*.html"} {
		matches, err := filepath.Glob(filepath.Join(root, pattern))
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, matches...)
	}
	samples := 0
	for _, path := range files {
		text := html.UnescapeString(htmlTag.ReplaceAllString(read(t, path), ""))
		for _, sample := range healthzSample.FindAllString(text, -1) {
			samples++
			var got []string
			for _, match := range jsonKey.FindAllStringSubmatch(sample, -1) {
				got = append(got, match[1])
			}
			if strings.Join(got, ",") != want {
				relative, _ := filepath.Rel(root, path)
				t.Errorf("%s prints /healthz as\n  %s\nwhose keys are %s; the reflector writes %s", relative, sample, strings.Join(got, ","), want)
			}
		}
	}
	// The installation guides print several: none found is a search that
	// no longer matches them, not guides that agree.
	if samples < 4 {
		t.Fatalf("found %d /healthz samples in the guides, want the installation guides' four or more", samples)
	}
}
