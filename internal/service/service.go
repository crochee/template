package service

import (
	"github.com/crochee/devt/internal/service/authenticate"
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
