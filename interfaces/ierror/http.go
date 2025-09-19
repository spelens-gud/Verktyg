package ierror

import (
	"google.golang.org/genproto/googleapis/rpc/code"
)

// from "cloud.google.com/go/internal/trace"
func FromHttpStatusCode(httpStatusCode int) Code {
	switch httpStatusCode {
	case 200:
		return Code(code.Code_OK)
	case 499:
		return Code(code.Code_CANCELLED)
	case 500:
		return Code(code.Code_UNKNOWN) // Could also be Code_INTERNAL, Code_DATA_LOSS
	case 400:
		return Code(code.Code_INVALID_ARGUMENT) // Could also be Code_OUT_OF_RANGE
	case 504:
		return Code(code.Code_DEADLINE_EXCEEDED)
	case 404:
		return Code(code.Code_NOT_FOUND)
	case 409:
		return Code(code.Code_ALREADY_EXISTS) // Could also be Code_ABORTED
	case 403:
		return Code(code.Code_PERMISSION_DENIED)
	case 401:
		return Code(code.Code_UNAUTHENTICATED)
	case 429:
		return Code(code.Code_RESOURCE_EXHAUSTED)
	case 501:
		return Code(code.Code_UNIMPLEMENTED)
	case 503:
		return Code(code.Code_UNAVAILABLE)
	default:
		return Code(code.Code_UNKNOWN)
	}
}
