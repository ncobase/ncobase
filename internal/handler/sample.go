package handler

import (
	"net/http"
	"stocms/internal/data/structs"
	"stocms/pkg/resp"
	"stocms/pkg/validator"

	"github.com/gin-gonic/gin"
)

// Hello is a handler for the "Hello" endpoint.
// @Produce json
// @Success 200 {object} resp.Exception "ok"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /sample/hello [get]
func (h *Handler) Hello(ctx *gin.Context) {
	var (
		err  error
		body structs.Sample
	)

	if err = ctx.ShouldBind(&body); err != nil {
		exception := &resp.Exception{
			Status:  http.StatusBadRequest,
			Message: err.Error(),
		}
		resp.Fail(ctx, exception)
		return
	}

	result, err := h.svc.Hello(ctx, body)
	if validator.IsNotNil(err) {
		resp.Fail(ctx, result)
		return
	}

	resp.Success(ctx, result)
}
