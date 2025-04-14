package service

import (
	"context"
	"errors"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/repository"
	"ncobase/core/auth/structs"

	"github.com/ncobase/ncore/ecode"
	"github.com/ncobase/ncore/types"

	"github.com/dchest/captcha"
)

// CaptchaServiceInterface is the interface for the service.
type CaptchaServiceInterface interface {
	GenerateCaptcha(ctx context.Context, ext string) (*types.JSON, error)
	GetCaptcha(ctx context.Context, id string) (*types.JSON, error)
	ValidateCaptcha(ctx context.Context, body *structs.Captcha) error
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

// GenerateCaptcha generates a new captcha ID and image URL.
func (s *captchaService) GenerateCaptcha(ctx context.Context, ext string) (*types.JSON, error) {
	captchaID := captcha.New()
	captchaURL := "/auth/captcha/" + captchaID + ext

	// Set captcha ID in cache
	if err := s.captcha.Set(ctx, captchaID, &types.JSON{"id": captchaID, "url": captchaURL}); err != nil {
		return nil, err
	}

	return &types.JSON{"url": captchaURL}, nil
}

// GetCaptcha gets the captcha from the cache.
func (s *captchaService) GetCaptcha(ctx context.Context, id string) (*types.JSON, error) {
	return s.captcha.Get(ctx, id)
}

// ValidateCaptcha validates the captcha code.
func (s *captchaService) ValidateCaptcha(ctx context.Context, body *structs.Captcha) error {
	if body == nil || !captcha.VerifyString(body.ID, body.Solution) {
		return errors.New(ecode.FieldIsInvalid("captcha"))
	}

	// Delete captcha after verification
	if err := s.captcha.Delete(ctx, body.ID); err != nil {
		return err
	}

	return nil
}
