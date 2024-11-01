package handler

import (
	"bytes"
	"fmt"
	"io"
	"ncobase/common/helper"
	"ncobase/common/resp"
	"ncobase/core/auth/service"
	"ncobase/core/auth/structs"
	"path/filepath"
	"strings"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

// CaptchaHandlerInterface is the interface for the handler.
type CaptchaHandlerInterface interface {
	GenerateCaptcha(c *gin.Context)
	ValidateCaptcha(c *gin.Context)
	CaptchaStream(c *gin.Context)
}

// captchaHandler represents the handler.
type captchaHandler struct {
	s *service.Service
}

// NewCaptchaHandler creates a new handler.
func NewCaptchaHandler(svc *service.Service) CaptchaHandlerInterface {
	return &captchaHandler{
		s: svc,
	}
}

// GenerateCaptcha handles generating a captcha image.
//
// @Summary Generate captcha
// @Description Generate a captcha image.
// @Tags authentication
// @Produce json
// @Param type query string false "Captcha type" Enums(png, wav)
// @Success 200 {object} map[string]any{id=string,url=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/captcha/generate [get]
func (h *captchaHandler) GenerateCaptcha(c *gin.Context) {
	ext := c.Query("type")
	switch ext {
	case "png":
		ext = ".png"
	case "wav":
		ext = ".wav"
	default:
		ext = ".png"
	}
	result, err := h.s.Captcha.GenerateCaptcha(c.Request.Context(), ext)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ValidateCaptcha handles validating a captcha code.
//
// @Summary Validate captcha
// @Description Validate a captcha code.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.Captcha true "Captcha object"
// @Success 200 {object} map[string]any{message=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/captcha/validate [post]
func (h *captchaHandler) ValidateCaptcha(c *gin.Context) {
	body := &structs.Captcha{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}
	if err := h.s.Captcha.ValidateCaptcha(c.Request.Context(), body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer)
}

// CaptchaStream handles streaming a captcha image.
//
// @Summary Stream captcha
// @Description Stream a captcha image.
// @Tags authentication
// @Produce json
// @Param captcha_id path string true "Captcha ID With Extension (png, wav)"
// @Success 200 {file} octet-stream
// @Failure 400 {object} resp.Exception "bad request"
// @Router /iam/captcha/{captcha_id} [get]
// CaptchaStreamHandler handles streaming a captcha image or audio.
func (h *captchaHandler) CaptchaStream(c *gin.Context) {
	captchaID := c.Param("captcha")
	ext := filepath.Ext(captchaID)

	// Set default extension if not provided
	if ext == "" {
		ext = ".png"
	}

	id := strings.TrimSuffix(captchaID, ext)
	validExts := map[string]string{
		".png": "image/png",
		".wav": "audio/x-wav",
	}

	// Validate captcha ID and extension
	if id == "" || validExts[ext] == "" {
		resp.Fail(c.Writer, resp.BadRequest("Invalid captcha"))
		return
	}

	result, err := h.s.Captcha.GetCaptcha(c.Request.Context(), id)
	if err != nil || result == nil {
		resp.Fail(c.Writer, resp.BadRequest("Invalid captcha"))
		return
	}

	cachedCaptchaID, ok := (*result)["id"].(string)
	if !ok {
		resp.Fail(c.Writer, resp.InternalServer("Invalid captcha ID type"))
		return
	}

	var content bytes.Buffer

	switch ext {
	case ".png":
		err = captcha.WriteImage(&content, cachedCaptchaID, captcha.StdWidth, captcha.StdHeight)
	case ".wav":
		lang := c.Query("lang")
		err = captcha.WriteAudio(&content, cachedCaptchaID, lang)
	}

	if err != nil {
		resp.Fail(c.Writer, resp.InternalServer("Failed to generate captcha"))
		return
	}

	// Set response headers
	contentType := validExts[ext]
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%s%s", cachedCaptchaID, ext))
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")

	// Write content to response
	if _, err := io.Copy(c.Writer, &content); err != nil {
		resp.Fail(c.Writer, resp.InternalServer(err.Error()))
	}
}
