// Package msg
package msg

type Plugin interface {
	Name() string
	Use(*Metadata)
}
