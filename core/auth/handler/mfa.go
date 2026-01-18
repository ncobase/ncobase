package handler

import (
	"ncobase/core/auth/service"
	"ncobase/core/auth/structs"

	"github.com/ncobase/ncore/net/resp"
	"github.com/ncobase/ncore/validation"

	"github.com/gin-gonic/gin"
)

type MFAHandlerInterface interface {
	LoginMFA(c *gin.Context)

	GetTwoFactorStatus(c *gin.Context)
	SetupTwoFactor(c *gin.Context)
	VerifyTwoFactor(c *gin.Context)
	DisableTwoFactor(c *gin.Context)
	GetBackupCodes(c *gin.Context)
	RegenerateBackupCodes(c *gin.Context)
}

type mfaHandler struct {
	s *service.Service
}

func NewMFAHandler(svc *service.Service) MFAHandlerInterface {
	return &mfaHandler{s: svc}
}

// LoginMFA verifies MFA login challenge and returns auth tokens.
//
// @Summary Login MFA
// @Description Verify MFA challenge token and one-time code.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body structs.MFALoginVerifyBody true "MFALoginVerifyBody object"
// @Success 200 {object} service.AuthResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /login/mfa [post]
func (h *mfaHandler) LoginMFA(c *gin.Context) {
	body := &structs.MFALoginVerifyBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.Account.LoginMFA(c.Request.Context(), body)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}

	if result.SessionID != "" {
		_ = service.SetSessionCookie(c.Request.Context(), c.Writer, c.Request, result.SessionID)
	}

	resp.Success(c.Writer, result)
}

// GetTwoFactorStatus returns current 2FA status.
//
// @Summary 2FA status
// @Description Get current two-factor authentication status.
// @Tags auth
// @Produce json
// @Success 200 {object} structs.TwoFactorStatusResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/2fa/status [get]
// @Security Bearer
func (h *mfaHandler) GetTwoFactorStatus(c *gin.Context) {
	result, err := h.s.MFA.GetTwoFactorStatus(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// SetupTwoFactor starts 2FA setup and returns provisioning data.
//
// @Summary 2FA setup
// @Description Start two-factor authentication setup.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body structs.TwoFactorSetupBody true "TwoFactorSetupBody object"
// @Success 200 {object} structs.TwoFactorSetupResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/2fa/setup [post]
// @Security Bearer
func (h *mfaHandler) SetupTwoFactor(c *gin.Context) {
	body := &structs.TwoFactorSetupBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.MFA.SetupTwoFactor(c.Request.Context(), body.Method)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// VerifyTwoFactor enables 2FA after validating the OTP code.
//
// @Summary 2FA verify
// @Description Enable two-factor authentication.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body structs.TwoFactorVerifyBody true "TwoFactorVerifyBody object"
// @Success 200 {object} structs.RecoveryCodesResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/2fa/verify [post]
// @Security Bearer
func (h *mfaHandler) VerifyTwoFactor(c *gin.Context) {
	body := &structs.TwoFactorVerifyBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.MFA.VerifyTwoFactor(c.Request.Context(), body.Code, body.Method)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}

// DisableTwoFactor disables 2FA.
//
// @Summary 2FA disable
// @Description Disable two-factor authentication.
// @Tags auth
// @Accept json
// @Produce json
// @Param body body structs.TwoFactorDisableBody true "TwoFactorDisableBody object"
// @Success 200 {object} resp.Exception "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/2fa/disable [post]
// @Security Bearer
func (h *mfaHandler) DisableTwoFactor(c *gin.Context) {
	body := &structs.TwoFactorDisableBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	if err := h.s.MFA.DisableTwoFactor(c.Request.Context(), body.Password, body.Code, body.RecoveryCode); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, nil)
}

// GetBackupCodes returns remaining count (codes are only returned on generation).
//
// @Summary Backup codes
// @Description Get remaining backup codes count.
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]any{remaining=int} "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/2fa/backup-codes [get]
// @Security Bearer
func (h *mfaHandler) GetBackupCodes(c *gin.Context) {
	status, err := h.s.MFA.GetTwoFactorStatus(c.Request.Context())
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, map[string]any{"remaining": status.RecoveryCodesRemaining})
}

// RegenerateBackupCodes regenerates backup codes.
//
// @Summary Backup codes regenerate
// @Description Regenerate backup codes (requires OTP code).
// @Tags auth
// @Accept json
// @Produce json
// @Param body body structs.TwoFactorVerifyBody true "TwoFactorVerifyBody object (use code, method=app)"
// @Success 200 {object} structs.RecoveryCodesResponse "success"
// @Failure 400 {object} resp.Exception "bad request"
// @Router /account/2fa/backup-codes/regenerate [post]
// @Security Bearer
func (h *mfaHandler) RegenerateBackupCodes(c *gin.Context) {
	body := &structs.TwoFactorVerifyBody{}
	if validationErrors, err := validation.ShouldBindAndValidateStruct(c, body); err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	} else if len(validationErrors) > 0 {
		resp.Fail(c.Writer, resp.BadRequest("Invalid parameters", validationErrors))
		return
	}

	result, err := h.s.MFA.RegenerateRecoveryCodes(c.Request.Context(), body.Code)
	if err != nil {
		resp.Fail(c.Writer, resp.BadRequest(err.Error()))
		return
	}
	resp.Success(c.Writer, result)
}
