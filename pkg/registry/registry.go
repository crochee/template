package registry

type Registry interface {
	Register(info *Info) error
	Deregister(info *Info) error
}

// Info is used for registry.
// The fields are just suggested, which is used depends on design.
type Info struct {
	UUID        string
	ServiceName string
	Addr        string
	Weight      int
	Tags        map[string]interface{}
}
