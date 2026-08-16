package dep

import "errs"

// ErrShared is the sentinel a consumer should alias rather than retype.
const ErrShared errs.Const = "not a regular file"

// ErrOther exists so a consumer copying a DIFFERENT message is still reported.
const ErrOther errs.Const = "file is too large to analyze"

// unexported is never visible to a consumer, so copying its text is not a
// duplicate of anything they could have aliased.
const unexported errs.Const = "private business"
