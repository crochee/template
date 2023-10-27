package ptf

import (
	"io"
	"os"
)

// Headers set csv header
func Headers(headers []string) Option {
	return func(o *option) {
		o.headers = headers
	}
}

func FieldNames(fieldNames []string) Option {
	return func(o *option) {
		o.fieldNames = fieldNames
	}
}

func TagName(tagName string) Option {
	return func(o *option) {
		o.tagName = tagName
	}
}

func Writer(w io.Writer) Option {
	return func(o *option) {
		o.writer = w
	}
}

func SetHandler(h Handler) Option {
	return func(o *option) {
		o.handler = h
	}
}

func Sheet(sheet string) Option {
	return func(o *option) {
		o.sheet = sheet
	}
}

type Handler func(name string, headers []string, rows [][]interface{}, w io.Writer) error

type option struct {
	tagName    string
	fieldNames []string
	headers    []string
	writer     io.Writer
	handler    Handler
	sheet      string
}

type Option func(*option)

func NewMarshal(opts ...Option) *marshal {
	opt := option{
		tagName:    "csv",
		fieldNames: []string{"List"},
		headers:    []string{},
		writer:     os.Stdout,
		handler: func(name string, headers []string, rows [][]interface{}, w io.Writer) error {
			return nil
		},
		sheet: "csv",
	}
	for _, f := range opts {
		f(&opt)
	}
	return &marshal{
		fieldNames: opt.fieldNames,
		headers:    opt.headers,
		expose:     expose{},
		parser:     &parse{tagName: opt.tagName},
		w:          opt.writer,
		handler:    opt.handler,
		sheet:      opt.sheet,
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
