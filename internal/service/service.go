package service

import "go_template/internal/service/user"

type Service interface {
	Users() user.UserSrv
}

func NewService() Service {
	return &service{}
}

type service struct {
}

func (s *service) Users() user.UserSrv {
	return user.NewUserService()
}
