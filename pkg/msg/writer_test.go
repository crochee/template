// Package msg
package msg

import (
	"testing"

	"github.com/json-iterator/go"
)

func TestNull(t *testing.T) {
	data, err := jsoniter.ConfigCompatibleWithStandardLibrary.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s", data)
}
