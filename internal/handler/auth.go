package handler

import (
	"bytes"
	"fmt"
	"io"
	"ncobase/helper"
	"ncobase/internal/data/structs"
	"path/filepath"
	"strings"

	"ncobase/common/cookie"
	"ncobase/common/resp"
	"ncobase/common/types"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

// SendCodeHandler handles sending a verification code.
//
// @Summary Send verification code
// @Description Send a verification code to the specified destination.
// @Tags authorization
// @Accept json
// @Produce json
// @Param body body structs.SendCodeBody true "SendCodeBody object"
// @Success 200 {object} types.JSON{registered=bool} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/authorize/send [post]
func (h *Handler) SendCodeHandler(c *gin.Context) {
	body := &structs.SendCodeBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.svc.SendCodeService(c, body)
	resp.Success(c.Writer, result)
}

// CodeAuthHandler handles verifying a code.
//
// @Summary Verify code
// @Description Verify the provided code.
// @Tags authorization
// @Tags authentication
// @Accept json
// @Produce json
// @Param code path string true "Verification code"
// @Success 200 {object} types.JSON{id=string,access_token=string,email=string,register_token=string}  "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/authorize/{code} [get]
func (h *Handler) CodeAuthHandler(c *gin.Context) {
	result, err := h.svc.CodeAuthService(c, c.Param("code"))
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// RegisterHandler handles user registration.
//
// @Summary Register
// @Description Register a new user.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.RegisterBody true "RegisterBody object"
// @Success 200 {object} types.JSON{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/register [post]
func (h *Handler) RegisterHandler(c *gin.Context) {
	body := &structs.RegisterBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, _ := h.svc.RegisterService(c, body)
	resp.Success(c.Writer, result)
}

// LogoutHandler handles user logout.
//
// @Summary Logout
// @Description Logout the current user.
// @Tags authentication
// @Produce json
// @Success 200 {object} resp.Exception "success"
// @Router /v1/logout [post]
// @Security Bearer
func (h *Handler) LogoutHandler(c *gin.Context) {
	cookie.ClearAll(c.Writer)
	resp.Success(c.Writer, nil)
}

// LoginHandler handles user login.
//
// @Summary Login
// @Description Log in a user.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.LoginBody true "LoginBody object"
// @Success 200 {object} types.JSON{id=string,access_token=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/login [post]
func (h *Handler) LoginHandler(c *gin.Context) {
	body := &structs.LoginBody{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	// Validate captcha
	if body.Captcha != nil && body.Captcha.ID != "" && body.Captcha.Solution != "" {
		if result := h.svc.ValidateCaptchaService(c, body.Captcha); result.Code != 0 {
			resp.Fail(c.Writer, result)
			return
		}
	}

	result, err := h.svc.LoginService(c, body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// // RefreshHandler handles user token refresh.
// //
// // @Summary Refresh
// // @Description Refresh the current user's access token.
// // @Tags authentication
// // @Produce json
// // @Success 200 {object} types.JSON{id=string,access_token=string} "success"
// // @Failure 400 {object} resp.Exception "bad request"
// // @Router /v1/refresh [post]
// // @Security Bearer
// func (h *Handler) RefreshHandler(c *gin.Context) {
// 	result, err := h.svc.RefreshService(c)
// 	if err != nil {
// 		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
// 		return
// 	}
// 	resp.Success(c.Writer, result)
// }

// GenerateCaptchaHandler handles generating a captcha image.
//
// @Summary Generate captcha
// @Description Generate a captcha image.
// @Tags authentication
// @Produce json
// @Param type query string false "Captcha type" Enums(png, wav)
// @Success 200 {object} types.JSON{id=string,url=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/captcha/generate [get]
func (h *Handler) GenerateCaptchaHandler(c *gin.Context) {
	ext := c.Query("type")
	switch ext {
	case "png":
		ext = ".png"
	case "wav":
		ext = ".wav"
	default:
		ext = ".png"
	}
	result, err := h.svc.GenerateCaptchaService(c, ext)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// ValidateCaptchaHandler handles validating a captcha code.
//
// @Summary Validate captcha
// @Description Validate a captcha code.
// @Tags authentication
// @Accept json
// @Produce json
// @Param body body structs.Captcha true "Captcha object"
// @Success 200 {object} types.JSON{message=string} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/captcha/validate [post]
func (h *Handler) ValidateCaptchaHandler(c *gin.Context) {
	body := &structs.Captcha{}
	if validationErrors, err := helper.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}
	result := h.svc.ValidateCaptchaService(c, body)
	resp.Success(c.Writer, result)
}

// CaptchaStreamHandler handles streaming a captcha image.
//
// @Summary Stream captcha
// @Description Stream a captcha image.
// @Tags authentication
// @Produce json
// @Param captcha_id path string true "Captcha ID With Extension (png, wav)"
// @Success 200 {file} octet-stream
// @Failure 400 {object} resp.Exception "bad request"
// @Router /v1/captcha/{captcha_id} [get]
// CaptchaStreamHandler handles streaming a captcha image or audio.
func (h *Handler) CaptchaStreamHandler(c *gin.Context) {
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

	result := h.svc.GetCaptchaService(c, id)
	if result.Code != 0 {
		resp.Fail(c.Writer, result)
		return
	}

	data, ok := result.Data.(*types.JSON)
	if !ok {
		resp.Fail(c.Writer, resp.InternalServer("Invalid data type in response"))
		return
	}

	cachedCaptchaID, ok := (*data)["id"].(string)
	if !ok {
		resp.Fail(c.Writer, resp.InternalServer("Invalid captcha ID type"))
		return
	}

	var content bytes.Buffer
	var err error

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
