package argocd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"
)

// The wait loop cannot be driven from a unit test — it spends 30s in the
// bootstrap phase before the main loop starts. But the defect this guards is
// structural, not behavioural: repo-server recovery, a CORRECTIVE action, was
// nested inside `if config.Verbose`, so a user who did not ask for extra
// logging silently lost the recovery and burned the entire 60-minute timeout
// against a wedged repo-server. The same block hid the spinner's progress text.
//
// This test parses wait.go and asserts that neither call sits under a
// verbosity check. It fails if someone re-nests them.

// verboseGuardedCalls returns the names of the given functions that appear
// somewhere inside an `if config.Verbose ...` statement in file.
func verboseGuardedCalls(t *testing.T, file string, watch map[string]bool) []string {
	t.Helper()

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, file, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", file, err)
	}

	var found []string
	ast.Inspect(f, func(n ast.Node) bool {
		ifStmt, ok := n.(*ast.IfStmt)
		if !ok || !mentionsVerbose(ifStmt.Cond) {
			return true
		}
		// Walk only the `then` branch: an `else` of a verbose check is the
		// non-verbose path, which is exactly where these calls belong.
		ast.Inspect(ifStmt.Body, func(inner ast.Node) bool {
			call, ok := inner.(*ast.CallExpr)
			if !ok {
				return true
			}
			if name := calleeName(call); watch[name] {
				found = append(found, name)
			}
			return true
		})
		return true
	})
	return found
}

// mentionsVerbose reports whether expr references config.Verbose.
func mentionsVerbose(expr ast.Expr) bool {
	var hit bool
	ast.Inspect(expr, func(n ast.Node) bool {
		if sel, ok := n.(*ast.SelectorExpr); ok && sel.Sel.Name == "Verbose" {
			hit = true
		}
		return !hit
	})
	return hit
}

// calleeName returns the bare function or method name of a call expression.
func calleeName(call *ast.CallExpr) string {
	switch fn := call.Fun.(type) {
	case *ast.Ident:
		return fn.Name
	case *ast.SelectorExpr:
		return fn.Sel.Name
	}
	return ""
}

func TestRecoveryAndProgressAreNotGatedOnVerbose(t *testing.T) {
	watch := map[string]bool{
		"triggerRepoServerRecovery": true, // corrective action
		"checkRepoServerHealth":     true, // feeds the corrective action
		"UpdateText":                true, // spinner progress
	}

	guarded := verboseGuardedCalls(t, "wait.go", watch)

	if len(guarded) > 0 {
		t.Errorf("these calls are nested under `if config.Verbose` in wait.go and "+
			"therefore never run for a default invocation: %s\n"+
			"Recovery must not depend on a logging flag; progress is the default UX.",
			strings.Join(guarded, ", "))
	}
}
