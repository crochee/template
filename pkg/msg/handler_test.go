package msg

import (
	"testing"
)

func TestCallerFunc(t *testing.T) {
	t.Log(op())
}

func op() string {
	return CallerFunc(0)
}
