package utils

// Error is a generic error which includes a Sender, the instance that raised
// the error. It's used for passing errors down channels and identifying the sender
type Error struct {
	Error  error
	Sender interface{}
}

func NewError(s interface{}, err error) Error {
	return Error{
		Error:  err,
		Sender: s,
	}
}
