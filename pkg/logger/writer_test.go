package logger_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"template/pkg/logger"
)

func Test_writerManager_GetLogger(t *testing.T) {
	w := logger.NewWriterManager()

	t.Run("success", func(t *testing.T) {
		path := "/tmp/test.log"
		writer1 := w.GetLogger(path, 100, 30, 10, true, true)
		writer2 := w.GetLogger(path, 100, 30, 10, true, true)
		assert.Equal(t, reflect.ValueOf(writer1).Pointer(), reflect.ValueOf(writer2).Pointer())
	})

}
