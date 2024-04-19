package utils

// Status 存储目前的状态
type Status struct {
	Flag int
}

// SetStatus 设置状态
func (s *Status) SetStatus(status int) {
	s.Flag = status
}

// AddStatus 添加一种或多种状态
func (s *Status) AddStatus(status int) {
	s.Flag |= status
}

// DeleteStatus 删除一种或者多种状态
func (s *Status) DeleteStatus(status int) {
	s.Flag &= ^status
}

// HasStatus 是否具有某些状态
func (s *Status) HasStatus(status int) bool {
	return (s.Flag & status) == status
}

// NotHasStatus 是否不具有某些状态
func (s *Status) NotHasStatus(status int) bool {
	return (s.Flag & status) == 0
}

// OnlyHas 是否仅仅具有某些状态
func (s *Status) OnlyHas(status int) bool {
	return s.Flag == status
}
