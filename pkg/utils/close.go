// Package utils
package utils

import (
	"io"
	"os"
)

// Close clear ,use defer
func Close(c io.Closer) {
	if err := c.Close(); err != nil {
		_, _ = os.Stderr.WriteString(err.Error())
	}
}
