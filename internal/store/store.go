package store

// Store defines the storage interface.
type Store interface {
	Begin() Store
	Commit()
	Rollback()
	Area() AreaStore
}
