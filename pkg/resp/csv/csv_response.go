package csv

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/afero/mem"

	"template/pkg/utils/v"
)

func NewCsvRender(file *mem.File, r *http.Request) *csvResponse {
	return &csvResponse{
		file:    file,
		request: r,
	}
}

type csvResponse struct {
	file    *mem.File
	request *http.Request
}

func (c *csvResponse) Render(writer http.ResponseWriter) error {
	now := time.Now().Local()
	fileName := url.QueryEscape(fmt.Sprintf("dcs_%s_%s.csv", c.file.Name(), now.Format("2006-01-02")))
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	http.ServeContent(writer, c.request, fileName, c.file.Info().ModTime(), c.file)
	return nil
}

func (c *csvResponse) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
}

func CsvHandler(name string, headers []string, rows [][]interface{}, w io.Writer) error {
	if _, err := w.Write([]byte("\xEF\xBB\xBF")); err != nil { // 写入UTF-8 BOM，防止中文乱码
		return err
	}
	writer := csv.NewWriter(w)
	if err := writer.Write(headers); err != nil {
		return err
	}
	results := make([][]string, len(rows))
	for i, row := range rows {
		temp := make([]string, len(row))
		for j, v := range row {
			temp[j] = toString(v)
		}
		results[i] = temp
	}
	return writer.WriteAll(results)
}

func toString(d interface{}) string {
	switch value := d.(type) {
	case bool:
		if value {
			return v.True
		}
		return v.False
	case *bool:
		if *value {
			return v.True
		}
		return v.False
	case string:
		return value
	case *string:
		return *value
	case int, *int, int64, *int64:
		return fmt.Sprintf("%d", d)
	default:
		j, _ := jsoniter.ConfigCompatibleWithStandardLibrary.MarshalToString(d)
		return j
	}
}
