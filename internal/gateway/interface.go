package gateway

import (
	"github.com/spf13/viper"

	"go_template/internal/gateway/dcs"
	"go_template/pkg/client"
)

type IClient interface {
	Area() dcs.AreaSrv
}

var iClient IClient

func SetClient(c IClient) {
	iClient = c
}

func NewClient() IClient {
	return iClient
}

func NewBaseClient() IClient {
	return &baseClient{client.NewResource().AddEndpoint(viper.GetString("ifp.url"))}
}

type baseClient struct {
	client.IRequest
}

func (c baseClient) Area() dcs.AreaSrv {
	return dcs.AreaClient{c.AddPath("areas_detail")}
}
