package ierror

import "net/http"

type CodeError interface {
	Code() Code
	Unwrap() error
	error
}

type HttpStatusError interface {
	error
	HttpStatus() int
}

func HttpCodeMap(code Code) (status int) {
	status, ok := codeMap[code]
	if ok {
		return
	}
	status = http.StatusInternalServerError
	return
}

type Code int

func (code Code) String() string {
	return CodeText(code)
}

func (code Code) HttpStatus() int {
	return HttpCodeMap(code)
}

func CodeText(code Code) string {
	return statusText[code]
}

var statusText = map[Code]string{
	Cancelled:          "Cancelled",
	Unknown:            "Unknown",
	InvalidArgument:    "Invalid Argument",
	DeadlineExceeded:   "Deadline Exceeded",
	NotFound:           "Not Found",
	AlreadyExists:      "Already Exists",
	PermissionDenied:   "Permission Denied",
	Unauthenticated:    "Unauthenticated",
	ResourceExhausted:  "Resource Exhausted",
	FailedPrecondition: "Failed Precondition",
	Aborted:            "Aborted",
	OutOfRange:         "Out Of Range",
	Unimplemented:      "Unimplemented",
	Internal:           "Internal",
	Unavailable:        "Unavailable",
	DataLoss:           "Data Loss",
}

var codeMap = map[Code]int{
	Cancelled:          499, // Client Closed Request
	Unknown:            http.StatusInternalServerError,
	InvalidArgument:    http.StatusBadRequest,
	DeadlineExceeded:   http.StatusGatewayTimeout,
	NotFound:           http.StatusNotFound,
	AlreadyExists:      http.StatusConflict,
	PermissionDenied:   http.StatusForbidden,
	Unauthenticated:    http.StatusUnauthorized,
	ResourceExhausted:  http.StatusTooManyRequests,
	FailedPrecondition: http.StatusBadRequest,
	Aborted:            http.StatusConflict,
	OutOfRange:         http.StatusBadRequest,
	Unimplemented:      http.StatusNotImplemented,
	Internal:           http.StatusInternalServerError,
	Unavailable:        http.StatusServiceUnavailable,
	DataLoss:           http.StatusInternalServerError,
}

/*
错误类型模板
func ErrCancelled(msg string) (err error)          { return }
func ErrUnknown(msg string) (err error)            { return }
func ErrInvalidArgument(msg string) (err error)    { return }
func ErrDeadlineExceeded(msg string) (err error)   { return }
func ErrNotFound(msg string) (err error)           { return }
func ErrAlreadyExists(msg string) (err error)      { return }
func ErrPermissionDenied(msg string) (err error)   { return }
func ErrUnauthenticated(msg string) (err error)    { return }
func ErrResourceExhausted(msg string) (err error)  { return }
func ErrFailedPrecondition(msg string) (err error) { return }
func ErrAborted(msg string) (err error)            { return }
func ErrOutOfRange(msg string) (err error)         { return }
func ErrUnimplemented(msg string) (err error)      { return }
func ErrInternal(msg string) (err error)           { return }
func ErrUnavailable(msg string) (err error)        { return }
func ErrDataLoss(msg string) (err error)           { return }
*/
