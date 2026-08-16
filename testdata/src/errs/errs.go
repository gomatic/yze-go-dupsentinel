package errs

// Const is the fleet's error-constant type, reproduced here so the fixture needs
// no module outside testdata.
type Const string

func (c Const) Error() string { return string(c) }
