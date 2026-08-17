package a

import (
	"dep"
	"errs"
)

// ErrCopied retypes a message its own dependency publishes.
const ErrCopied errs.Const = "not a regular file" // want `sentinel ErrCopied duplicates dep.ErrShared`

// ErrAlsoCopied is the same defect in a grouped declaration.
const (
	ErrFine       errs.Const = "something only this package can raise"
	ErrAlsoCopied errs.Const = "file is too large to analyze" // want `sentinel ErrAlsoCopied duplicates dep.ErrOther`
)

// ErrAliased is the correct spelling: one identifier, held together by the
// compiler rather than by a string somebody typed twice.
const ErrAliased = dep.ErrShared

// ErrPrivate copies a message no consumer can see, so there is nothing to alias.
const ErrPrivate errs.Const = "private business"

// NotASentinel is an ordinary string constant that happens to read alike.
const NotASentinel = "not a regular file"

// A const spec with NO VALUE of its own repeats the previous one, so it carries
// the same duplicated text — and there is no fresh literal on its line to be a
// duplicate of anything. The head is reported at its own line; the tail is not
// reported twice for the same string.
const (
	ErrRepeatedHead errs.Const = "not a regular file" // want `sentinel ErrRepeatedHead duplicates dep.ErrShared`
	ErrRepeatedTail
)

// A duplicated sentinel whose value is an EXPRESSION rather than an identifier:
// the root of `errs.Const("not a regular file") + ""` is neither an Ident nor a
// SelectorExpr, so it names nothing that could be an alias and the declaration
// is judged on its own text like any other copy.
const ErrFromExpression = errs.Const("not a regular file") + "" // want `sentinel ErrFromExpression duplicates dep.ErrShared`
