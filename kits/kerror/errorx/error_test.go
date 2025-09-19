package errorx

import (
	"testing"
)

func TestError(t *testing.T) {
	err := ErrInvalidArgument("test")
	t.Logf("%+v", err)
}
