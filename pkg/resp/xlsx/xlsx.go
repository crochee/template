package xlsx

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/spf13/afero/mem"
	"github.com/xuri/excelize/v2"
)

func NewXslxRender(file *mem.File, r *http.Request) *xlsxResponse {
	return &xlsxResponse{
		file:    file,
		request: r,
	}
}

type xlsxResponse struct {
	file    *mem.File
	request *http.Request
}

func (c *xlsxResponse) Render(writer http.ResponseWriter) error {
	fileName := url.QueryEscape(fmt.Sprintf("%s%d.xls", c.file.Name(), time.Now().Second()))
	writer.Header().Set("Content-Type", "application/vnd.ms-excel; charset=utf-8")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", fileName))
	http.ServeContent(writer, c.request, fileName, c.file.Info().ModTime(), c.file)
	return nil
}

func (c *xlsxResponse) WriteContentType(w http.ResponseWriter) {
}

func index2Chara(i int) string {
	if i >= 24 {
		return "nil"
	}
	return string(rune(65 + i))
}

func id2index(charaID int, i int) string {
	return index2Chara(charaID) + strconv.Itoa(i)
}

func XlsxHandler(sheet string, headers []string, rows [][]interface{}, w io.Writer) error {
	f := excelize.NewFile()
	index := f.NewSheet(sheet)
	f.SetActiveSheet(index)
	f.DeleteSheet("Sheet1")
	for k, v := range headers {
		if err := f.SetCellValue(sheet, id2index(k, 1), v); err != nil {
			return err
		}
	}
	for i, row := range rows {
		for j, v := range row {
			switch value := v.(type) {
			case uint64, int64:
				if err := f.SetCellStr(sheet, id2index(j, i+2), fmt.Sprint(value)); err != nil {
					return err
				}
			default:
				if err := f.SetCellValue(sheet, id2index(j, i+2), value); err != nil {
					return err
				}
			}
		}
	}
	return f.Write(w)
}
