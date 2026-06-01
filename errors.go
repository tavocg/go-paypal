package paypal

import "fmt"

type Error struct {
	status int
	body   []byte
}

func newError(status int, body []byte) *Error {
	return &Error{
		status: status,
		body:   body,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("paypal request failed: status %d: %s", e.status, string(e.body))
}

func (e *Error) Status() int {
	return e.status
}

func (e *Error) Body() []byte {
	return append([]byte(nil), e.body...)
}
