package service

import (
	"template/internal/gateway"
	"template/internal/service/area"
	"template/internal/store"
)

type Service interface {
	Area() area.AreaSrv
}

func NewService(store store.Store, client gateway.Client) Service {
	return service{
		store:  store,
		client: client,
	}
}

type service struct {
	store  store.Store
	client gateway.Client
}

func (s service) Area() area.AreaSrv {
	return area.NewAreaSrv(s.store, s.client)
}
