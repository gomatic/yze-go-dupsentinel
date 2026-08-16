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
