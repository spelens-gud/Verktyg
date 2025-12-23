package errorx

import (
	"fmt"
	"io"

	"github.com/spelens-gud/Verktyg/interfaces/ierror"
	"github.com/spelens-gud/Verktyg/kits/kerror/internal/errors"
)

type Error struct {
	error
	code ierror.Code
}

func New(msg string, code ierror.Code) (err error) {
	return Error{error: errors.New(msg), code: code}
}

func Wrap(err error, code ierror.Code) error {
	if err == nil {
		return nil
	}
	return Error{
		error: err,
		code:  code,
	}
}

func Errorf(code ierror.Code, msg string, args ...interface{}) (err error) {
	if code == 0 {
		code = ierror.Unknown
	}
	return Error{error: errors.Errorf(msg, args...), code: code}
}

func (e Error) Code() ierror.Code {
	return e.code
}

func (e Error) Unwrap() error {
	return e.error
}

func (e Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			_, _ = io.WriteString(s, fmt.Sprintf(e.code.String()+": %+v", e.error))
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.code.String()+": "+e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.code.String()+": "+e.Error())
	}
}

func ErrCancelled(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Cancelled, msg, args...)
}
func ErrUnknown(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Unknown, msg, args...)
}
func ErrInvalidArgument(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.InvalidArgument, msg, args...)
}
func ErrDeadlineExceeded(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.DeadlineExceeded, msg, args...)
}
func ErrNotFound(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.NotFound, msg, args...)
}
func ErrAlreadyExists(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.AlreadyExists, msg, args...)
}
func ErrPermissionDenied(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.PermissionDenied, msg, args...)
}
func ErrUnauthenticated(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Unauthenticated, msg, args...)
}
func ErrResourceExhausted(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.ResourceExhausted, msg, args...)
}
func ErrFailedPrecondition(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.FailedPrecondition, msg, args...)
}
func ErrAborted(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Aborted, msg, args...)
}
func ErrOutOfRange(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.OutOfRange, msg, args...)
}
func ErrUnimplemented(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Unimplemented, msg, args...)
}
func ErrInternal(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Internal, msg, args...)
}
func ErrUnavailable(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.Unavailable, msg, args...)
}
func ErrDataLoss(msg string, args ...interface{}) (err error) {
	return Errorf(ierror.DataLoss, msg, args...)
}
