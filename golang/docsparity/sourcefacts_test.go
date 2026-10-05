package docsparity

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// The facts the guide states, re-derived from source. Each reader here is the
// one place the docs suite learns a fact, so a test comparing the guide with
// it compares two independent things: what was written, and what is true.

// repoRoot walks up from this test until it finds the umbrella repository:
// the directory holding packetpulsego/, packetpulseflutter/ and packetpulsetest/.
func repoRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if isDir(filepath.Join(directory, "packetpulsego")) && isDir(filepath.Join(directory, "packetpulseflutter")) &&
			isDir(filepath.Join(directory, "packetpulsetest")) {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("could not find the repository root above the docs suite")
		}
		directory = parent
	}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func read(t *testing.T, path string) string {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

// route is one registered API route as the guide's Appendix A describes it.
//
// Capability is what the route requires, as capability names: one name, "a|b"
// for either of two, and "a+b" for both - "a|b+c" is either a or b, and c.
type route struct {
	Method, Path, Access, Capability string
	Licensed, Audited                bool
}

func (r route) key() string { return r.Method + " " + r.Path }

// sourceRoutes reads every RegisterRoutes in packetpulsego with go/parser: the
// path from its constant, and from the middleware the group and the route
// add, who may call it, the capability it needs and whether a licence gates
// it. Audited comes from the audit registry, matched as the middleware
// matches - by method and path suffix.
func sourceRoutes(t *testing.T, root string) []route {
	t.Helper()
	consts := map[string]string{}
	var files []string
	for _, pattern := range []string{"packetpulsego/pkg/*/*constants/*.go", "packetpulsego/pkg/*/*app/*.go"} {
		found, _ := filepath.Glob(filepath.Join(root, pattern))
		files = append(files, found...)
	}
	parsed := map[string]*ast.File{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		parsed[path] = file
		for _, decl := range file.Decls {
			gen, ok := decl.(*ast.GenDecl)
			if !ok || gen.Tok != token.CONST {
				continue
			}
			for _, spec := range gen.Specs {
				value := spec.(*ast.ValueSpec)
				for index, name := range value.Names {
					if index >= len(value.Values) {
						continue
					}
					if literal, ok := value.Values[index].(*ast.BasicLit); ok && literal.Kind == token.STRING {
						text, _ := strconv.Unquote(literal.Value)
						consts[file.Name.Name+"."+name.Name] = text
					}
				}
			}
		}
	}

	registry := regexp.MustCompile(`\{Method: "(\w+)", PathSuffix: "([^"]+)"`).FindAllStringSubmatch(
		read(t, filepath.Join(root, "packetpulsego/pkg/auditlogmicroservice/auditlogconstants/AuditLogRegistry.go")), -1)

	var routes []route
	for path, file := range parsed {
		if !strings.Contains(filepath.Dir(path), "app") {
			continue
		}
		for _, decl := range file.Decls {
			function, ok := decl.(*ast.FuncDecl)
			if !ok || function.Name.Name != "RegisterRoutes" {
				continue
			}
			collectRoutes(t, function.Body.List, nil, map[string]string{}, consts, file.Name.Name, &routes)
		}
	}
	for index := range routes {
		for _, entry := range registry {
			if entry[1] == routes[index].Method && strings.HasSuffix(routes[index].Path, entry[2]) {
				routes[index].Audited = true
			}
		}
	}
	sort.Slice(routes, func(i, j int) bool { return routes[i].key() < routes[j].key() })
	return routes
}

func collectRoutes(t *testing.T, statements []ast.Stmt, inherited []string, inheritedVars map[string]string, consts map[string]string, pkg string, routes *[]route) {
	local := append([]string(nil), inherited...)
	// A guard built once and named - readAny := RequireAnyCapability(...) -
	// is read where it is used, as the guard it holds.
	vars := map[string]string{}
	for name, value := range inheritedVars {
		vars[name] = value
	}
	guard := func(argument ast.Expr) string {
		if name, ok := argument.(*ast.Ident); ok && vars[name.Name] != "" {
			return vars[name.Name]
		}
		return render(argument)
	}
	for _, statement := range statements {
		if assign, ok := statement.(*ast.AssignStmt); ok && assign.Tok == token.DEFINE {
			for index, name := range assign.Lhs {
				if identifier, ok := name.(*ast.Ident); ok && index < len(assign.Rhs) {
					vars[identifier.Name] = render(assign.Rhs[index])
				}
			}
			continue
		}
		expression, ok := statement.(*ast.ExprStmt)
		if !ok {
			continue
		}
		call, ok := expression.X.(*ast.CallExpr)
		if !ok {
			continue
		}
		selector, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		switch selector.Sel.Name {
		case "Use":
			for _, argument := range call.Args {
				local = append(local, guard(argument))
			}
		case "Group":
			if literal, ok := call.Args[0].(*ast.FuncLit); ok {
				collectRoutes(t, literal.Body.List, local, vars, consts, pkg, routes)
			}
		case "Get", "Post", "Put", "Patch", "Delete":
			guards := append([]string(nil), local...)
			if with, ok := selector.X.(*ast.CallExpr); ok {
				if withSelector, ok := with.Fun.(*ast.SelectorExpr); ok && withSelector.Sel.Name == "With" {
					for _, argument := range with.Args {
						guards = append(guards, guard(argument))
					}
				}
			}
			path := ""
			switch name := call.Args[0].(type) {
			case *ast.SelectorExpr:
				path = consts[render(name.X)+"."+name.Sel.Name]
			case *ast.Ident:
				path = consts[pkg+"."+name.Name]
			}
			if path == "" {
				t.Errorf("a %s route in package %s names no constant the suite can read", selector.Sel.Name, pkg)
				continue
			}
			found := route{Method: strings.ToUpper(selector.Sel.Name), Path: path, Access: "public"}
			var needs []string
			for _, guard := range guards {
				switch {
				case strings.Contains(guard, "RequireCapability("), strings.Contains(guard, "RequireAnyCapability("):
					arguments := guard[strings.Index(guard, "(")+1 : strings.LastIndex(guard, ")")]
					var either []string
					for _, argument := range strings.Split(arguments, ",") {
						either = append(either, strings.TrimPrefix(strings.TrimSpace(argument), "packetpulseaccess."))
					}
					needs = append(needs, strings.Join(either, "|"))
				case strings.Contains(guard, "RequireSuperUser"):
					found.Access = "superuser"
				case strings.Contains(guard, "RequireApiKey"):
					found.Access = "api key"
				case strings.Contains(guard, "RequireAuth") && found.Access == "public":
					found.Access = "session"
				case strings.Contains(guard, "RequireOrganisationScope") && found.Access == "session":
					found.Access = "member"
				case strings.Contains(guard, "entitlementGate.Require"):
					found.Licensed = true
				}
			}
			found.Capability = strings.Join(needs, "+")
			*routes = append(*routes, found)
		}
	}
}

// render prints an expression as source, enough to recognise a guard.
func render(expression ast.Expr) string {
	switch e := expression.(type) {
	case *ast.SelectorExpr:
		return render(e.X) + "." + e.Sel.Name
	case *ast.Ident:
		return e.Name
	case *ast.CallExpr:
		arguments := make([]string, 0, len(e.Args))
		for _, argument := range e.Args {
			arguments = append(arguments, render(argument))
		}
		return render(e.Fun) + "(" + strings.Join(arguments, ",") + ")"
	case *ast.BasicLit:
		return e.Value
	}
	return "?"
}

// capabilities maps each capability's Go name to its code and description.
func capabilities(t *testing.T, root string) (codes map[string]string, descriptions map[string]string) {
	t.Helper()
	text := read(t, filepath.Join(root, "packetpulsego/pkg/common/packetpulseaccess/PacketPulseAccessCategory.go"))
	codes, descriptions = map[string]string{}, map[string]string{}
	for _, match := range regexp.MustCompile(`(?m)^\s+(\w+)\s*=\s*"([a-z_]+)"`).FindAllStringSubmatch(text, -1) {
		codes[match[1]] = match[2]
	}
	for _, match := range regexp.MustCompile(`(?m)^\s+(\w+):\s+"([^"]+)",`).FindAllStringSubmatch(text, -1) {
		if code, ok := codes[match[1]]; ok {
			descriptions[code] = match[2]
		}
	}
	return codes, descriptions
}

// builtInGrants reads the built-in roles' grants from the migration that
// seeds them: role name -> the capability codes it holds.
func builtInGrants(t *testing.T, root string) map[string]map[string]bool {
	t.Helper()
	text := read(t, filepath.Join(root, "packetpulsego/pkg/common/dbclient/migrations/0002_2026_09_30_tenancy_staff_and_access.sql"))
	grants := map[string]map[string]bool{}
	pattern := regexp.MustCompile(`INSERT INTO role_access \(role_id, ([^)]*)\)\s*SELECT role_id, ([0-9,\s]+)\s*FROM staff_role WHERE organisation_id IS NULL AND role_name = '([^']+)'`)
	for _, match := range pattern.FindAllStringSubmatch(text, -1) {
		columns := strings.Split(strings.Join(strings.Fields(match[1]), ""), ",")
		values := strings.Split(strings.Join(strings.Fields(match[2]), ""), ",")
		held := map[string]bool{}
		for index, column := range columns {
			if index < len(values) && values[index] == "1" {
				held[column] = true
			}
		}
		grants[match[3]] = held
	}
	// Later migrations change a built-in role's grants in place, and what a
	// role holds today is the seed with every change applied in order.
	update := regexp.MustCompile(`(?s)UPDATE role_access ra\s+SET ([^;]*?)\s+FROM staff_role sr\s+WHERE [^;]*?sr\.role_name = '([^']+)'`)
	assignment := regexp.MustCompile(`([a-z_]+)\s*=\s*([01])`)
	for _, name := range migrations(t, root) {
		text := read(t, filepath.Join(root, "packetpulsego/pkg/common/dbclient/migrations", name))
		for _, match := range update.FindAllStringSubmatch(text, -1) {
			if grants[match[2]] == nil {
				t.Errorf("%s changes the grants of %q, which the seed does not create", name, match[2])
				continue
			}
			for _, set := range assignment.FindAllStringSubmatch(match[1], -1) {
				grants[match[2]][set[1]] = set[2] == "1"
			}
		}
	}
	return grants
}

func migrations(t *testing.T, root string) []string {
	t.Helper()
	found, _ := filepath.Glob(filepath.Join(root, "packetpulsego/pkg/common/dbclient/migrations/*.sql"))
	names := make([]string, 0, len(found))
	for _, path := range found {
		names = append(names, filepath.Base(path))
	}
	sort.Strings(names)
	return names
}

// suites reads the suites a plain run executes, and the opt-in ones.
func suites(t *testing.T, root string) (all []string, optIn []string) {
	t.Helper()
	text := read(t, filepath.Join(root, "packetpulsetest.sh"))
	match := regexp.MustCompile(`(?m)^ALL_SUITES="([^"]+)"`).FindStringSubmatch(text)
	if match == nil {
		t.Fatal("packetpulsetest.sh declares no ALL_SUITES")
	}
	all = strings.Fields(match[1])
	known := map[string]bool{}
	for _, name := range all {
		known[name] = true
	}
	for _, match := range regexp.MustCompile(`(?m)^\s+(\w+)\)\s+echo `).FindAllStringSubmatch(text, -1) {
		if !known[match[1]] {
			optIn = append(optIn, match[1])
		}
	}
	return all, optIn
}
