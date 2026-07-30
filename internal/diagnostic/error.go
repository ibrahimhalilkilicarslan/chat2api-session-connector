package diagnostic

import "errors"

const fallbackCode = "CONNECTOR_ERROR"

type Error struct {
	code    string
	message string
	hint    string
}

func New(code string, message string, hint string) error {
	return &Error{code: code, message: message, hint: hint}
}

func (err *Error) Error() string {
	return err.message
}

func (err *Error) DiagnosticCode() string {
	return err.code
}

func (err *Error) DiagnosticHint() string {
	return err.hint
}

type publicError interface {
	error
	DiagnosticCode() string
	DiagnosticHint() string
}

func Details(err error) (string, string) {
	var candidate publicError
	if errors.As(err, &candidate) {
		return candidate.DiagnosticCode(), candidate.DiagnosticHint()
	}
	return fallbackCode, "Connector'ı kapatıp yeniden açın. Sorun sürerse uygulamadaki tanı kodunu paylaşın."
}
