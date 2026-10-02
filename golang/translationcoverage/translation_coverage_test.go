// Package translationcoverage reports and verifies localization coverage.
//
// Statically inspects the served language tables in pinglego to verify that:
//   - No translatable string is left untranslated or empty
//   - No non-Latin language contains raw English without native script
//   - No Latin language contains verbatim English copies (outside allowable acronyms/brands)
//   - Every language has complete, 1:1 positional array parity with English
//   - 100% translation coverage is maintained across all 23 languages
//
// In addition, when the Pingle server is running, integration tests verify that
// the API wire format correctly serves all 23 dense, complete arrays.
package translationcoverage

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"

	pingletest "github.com/tahamj/pingletest"
)

const expectedLanguageCount = 23

// ── Language classification ──────────────────────────────────────────────────

var latinScript = map[string]bool{
	"English": true, "French": true, "Spanish": true, "German": true, "Swahili": true,
}

const referenceLang = "English"

// ── Parsing ──────────────────────────────────────────────────────────────────

var stringIDFieldRe = regexp.MustCompile(`StringId:\s*(\d+)`)

type kv struct {
	key   string
	value string
}

func findMatchingBrace(text string, start int) int {
	depth := 1
	inStr, inTick := false, false
	for i := start; i < len(text); i++ {
		c := text[i]
		if inStr {
			if c == '"' && text[i-1] != '\\' {
				inStr = false
			}
			continue
		}
		if inTick {
			if c == '`' {
				inTick = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '`':
			inTick = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func sliceBody(content string) string {
	loc := regexp.MustCompile(`=\s*\[\]\w[\w.]*\s*\{`).FindStringIndex(content)
	if loc == nil {
		return ""
	}
	open := loc[1] - 1
	close := findMatchingBrace(content, open+1)
	if close == -1 {
		return ""
	}
	return content[open+1 : close]
}

func elementStrings(body string) []string {
	var els []string
	n := len(body)
	i := 0
	for i < n {
		for i < n && body[i] != '{' {
			i++
		}
		if i >= n {
			break
		}
		open := i
		close := findMatchingBrace(body, open+1)
		if close == -1 {
			break
		}
		end := close + 1
		for end < n && body[end] != '\n' {
			end++
		}
		els = append(els, body[open:end])
		i = end
	}
	return els
}

func stripTrailingComment(elem string) string {
	inDouble, inRaw, inRune := false, false, false
	depth := 0
	for i := 0; i < len(elem); i++ {
		c := elem[i]
		switch {
		case inRaw:
			if c == '`' {
				inRaw = false
			}
		case inDouble, inRune:
			if c == '\\' {
				i++
				continue
			}
			if inDouble && c == '"' {
				inDouble = false
			} else if inRune && c == '\'' {
				inRune = false
			}
		case c == '"':
			inDouble = true
		case c == '`':
			inRaw = true
		case c == '\'':
			inRune = true
		case c == '{':
			depth++
		case c == '}':
			depth--
		case depth <= 0 && c == '/' && i+1 < len(elem) && elem[i+1] == '/':
			return elem[:i]
		}
	}
	return elem
}

func readStringField(elem, field string) (string, bool) {
	idx := strings.Index(elem, field+":")
	if idx == -1 {
		return "", false
	}
	j := idx + len(field) + 1
	for j < len(elem) && (elem[j] == ' ' || elem[j] == '\t') {
		j++
	}
	if j >= len(elem) {
		return "", false
	}
	switch elem[j] {
	case '`':
		end := strings.IndexByte(elem[j+1:], '`')
		if end == -1 {
			return "", false
		}
		return elem[j+1 : j+1+end], true
	case '"':
		var sb strings.Builder
		for k := j + 1; k < len(elem); k++ {
			c := elem[k]
			if c == '\\' && k+1 < len(elem) {
				sb.WriteByte(elem[k+1])
				k++
				continue
			}
			if c == '"' {
				return sb.String(), true
			}
			sb.WriteByte(c)
		}
	}
	return "", false
}

func parseStringValues(path string) ([]kv, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var out []kv
	for _, el := range elementStrings(sliceBody(string(b))) {
		el = stripTrailingComment(el)
		m := stringIDFieldRe.FindStringSubmatch(el)
		if m == nil {
			continue
		}
		v, _ := readStringField(el, "StringValue")
		out = append(out, kv{key: m[1], value: v})
	}
	return out, nil
}

func parseNotTranslatable(path string) (map[string]bool, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	exempt := make(map[string]bool)
	re := regexp.MustCompile(`(\d+):\s*true`)
	for _, match := range re.FindAllStringSubmatch(string(b), -1) {
		exempt[match[1]] = true
	}
	return exempt, nil
}

// ── Untranslated detection ───────────────────────────────────────────────────

var (
	htmlTagRe  = regexp.MustCompile(`<[^>]+>`)
	urlRe      = regexp.MustCompile(`https?://\S+`)
	emailRe    = regexp.MustCompile(`\S+@\S+\.\S+`)
	fmtTokenRe = regexp.MustCompile(`\{[A-Za-z0-9_]+\}|%[sdvft]`)
	wordRe     = regexp.MustCompile(`[A-Za-z]+`)
)

var brandAllow = map[string]bool{
	"pingle": true, "dns": true, "sla": true, "ip": true, "ipv6": true,
	"pem": true, "ca": true, "tt": true, "rtt": true, "ms": true,
	"lan": true, "led": true, "id": true, "tls": true, "starttls": true,
	"ldap": true, "ldaps": true, "mos": true, "pdf": true, "excel": true,
	"csv": true, "api": true, "url": true, "http": true, "https": true,
	"pop": true, "smtp": true, "imap": true, "ok": true, "status": true,
	"ping": true, "total": true, "original": true, "active": true,
	"cancel": true, "default": true, "filter": true, "service": true,
	"services": true, "protocol": true, "version": true, "date": true,
	"format": true, "menu": true, "help": true, "info": true, "code": true,
	"mode": true, "host": true, "port": true, "agent": true, "probe": true,
	"probes": true, "token": true, "tokens": true, "session": true,
	"sessions": true, "admin": true, "root": true, "client": true,
	"clients": true, "server": true, "servers": true, "site": true,
	"sites": true, "test": true, "tests": true, "reset": true, "clear": true,
	"tag": true, "tags": true, "excellent": true, "fair": true, "bon": true,
	"pause": true, "resume": true, "plan": true, "plans": true, "login": true,
	"logout": true, "enter": true, "tab": true, "shift": true, "alt": true,
	"ctrl": true, "esc": true, "space": true, "del": true, "backspace": true,
	"home": true, "end": true, "page": true,
	// Domain words that are legitimately identical in Latin-script languages
	"diagnostics": true, "jitter": true, "licence": true, "hostname": true,
	"optional": true, "region": true, "notes": true, "roles": true,
	"webhook": true, "description": true, "warm": true, "graphite": true,
	"system": true, "compact": true, "ticket": true, "tickets": true,
	"type": true, "rfc": true,
}

func hasNativeLetter(s string) bool {
	for _, r := range s {
		if r > 127 && unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func looksEnglish(value string) bool {
	s := htmlTagRe.ReplaceAllString(value, " ")
	s = urlRe.ReplaceAllString(s, " ")
	s = emailRe.ReplaceAllString(s, " ")
	s = fmtTokenRe.ReplaceAllString(s, " ")
	for _, w := range wordRe.FindAllString(s, -1) {
		if len(w) < 3 {
			continue
		}
		if w == strings.ToUpper(w) {
			continue
		}
		if brandAllow[strings.ToLower(w)] {
			continue
		}
		return true
	}
	return false
}

func isUntranslated(lang, value, eng string) bool {
	if strings.TrimSpace(value) == "" {
		return false
	}
	if !looksEnglish(value) {
		return false
	}
	if latinScript[lang] {
		return value == eng
	}
	if hasNativeLetter(value) {
		return false
	}
	return looksEnglish(value)
}

func langOf(path string) string {
	name := filepath.Base(path)
	if i := strings.IndexByte(name, '_'); i > 0 {
		return name[:i]
	}
	return name
}

func repoRoot() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	dir := cwd
	for i := 0; i < 10; i++ {
		if isDir(filepath.Join(dir, "pinglego")) && isDir(filepath.Join(dir, "pingledb")) {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("repo root not found from %s", cwd)
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

func rel(root, p string) string {
	if r, err := filepath.Rel(root, p); err == nil {
		return r
	}
	return p
}

func truncate(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) > 60 {
		return s[:60] + "…"
	}
	return s
}

// ── Static Offline Translation Tests ─────────────────────────────────────────

func getLangFiles(t *testing.T) (refPath string, otherPaths []string, exempt map[string]bool, root string) {
	t.Helper()
	var err error
	root, err = repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	langDir := filepath.Join(root, "pinglego", "pkg", "initmicroservice", "initconstants", "langstrings")
	refPath = filepath.Join(langDir, "English_PingleStrings.go")
	if _, err := os.Stat(refPath); err != nil {
		t.Fatalf("missing English reference file: %v", err)
	}

	files, err := os.ReadDir(langDir)
	if err != nil {
		t.Fatalf("read langstrings dir: %v", err)
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), "_PingleStrings.go") {
			continue
		}
		if f.Name() == "English_PingleStrings.go" {
			continue
		}
		otherPaths = append(otherPaths, filepath.Join(langDir, f.Name()))
	}
	sort.Strings(otherPaths)

	notTranslatablePath := filepath.Join(root, "pinglego", "pkg", "initmicroservice", "initconstants", "PingleStringsNotTranslatable.go")
	exempt, err = parseNotTranslatable(notTranslatablePath)
	if err != nil {
		t.Fatalf("parse PingleStringsNotTranslatable: %v", err)
	}

	return refPath, otherPaths, exempt, root
}

func TestNoUntranslatedStrings(t *testing.T) {
	refPath, otherPaths, exempt, root := getLangFiles(t)

	refEls, err := parseStringValues(refPath)
	if err != nil {
		t.Fatalf("parse reference strings: %v", err)
	}
	refByKey := make(map[string]string, len(refEls))
	for _, e := range refEls {
		refByKey[e.key] = e.value
	}

	var failures []string
	totalChecked, totalFlagged := 0, 0

	for _, lp := range otherPaths {
		lang := langOf(lp)
		els, err := parseStringValues(lp)
		if err != nil {
			t.Errorf("parse %s: %v", lp, err)
			continue
		}
		for _, e := range els {
			if exempt[e.key] {
				continue
			}
			totalChecked++
			if isUntranslated(lang, e.value, refByKey[e.key]) {
				totalFlagged++
				failures = append(failures, fmt.Sprintf("%s [id %s]: untranslated %q", rel(root, lp), e.key, truncate(e.value)))
			}
		}
	}

	t.Logf("untranslated check: inspected %d non-exempt values, flagged %d", totalChecked, totalFlagged)
	if len(failures) > 0 {
		sort.Strings(failures)
		t.Errorf("untranslated strings (%d):\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

func TestNoMissingLanguageStrings(t *testing.T) {
	refPath, otherPaths, _, root := getLangFiles(t)

	refEls, err := parseStringValues(refPath)
	if err != nil {
		t.Fatalf("parse reference strings: %v", err)
	}
	refKeys := make(map[string]bool, len(refEls))
	for _, e := range refEls {
		refKeys[e.key] = true
	}

	var failures []string
	for _, lp := range otherPaths {
		els, err := parseStringValues(lp)
		if err != nil {
			t.Errorf("parse %s: %v", lp, err)
			continue
		}

		have := make(map[string]bool, len(els))
		for index, e := range els {
			have[e.key] = true
			if e.key != strconv.Itoa(index) {
				failures = append(failures, fmt.Sprintf("%s: position %d has id %s (misaligned)", rel(root, lp), index, e.key))
			}
		}

		var missing []string
		for k := range refKeys {
			if !have[k] {
				missing = append(missing, k)
			}
		}
		if len(missing) > 0 {
			sort.Strings(missing)
			failures = append(failures, fmt.Sprintf("%s: missing %d ids: %s", rel(root, lp), len(missing), strings.Join(missing, ",")))
		}
		if len(els) != len(refEls) {
			failures = append(failures, fmt.Sprintf("%s: count %d != English %d", rel(root, lp), len(els), len(refEls)))
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Errorf("missing or misaligned language entries (%d):\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

func TestNoEmptyTranslations(t *testing.T) {
	refPath, otherPaths, exempt, root := getLangFiles(t)

	refEls, err := parseStringValues(refPath)
	if err != nil {
		t.Fatalf("parse reference strings: %v", err)
	}
	refByKey := make(map[string]string, len(refEls))
	for _, e := range refEls {
		refByKey[e.key] = e.value
	}

	var failures []string
	for _, lp := range otherPaths {
		els, err := parseStringValues(lp)
		if err != nil {
			continue
		}
		for _, e := range els {
			if exempt[e.key] {
				continue
			}
			if strings.TrimSpace(e.value) == "" && strings.TrimSpace(refByKey[e.key]) != "" {
				failures = append(failures, fmt.Sprintf("%s [id %s]: empty while English is %q", rel(root, lp), e.key, truncate(refByKey[e.key])))
			}
		}
	}

	if len(failures) > 0 {
		sort.Strings(failures)
		t.Errorf("empty translations (%d):\n%s", len(failures), strings.Join(failures, "\n"))
	}
}

func TestCoverageIs100Percent(t *testing.T) {
	refPath, otherPaths, exempt, _ := getLangFiles(t)

	refEls, err := parseStringValues(refPath)
	if err != nil {
		t.Fatalf("parse reference strings: %v", err)
	}

	translatableTotal := 0
	for _, e := range refEls {
		if !exempt[e.key] {
			translatableTotal++
		}
	}

	for _, lp := range otherPaths {
		lang := langOf(lp)
		els, err := parseStringValues(lp)
		if err != nil {
			t.Errorf("parse %s: %v", lp, err)
			continue
		}

		translated := 0
		for _, e := range els {
			if !exempt[e.key] && strings.TrimSpace(e.value) != "" {
				translated++
			}
		}

		percent := translated * 100 / translatableTotal
		if percent != 100 {
			t.Errorf("%s: translation coverage is %d%% (%d/%d translatable strings)", lang, percent, translated, translatableTotal)
		}
	}
}

func TestHelpTranslationsAreCompleteAndPreserveTags(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("locate repo root: %v", err)
	}

	helpDir := filepath.Join(root, "pinglego", "pkg", "initmicroservice", "initconstants", "langhelp")
	enPath := filepath.Join(helpDir, "English_PingleHelp.go")
	enContent, err := os.ReadFile(enPath)
	if err != nil {
		t.Fatalf("read English help: %v", err)
	}

	reHelp := regexp.MustCompile("(?s)HelpText:\\s*`([^`]*)`")
	enMatches := reHelp.FindAllStringSubmatch(string(enContent), -1)
	if len(enMatches) != 15 {
		t.Fatalf("English help has %d docs, want 15", len(enMatches))
	}
	enDocs := make([]string, len(enMatches))
	for i, m := range enMatches {
		enDocs[i] = m[1]
	}

	tags := []string{"<h2>", "<h3>", "<p>", "<ul>", "<li>", "<strong>", "<em>", "<code>", "</h2>", "</h3>", "</p>", "</ul>", "</li>", "</strong>", "</em>", "</code>"}
	tagShape := func(html string) map[string]int {
		res := make(map[string]int)
		for _, tag := range tags {
			res[tag] = strings.Count(html, tag)
		}
		return res
	}

	files, err := os.ReadDir(helpDir)
	if err != nil {
		t.Fatalf("read langhelp dir: %v", err)
	}

	checkedLangs := 0
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), "_PingleHelp.go") || f.Name() == "English_PingleHelp.go" {
			continue
		}
		checkedLangs++
		content, err := os.ReadFile(filepath.Join(helpDir, f.Name()))
		if err != nil {
			t.Errorf("%s: read error: %v", f.Name(), err)
			continue
		}
		matches := reHelp.FindAllStringSubmatch(string(content), -1)
		if len(matches) != 15 {
			t.Errorf("%s: has %d docs, want 15", f.Name(), len(matches))
			continue
		}
		for i, m := range matches {
			doc := m[1]
			enDoc := enDocs[i]
			if strings.TrimSpace(doc) == strings.TrimSpace(enDoc) {
				t.Errorf("%s: doc %d is untranslated English", f.Name(), i)
			}
			sDoc := tagShape(doc)
			sEn := tagShape(enDoc)
			for _, tag := range tags {
				if sDoc[tag] != sEn[tag] {
					t.Errorf("%s: doc %d tag %s count = %d, want %d", f.Name(), i, tag, sDoc[tag], sEn[tag])
				}
			}
		}
	}

	if checkedLangs != 22 {
		t.Errorf("checked %d non-English help files, want 22", checkedLangs)
	}
}

// ── Server Integration Tests (RequireServer) ─────────────────────────────────

func TestAllLanguagesArePresent(t *testing.T) {
	pingletest.RequireServer(t)

	response := pingletest.Call(t, http.MethodGet, "/init/language/list", "", nil)
	if response.Status != http.StatusOK {
		t.Fatalf("listing languages: %d", response.Status)
	}

	languages := pingletest.ListOf(t, response, "languages")
	if len(languages) != expectedLanguageCount {
		t.Errorf("expected %d languages, got %d", expectedLanguageCount, len(languages))
	}

	rightToLeft := 0
	for _, language := range languages {
		if isRtl, _ := language["is_right_to_left"].(bool); isRtl {
			rightToLeft++
		}
		for _, field := range []string{"language_id", "code", "english_name", "native_name"} {
			if value, _ := language[field].(string); value == "" {
				t.Errorf("language %v has no %s", language["language_id"], field)
			}
		}
	}

	if rightToLeft != 4 {
		t.Errorf("expected 4 right-to-left languages, got %d", rightToLeft)
	}
}

func TestEveryLanguageServesACompleteStringArray(t *testing.T) {
	pingletest.RequireServer(t)

	english := pingletest.Call(t, http.MethodGet, "/init?language_id=0", "", nil)
	expected := len(pingletest.ListOf(t, english, "strings"))
	if expected == 0 {
		t.Fatal("the English catalogue is empty")
	}

	for languageId := 0; languageId < expectedLanguageCount; languageId++ {
		response := pingletest.Call(t, http.MethodGet,
			"/init?language_id="+itoa(languageId), "", nil)
		if response.Status != http.StatusOK {
			t.Errorf("language %d: %d", languageId, response.Status)
			continue
		}

		strings := pingletest.ListOf(t, response, "strings")
		if len(strings) != expected {
			t.Errorf("language %d serves %d strings, English has %d: "+
				"positional lookup would crash the client",
				languageId, len(strings), expected)
		}
		for index, entry := range strings {
			if value, _ := entry["string_value"].(string); value == "" {
				t.Errorf("language %d has an empty string at index %d: "+
					"the English fallback did not fill it", languageId, index)
			}
		}
	}
}

func TestCoverageIsReported(t *testing.T) {
	pingletest.RequireServer(t)

	response := pingletest.Call(t, http.MethodGet, "/init/coverage", "", nil)
	if response.Status != http.StatusOK {
		t.Fatalf("reading coverage: %d", response.Status)
	}

	rows := pingletest.ListOf(t, response, "languages")
	sort.Slice(rows, func(i, j int) bool {
		return toInt(rows[i]["coverage_percent"]) > toInt(rows[j]["coverage_percent"])
	})

	t.Logf("%d strings in the catalogue", int(response.Float("string_count")))
	t.Logf("%-12s %-12s %s", "LANGUAGE", "NATIVE", "COVERAGE")

	incomplete := 0
	for _, row := range rows {
		coverage := toInt(row["coverage_percent"])
		if coverage < 100 {
			incomplete++
		}
		marker := ""
		if isRtl, _ := row["is_right_to_left"].(bool); isRtl {
			marker = "  RTL"
		}
		t.Logf("%-12v %-12v %3d%%%s",
			row["english_name"], row["native_name"], coverage, marker)
	}

	if incomplete > 0 {
		t.Errorf("%d language(s) are not fully translated", incomplete)
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := ""
	for value > 0 {
		digits = string(rune('0'+value%10)) + digits
		value /= 10
	}
	return digits
}

func toInt(value any) int {
	if number, ok := value.(float64); ok {
		return int(number)
	}
	return 0
}
