package service

import (
	"context"
	"ncobase/common/ecode"
	"ncobase/common/resp"
	"ncobase/common/types"
	"ncobase/feature/auth/data"
	"ncobase/feature/auth/data/repository"
	"ncobase/feature/auth/structs"

	"github.com/dchest/captcha"
)

// CaptchaServiceInterface is the interface for the service.
type CaptchaServiceInterface interface {
	GenerateCaptchaService(ctx context.Context, ext string) (*resp.Exception, error)
	GetCaptchaService(ctx context.Context, id string) *resp.Exception
	ValidateCaptchaService(ctx context.Context, body *structs.Captcha) *resp.Exception
}

// captchaService is the struct for the service.
type captchaService struct {
	captcha repository.CaptchaRepositoryInterface
}

// NewCaptchaService creates a new service.
func NewCaptchaService(d *data.Data) CaptchaServiceInterface {
	return &captchaService{
		captcha: repository.NewCaptchaRepository(d),
	}
}

// GenerateCaptchaService generates a new captcha ID and image URL.
func (s *captchaService) GenerateCaptchaService(ctx context.Context, ext string) (*resp.Exception, error) {
	captchaID := captcha.New()
	captchaURL := "/v1/captcha/" + captchaID + ext

	// Set captcha ID in cache
	if err := s.captcha.Set(ctx, captchaID, &types.JSON{"id": captchaID, "url": captchaURL}); err != nil {
		return resp.InternalServer(err.Error()), nil
	}

	return &resp.Exception{
		Data: &types.JSON{"url": captchaURL},
	}, nil
}

// GetCaptchaService gets the captcha from the cache.
func (s *captchaService) GetCaptchaService(ctx context.Context, id string) *resp.Exception {
	cached, err := s.captcha.Get(ctx, id)
	if err != nil {
		return resp.NotFound(ecode.NotExist("captcha"))
	}
	return &resp.Exception{
		Data: cached,
	}
}

// ValidateCaptchaService validates the captcha code.
func (s *captchaService) ValidateCaptchaService(ctx context.Context, body *structs.Captcha) *resp.Exception {
	if body == nil || !captcha.VerifyString(body.ID, body.Solution) {
		return resp.BadRequest(ecode.FieldIsInvalid("captcha"))
	}

	// Delete captcha after verification
	if err := s.captcha.Delete(ctx, body.ID); err != nil {
		return resp.InternalServer(err.Error())
	}

	return &resp.Exception{}
}
