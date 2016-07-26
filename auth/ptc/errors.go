package ptc

import "fmt"

// LoginError - PTC login error handler when it is thrown
type LoginError struct {
	message string
}

func (e *LoginError) Error() string {
	return fmt.Sprintf("auth/ptc error: %s", e.message)
}

func loginError(message string) (string, error) {
	return "", &LoginError{message}
}
