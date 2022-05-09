package user

import (
	"github.com/crochee/lirity/e"
	"github.com/crochee/saty/internal/service"
	"github.com/crochee/saty/internal/service/user"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"net/http"
)

type ListRequest struct {
}

func List(ctx *gin.Context) {
	var request ListRequest
	if err := ctx.ShouldBindBodyWith(&request, binding.JSON); err != nil {
		e.Code(ctx, e.ErrInvalidParam.WithResult(err))
		return
	}
	resp, err := service.NewService().Users().List(ctx.Request.Context(), &user.ListOpts{})
	if err != nil {
		e.Error(ctx, err)
		return
	}
	ctx.JSONP(http.StatusOK, resp)
}
