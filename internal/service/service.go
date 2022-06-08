package service

import (
	"go_template/internal/service/authenticate"
)

type Service interface {
	Auth() authenticate.AuthenticateSrv
}

func NewService() Service {
	return &service{}
}

type service struct {
}

func (s *service) Auth() authenticate.AuthenticateSrv {
	return authenticate.NewAuthenticateSrv()
}
