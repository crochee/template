package gateway

import (
	"github.com/spf13/viper"

	"template/internal/gateway/dcs"
	"template/pkg/client"
)

type Client interface {
	Area() dcs.AreaSrv
}

func NewBaseClient() Client {
	return &baseClient{
		IRequest: client.NewResource().
			AddEndpoint(viper.GetString("ifp.url"))}
}

type baseClient struct {
	client.IRequest
}

func (c baseClient) Area() dcs.AreaSrv {
	return dcs.AreaClient{IRequest: c.AddPath("areas_detail")}
}
