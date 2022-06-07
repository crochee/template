package store

var client Factory

// Factory defines the storage interface.
type Factory interface {
	Begin() Factory
	Commit()
	Rollback()
	Auth() AuthorControlStore
	Flow() ChangeFlowStore
	Pkg() ResourcePkgStore
}

// Client return the store client instance.
func Client() Factory {
	return client
}

// SetClient set the iam store client.
func SetClient(factory Factory) {
	client = factory
}
