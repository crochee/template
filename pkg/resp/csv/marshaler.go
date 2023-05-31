package csv

import (
	"io"
	"os"
)

// Headers set csv header
func Headers(headers []string) func(*Option) {
	return func(o *Option) {
		o.headers = headers
	}
}

func FieldNames(fieldNames []string) func(*Option) {
	return func(o *Option) {
		o.fieldNames = fieldNames
	}
}

func TagName(tagName string) func(*Option) {
	return func(o *Option) {
		o.tagName = tagName
	}
}

func Writer(w io.Writer) func(*Option) {
	return func(o *Option) {
		o.writer = w
	}
}

func SetHandler(h Handler) func(*Option) {
	return func(o *Option) {
		o.handler = h
	}
}

func Sheet(sheet string) func(*Option) {
	return func(o *Option) {
		o.sheet = sheet
	}
}

type Handler func(name string, headers []string, rows [][]interface{}, w io.Writer) error

type Option struct {
	tagName    string
	fieldNames []string
	headers    []string
	writer     io.Writer
	handler    Handler
	sheet      string
}

func NewMarshal(opts ...func(*Option)) *marshal {
	option := Option{
		tagName:    "csv",
		fieldNames: []string{"List"},
		headers:    []string{},
		writer:     os.Stdout,
		handler:    CsvHandler,
		sheet:      "csv",
	}
	for _, opt := range opts {
		opt(&option)
	}
	return &marshal{
		fieldNames: option.fieldNames,
		headers:    option.headers,
		expose:     expose{},
		parser:     &parse{tagName: option.tagName},
		w:          option.writer,
		handler:    option.handler,
		sheet:      option.sheet,
	}
}

type marshal struct {
	fieldNames []string
	headers    []string
	expose     expose
	parser     *parse
	w          io.Writer
	handler    Handler
	sheet      string
}

func (m *marshal) Encode(obj interface{}) error {
	value, err := m.expose.GetStruct(obj, m.fieldNames...)
	if err != nil {
		return err
	}
	var data []*mapIndexValue
	if data, err = m.parser.parse(value); err != nil {
		return err
	}

	rows := make([][]interface{}, len(data))
	for i, d := range data {
		row := make([]interface{}, len(m.headers))
		for index, key := range m.headers {
			tempValue, found := d.data[key]
			if found {
				row[index] = tempValue
			}
		}
		rows[i] = row
	}
	return m.handler(m.sheet, m.headers, rows, m.w)
}
