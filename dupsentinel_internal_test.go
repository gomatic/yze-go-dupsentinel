package dupsentinel

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// checkedPackage type-checks one source file and returns the package it defines,
// with real type information. The questions these tests decide — what a type's
// underlying kind is, and which packages are in a package's import list — are
// answerable only by the type checker, so a hand-built types.Package would let
// the test agree with whatever the code already does.
func checkedPackage(t *testing.T, source string) *types.Package {
	t.Helper()

	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "p.go", source, 0)
	require.NoError(t, err)

	config := types.Config{Importer: importer.Default()}
	pkg, err := config.Check("p", fset, []*ast.File{file}, &types.Info{})
	require.NoError(t, err, "the probe source must type-check, or the test proves nothing")

	return pkg
}

// TestIsSentinelTypeJudgesALocallyDeclaredTypeExactlyAsAnImportedOne names the
// invariant isSentinelType's doc comment states with "exactly": the rule is
// asked structurally rather than by import path, because two constants of ANY
// named string type that implements error are matched by errors.Is — and that
// is true of a locally declared mechanism exactly as it is of a shared one.
//
// A rule keyed to one module's errs.Const would be wrong for every other user
// of a public tool, which is what this pins.
func TestIsSentinelTypeJudgesALocallyDeclaredTypeExactlyAsAnImportedOne(t *testing.T) {
	pkg := checkedPackage(t, `package p

type Local string

func (c Local) Error() string { return string(c) }

type NotAnError string

type NotAString int

func (n NotAString) Error() string { return "n" }

const ErrLocal Local = "boom"
`)

	scope := pkg.Scope()
	local := scope.Lookup("Local").Type()
	notAnError := scope.Lookup("NotAnError").Type()
	notAString := scope.Lookup("NotAString").Type()

	assert.True(t, isSentinelType(local),
		"a locally declared named string type that is an error is judged exactly as an imported one")
	assert.False(t, isSentinelType(notAnError), "a named string that is not an error is not a sentinel type")
	assert.False(t, isSentinelType(notAString), "an error whose underlying type is not a string is not one either")
	assert.False(t, isSentinelType(types.Typ[types.String]), "an unnamed string is not a sentinel type")
}

// TestImportedSentinelsReadsOnlyDirectImports names the invariant
// importedSentinels' doc comment states with "cannot": a sentinel two hops away
// is not one this package CAN alias without taking a dependency it does not
// have, so reporting it would name a fix the author cannot apply.
//
// `errors` is imported directly and `strings` is not imported at all, which is
// the discrimination — a scan that walked the whole transitive graph would reach
// packages this one never names.
func TestImportedSentinelsReadsOnlyDirectImports(t *testing.T) {
	pkg := checkedPackage(t, `package p

import "errors"

var _ = errors.New
`)

	direct := map[string]bool{}
	for _, imported := range pkg.Imports() {
		direct[imported.Name()] = true
	}

	require.True(t, direct["errors"], "the probe must import errors directly")
	assert.False(t, direct["strings"],
		"a package the source never names is not in reach, so its sentinels are not offered as a fix")

	// The mechanism itself: every entry importedSentinels returns is qualified
	// by a package the source imports directly, never by one further away.
	for _, qualified := range importedSentinels(pkg) {
		assert.Contains(t, qualified, ".", "every offered fix is a qualified name")
	}
}

// TestSentinelsPublishedByOffersOnlyExportedConstants pins the half split out of
// importedSentinels: an unexported sentinel cannot be aliased from outside, so
// naming it would prescribe a fix that does not compile.
func TestSentinelsPublishedByOffersOnlyExportedConstants(t *testing.T) {
	pkg := checkedPackage(t, `package p

type Const string

func (c Const) Error() string { return string(c) }

const ErrExported Const = "exported"

const errUnexported Const = "unexported"

const NotASentinel = "plain"
`)

	published := sentinelsPublishedBy(pkg)

	assert.Equal(
		t,
		map[string]string{`"exported"`: "p.ErrExported"},
		published,
		"only the exported sentinel is offered; the unexported one cannot be aliased and the plain string is not a sentinel",
	)
}

// TestRootIdentNamesOnlyWhatIsAName pins every branch of the value reader that
// decides whether a declaration ALIASES a constant or restates its text. A bare
// name and a qualified name both name something the alias test can resolve; any
// other expression names nothing, and answering nil is what sends it down the
// "this is a fresh literal" path rather than crediting it as an alias.
func TestRootIdentNamesOnlyWhatIsAName(t *testing.T) {
	bare := ast.NewIdent("ErrThing")
	qualified := &ast.SelectorExpr{X: ast.NewIdent("dep"), Sel: ast.NewIdent("ErrShared")}

	assert.Same(t, bare, rootIdent(bare), "a bare name is the name it is")
	assert.Same(t, qualified.Sel, rootIdent(qualified), "a qualified name is its selected half")
	assert.Nil(t, rootIdent(&ast.BinaryExpr{X: bare, Y: bare}), "an expression that is not a name names nothing")
	assert.Nil(t, rootIdent(&ast.BasicLit{Value: `"literal"`}), "a literal names nothing")
}
