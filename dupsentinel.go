// Package dupsentinel provides a go/analysis analyzer that forbids declaring an
// error sentinel a package already imports.
//
// The fleet's error mechanism is a constant: `const ErrThing errs.Const = "…"`,
// matched with errors.Is. Two constants carrying the SAME text are matched by
// errors.Is as equal, because errs.Const compares by value — so a package that
// re-declares a sentinel its dependency already publishes appears to work, and
// the two are coupled by nothing but a string somebody typed twice. Reword
// either message and every consumer's errors.Is silently starts answering false.
//
// The fix is an alias, not a copy: `const ErrThing = dep.ErrThing`. That is the
// same identifier, so the coupling is the compiler's rather than a coincidence.
package dupsentinel

import (
	"go/ast"
	"go/token"
	"go/types"

	goyze "github.com/gomatic/go-yze"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const message = "sentinel %s duplicates %s, which this package already imports; " +
	"two constants with one message are matched by errors.Is only because somebody typed the same string " +
	"twice — write `const %s = %s` so the compiler holds them together"

// isSentinelType reports the type an error constant is declared with: a named
// string type that IS an error.
//
// Asked structurally rather than by import path, for two reasons. A rule keyed
// to one module's `errs.Const` would be wrong for every other user of a public
// tool — the value belongs to whoever runs the check, not to the tool. And the
// defect is not about which package the mechanism came from: two constants of
// ANY such type carrying one message are matched by errors.Is because they
// compare by value, and that is true of a locally declared type exactly as it is
// of a shared one.
func isSentinelType(typed types.Type) bool {
	named, isNamed := typed.(*types.Named)
	if !isNamed {
		return false
	}
	basic, isBasic := named.Underlying().(*types.Basic)
	return isBasic && basic.Kind() == types.String && types.Implements(named, errorInterface)
}

// errorInterface is the built-in error type, looked up once.
var errorInterface = types.Universe.Lookup("error").Type().Underlying().(*types.Interface)

// Analyzer reports a sentinel constant whose text an imported package already
// publishes.
var Analyzer = &analysis.Analyzer{
	Name:     "dupsentinel",
	Doc:      "reports an error sentinel that duplicates one the package already imports",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

// Registration declares this analyzer to the yze framework.
var Registration = goyze.Registration{
	Name:       "dupsentinel",
	Categories: []goyze.Category{"errors"},
	URL:        "https://gomatic.github.io/docs.yze/",
	Analyzer:   Analyzer,
}

// run reports each declared sentinel an import already carries.
func run(pass *analysis.Pass) (any, error) {
	imported := importedSentinels(pass.Pkg)
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)
	insp.Preorder([]ast.Node{(*ast.GenDecl)(nil)}, func(node ast.Node) {
		decl := node.(*ast.GenDecl)
		if decl.Tok != token.CONST {
			return
		}
		for _, spec := range decl.Specs {
			reportDuplicates(pass, imported, spec.(*ast.ValueSpec))
		}
	})
	return nil, nil
}

// reportDuplicates reports each name in one const spec that repeats an imported
// sentinel's text.
func reportDuplicates(pass *analysis.Pass, imported map[string]string, spec *ast.ValueSpec) {
	for at, name := range spec.Names {
		declared, isSentinel := sentinelValue(pass.TypesInfo.ObjectOf(name))
		already, isDuplicate := imported[declared]
		switch {
		case !isSentinel || !isDuplicate:
		case isAlias(pass, spec, at):
			// The alias IS the fix. It has the same type and the same value as
			// the sentinel it names, because it is that sentinel — reporting it
			// would tell an author to replace a line with itself.
		default:
			pass.Reportf(name.Pos(), message, name.Name, already, name.Name, already)
		}
	}
}

// isAlias reports a declaration whose value NAMES another package's constant
// rather than restating its text.
//
// A spec with no values of its own repeats the previous one, which in a constant
// block is either an alias already reported at its own line or an iota form that
// carries no message at all; either way there is no fresh literal here to be a
// duplicate of anything.
func isAlias(pass *analysis.Pass, spec *ast.ValueSpec, at int) bool {
	if at >= len(spec.Values) {
		return true
	}
	named, isName := pass.TypesInfo.Uses[rootIdent(spec.Values[at])]
	return isName && named.Pkg() != nil && named.Pkg() != pass.Pkg
}

// rootIdent is the identifier a value expression names, if it names one: `x` for
// `x`, and `x` for `pkg.x`.
func rootIdent(value ast.Expr) *ast.Ident {
	switch typed := value.(type) {
	case *ast.Ident:
		return typed
	case *ast.SelectorExpr:
		return typed.Sel
	}
	return nil
}

// importedSentinels is every sentinel the package's imports publish, keyed by
// the text that decides what errors.Is matches.
//
// Only DIRECT imports are read. A sentinel two hops away is not one this package
// can alias without taking a dependency it does not have, so reporting it would
// name a fix the author cannot apply.
func importedSentinels(pkg *types.Package) map[string]string {
	found := map[string]string{}
	for _, imported := range pkg.Imports() {
		scope := imported.Scope()
		for _, name := range scope.Names() {
			object := scope.Lookup(name)
			if !object.Exported() {
				continue
			}
			if value, isSentinel := sentinelValue(object); isSentinel {
				found[value] = imported.Name() + "." + name
			}
		}
	}
	return found
}

// sentinelValue is the text of an error-constant declaration, reporting whether
// the object is one at all.
func sentinelValue(object types.Object) (string, bool) {
	constant, isConst := object.(*types.Const)
	if !isConst || !isSentinelType(constant.Type()) {
		return "", false
	}
	return constant.Val().ExactString(), true
}
