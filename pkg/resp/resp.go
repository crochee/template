package resp

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/spf13/afero/mem"

	"template/pkg/code"
	"template/pkg/ptf"
	"template/pkg/resp/csv"
	"template/pkg/resp/xlsx"
)

// Success provides response logic for traditional DCS
func Success(c *gin.Context, data ...interface{}) {
	if len(data) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	accept := c.Request.Header.Get("Accept")
	switch accept {
	case "application/vnd.ms-excel", "text/csv":
		SuccessWithFile(c, data...)
	case "*/*", "application/json", "":
		c.JSON(http.StatusOK, WrapResult(data...))
	}
}

// SuccessWithFile provide response logic for file content
func SuccessWithFile(c *gin.Context, data ...interface{}) {
	if len(data) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	if len(data) < 2 {
		Error(c, code.ErrInternalServerError)
		return
	}

	opts := make([]ptf.Option, 0, 4)
	if len(data) > 2 {
		fieldNames := make([]string, 0)
		for _, value := range data[2:] {
			switch temp := value.(type) {
			case string:
				fieldNames = append(fieldNames, temp)
			case ptf.Option:
				opts = append(opts, temp)
			default:
			}
		}
		if len(fieldNames) > 0 {
			opts = append(opts, ptf.FieldNames(fieldNames))
		}
	}
	file := mem.NewFileHandle(mem.CreateFile(fmt.Sprint(data[1])))
	opts = append(opts, ptf.Writer(file))

	var r render.Render

	accept := c.Request.Header.Get("Accept")
	switch accept {
	case "application/vnd.ms-excel":
		opts = append(opts, ptf.SetHandler(xlsx.XlsxHandler), ptf.Sheet(file.Name()))
		r = xlsx.NewXslxRender(file, c.Request)
	default:
		opts = append(opts, ptf.SetHandler(csv.CsvHandler))
		r = csv.NewCsvRender(file, c.Request)
	}
	err := ptf.NewMarshal(opts...).Encode(data[0])
	if err != nil {
		Error(c, err)
		return
	}
	_ = file.Sync()
	c.Render(http.StatusOK, r)
}

// WrapResult 包裹DCS返回结果
func WrapResult(result ...interface{}) interface{} {
	response := map[string]interface{}{
		"code":    "200",
		"message": "请求成功",
	}
	if len(result) > 0 {
		response["result"] = result[0]
	}
	return response
}
