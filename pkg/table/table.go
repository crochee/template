package table

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"template/pkg/json"
)

func RenderAsTable(i interface{}, fields []string) {
	switch x := i.(type) {
	case map[string]interface{}:
		renderShowTable(x, fields)
	case []map[string]interface{}:
		renderListTable(x, fields)
	default:
		fmt.Println("Unable to print data as table, please check and try later.")
	}
}

const (
	DefaultTransverseStringLength = 64
	DefaultPortraitStringLength   = 128
)

// renderListTable 打印list操作数据为ASCII表格， fields用于控制需要打印的列及列从左到右的显示顺序
func renderListTable(data []map[string]interface{}, fields []string) {
	header := make(table.Row, len(fields))
	rows := make([]table.Row, len(data))
	for i, d := range data {
		row := make(table.Row, len(fields))
		for k, v := range d {
			index := indexOf(fields, k)
			if index == -1 {
				continue
			}
			header[index] = k
			row[index] = text.WrapHard(toString(v), DefaultTransverseStringLength)
		}
		rows[i] = row
	}

	// If data is empty, no header will get, we use fields as header
	if len(data) == 0 {
		for i, f := range fields {
			header[i] = f
		}
	}
	render(header, rows)
}

// renderShowTable 打印show操作数据为ASCII表格， field用于控制需要打印的字段及字段从上到下出现的顺序
func renderShowTable(data map[string]interface{}, fields []string) {
	header := table.Row{"Field", "Value"}
	rows := make([]table.Row, len(data))
	for k, v := range data {
		index := indexOf(fields, k)
		if index == -1 {
			continue
		}
		rows[index] = table.Row{k, text.WrapHard(toString(v), DefaultPortraitStringLength)}
	}
	render(header, rows)
}

func render(header table.Row, rows []table.Row) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(header)
	t.AppendRows(rows)
	t.Render()
}

func indexOf(list []string, target string) int {
	for i, v := range list {
		if v == target {
			return i
		}
	}
	return -1
}

func toString(d interface{}) string {
	switch d.(type) {
	case bool, *bool:
		return fmt.Sprintf("%t", d)
	case string, *string:
		return fmt.Sprintf("%s", d)
	case int, *int, int64, *int64:
		return fmt.Sprintf("%d", d)
	default:
		j, _ := json.Marshal(d)
		return string(j)
	}
}
