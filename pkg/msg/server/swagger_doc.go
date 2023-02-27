package server

// swagger:parameters SwagNullRequest
type SwagNullRequest struct{}

// swagger:response SGetProduceConfigRes
type SGetProduceConfigRes struct {
	// in: body
	Body struct {
		ProduceConfig
	}
}

// swagger:parameters SUpdateProduceConfigRequest
type SUpdateProduceConfigRequest struct {
	ProduceConfig
}

// swagger:response ResponseCode
type ResponseCode struct {
	// 返回码
	// Required: true
	// Example: 200
	Code int `json:"code" binding:"required"`
	// 返回信息描述
	// Required: true
	// Example: success
	Msg string `json:"message" binding:"required"`
}
