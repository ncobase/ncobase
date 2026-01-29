package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/repository"
	"ncobase/core/auth/structs"
	"ncobase/core/auth/wrapper"
	userService "ncobase/core/user/service"
	"ncobase/internal/utils"
	"strings"
	"time"

	"github.com/ncobase/ncore/ctxutil"
	"github.com/ncobase/ncore/logging/logger"
	"github.com/ncobase/ncore/security/jwt"
	"github.com/ncobase/ncore/types"
	"github.com/ncobase/ncore/utils/nanoid"
	"github.com/ncobase/ncore/validation/validator"
	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

const (
	mfaTokenSubject          = "mfa"
	mfaTokenExpiry           = 5 * time.Minute
	mfaMaxFailedAttempts     = 5
	mfaLockDuration          = 10 * time.Minute
	mfaRecoveryCodesCount    = 10
	mfaRecoveryCodeBytes     = 10
	mfaRecoveryCodeGroupSize = 5
	mfaIssuer                = "Ncobase"
)

type MFAServiceInterface interface {
	IsEnabled(ctx context.Context, userID string) (bool, error)
	CreateLoginChallenge(ctx context.Context, userID string) (string, []string, error)
	VerifyLoginChallenge(ctx context.Context, mfaToken string, code string, recoveryCode string) (*AuthResponse, error)

	GetTwoFactorStatus(ctx context.Context) (*structs.TwoFactorStatusResponse, error)
	SetupTwoFactor(ctx context.Context, method string) (*structs.TwoFactorSetupResponse, error)
	VerifyTwoFactor(ctx context.Context, code string, method string) (*structs.RecoveryCodesResponse, error)
	DisableTwoFactor(ctx context.Context, password string, code string, recoveryCode string) error
	RegenerateRecoveryCodes(ctx context.Context, code string) (*structs.RecoveryCodesResponse, error)
}

type mfaService struct {
	d             *data.Data
	jtm           *jwt.TokenManager
	mfaRepo       repository.UserMFARepositoryInterface
	authTokenRepo repository.AuthTokenRepositoryInterface
	usw           *wrapper.UserServiceWrapper
	asw           *wrapper.AccessServiceWrapper
	tsw           *wrapper.SpaceServiceWrapper
	enc           *utils.EncryptionService
	ss            SessionServiceInterface
}

func NewMFAService(
	d *data.Data,
	jtm *jwt.TokenManager,
	usw *wrapper.UserServiceWrapper,
	asw *wrapper.AccessServiceWrapper,
	tsw *wrapper.SpaceServiceWrapper,
	ss SessionServiceInterface,
) MFAServiceInterface {
	var enc *utils.EncryptionService
	encryptionConfig := &utils.EncryptionConfig{
		MasterKey:       "",
		KeyVersion:      1,
		RotationEnabled: false,
	}
	enc, err := utils.NewEncryptionService(encryptionConfig)
	if err != nil {
		logger.Warnf(context.Background(), "MFA encryption service unavailable: %v", err)
	}

	return &mfaService{
		d:             d,
		jtm:           jtm,
		mfaRepo:       repository.NewUserMFARepository(d),
		authTokenRepo: repository.NewAuthTokenRepository(d),
		usw:           usw,
		asw:           asw,
		tsw:           tsw,
		enc:           enc,
		ss:            ss,
	}
}

func (s *mfaService) IsEnabled(ctx context.Context, userID string) (bool, error) {
	if userID == "" {
		return false, errors.New("user_id is required")
	}

	row, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return row.Enabled, nil
}

func (s *mfaService) CreateLoginChallenge(ctx context.Context, userID string) (string, []string, error) {
	ip, userAgent, _ := ctxutil.GetClientInfo(ctx)
	payload := map[string]any{
		"user_id":    userID,
		"ip":         ip,
		"user_agent": userAgent,
		"ts":         time.Now().UnixMilli(),
	}

	jti := nanoid.Must(32)
	token, err := s.jtm.GenerateRegisterToken(jti, payload, mfaTokenSubject, &jwt.TokenConfig{Expiry: mfaTokenExpiry})
	if err != nil {
		return "", nil, err
	}

	return token, []string{"totp", "recovery_code"}, nil
}

func (s *mfaService) VerifyLoginChallenge(ctx context.Context, mfaToken string, code string, recoveryCode string) (*AuthResponse, error) {
	if validator.IsEmpty(mfaToken) {
		return nil, errors.New("mfa_token is required")
	}

	claims, err := s.jtm.DecodeToken(mfaToken)
	if err != nil {
		return nil, errors.New("invalid mfa_token")
	}
	if claims["sub"] != mfaTokenSubject {
		return nil, errors.New("invalid mfa_token type")
	}

	payload, ok := claims["payload"].(map[string]any)
	if !ok {
		return nil, errors.New("invalid mfa_token payload")
	}

	userID, _ := payload["user_id"].(string)
	if userID == "" {
		return nil, errors.New("invalid mfa_token user")
	}

	tokenIP, _ := payload["ip"].(string)
	tokenUA, _ := payload["user_agent"].(string)
	ip, userAgent, _ := ctxutil.GetClientInfo(ctx)
	if tokenIP != "" && ip != "" && tokenIP != ip {
		return nil, errors.New("mfa_token context mismatch")
	}
	if tokenUA != "" && userAgent != "" && tokenUA != userAgent {
		return nil, errors.New("mfa_token context mismatch")
	}

	if err := s.verifyUserMFA(ctx, userID, code, recoveryCode); err != nil {
		return nil, err
	}

	user, err := s.usw.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	userSpaces, _ := s.tsw.GetUserSpaces(ctx, user.ID)
	var spaceIDs []string
	for _, t := range userSpaces {
		spaceIDs = append(spaceIDs, t.ID)
	}

	defaultSpace, err := s.tsw.GetUserSpace(ctx, user.ID)
	if err == nil && defaultSpace != nil {
		ctx = ctxutil.SetSpaceID(ctx, defaultSpace.ID)
	}

	tokenPayload, err := CreateUserTokenPayload(ctx, user, spaceIDs, s.asw, s.tsw)
	if err != nil {
		return nil, err
	}

	authResp, err := generateAuthResponse(ctx, s.jtm, s.authTokenRepo, tokenPayload, s.ss, "password+mfa")
	if err != nil {
		return nil, err
	}

	authResp.SpaceIDs = spaceIDs
	if defaultSpace != nil {
		authResp.DefaultSpace = mapToDefaultSpace(defaultSpace.ID, defaultSpace.Name)
	}

	return authResp, nil
}

func mapToDefaultSpace(id string, name string) *types.JSON {
	return &types.JSON{
		"id":   id,
		"name": name,
	}
}

func (s *mfaService) GetTwoFactorStatus(ctx context.Context) (*structs.TwoFactorStatusResponse, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	row, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		if repository.IsNotFound(err) {
			return &structs.TwoFactorStatusResponse{Enabled: false}, nil
		}
		return nil, err
	}

	remaining := len(row.RecoveryCodeHashes)

	method := ""
	if row.Enabled {
		method = "app"
	}
	return &structs.TwoFactorStatusResponse{
		Enabled:                row.Enabled,
		Method:                 method,
		RecoveryCodesRemaining: remaining,
	}, nil
}

func (s *mfaService) SetupTwoFactor(ctx context.Context, method string) (*structs.TwoFactorSetupResponse, error) {
	if method != "app" {
		return nil, errors.New("unsupported 2fa method")
	}

	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	user, err := s.usw.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if s.enc == nil {
		return nil, errors.New("encryption service not available")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      mfaIssuer,
		AccountName: user.Username,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate totp secret: %w", err)
	}

	encrypted, err := s.enc.Encrypt(key.Secret())
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt totp secret: %w", err)
	}

	// Try to create, fall back to update on constraint error
	_, err = s.mfaRepo.Create(ctx, userID, encrypted)
	if err != nil {
		if repository.IsConstraintError(err) {
			_, err = s.mfaRepo.UpdateSetup(ctx, userID, encrypted)
		}
		if err != nil {
			return nil, err
		}
	}

	png, err := qrcode.Encode(key.URL(), qrcode.Medium, 256)
	if err != nil {
		return nil, fmt.Errorf("failed to generate qr code: %w", err)
	}

	return &structs.TwoFactorSetupResponse{
		Method:     "app",
		Secret:     key.Secret(),
		OTPAuthURI: key.URL(),
		QRPNG:      base64.StdEncoding.EncodeToString(png),
	}, nil
}

func (s *mfaService) VerifyTwoFactor(ctx context.Context, code string, method string) (*structs.RecoveryCodesResponse, error) {
	if method != "app" {
		return nil, errors.New("unsupported 2fa method")
	}

	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	row, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errors.New("2fa setup not initialized")
	}
	if row.TotpSecret == "" {
		return nil, errors.New("2fa setup not initialized")
	}

	secret, err := s.decryptSecret(row.TotpSecret)
	if err != nil {
		return nil, err
	}

	ok, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{Period: 30, Skew: 1, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil || !ok {
		return nil, errors.New("invalid authentication code")
	}

	recoveryCodes, hashes, err := generateRecoveryCodes()
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	_, err = s.mfaRepo.Enable(ctx, userID, now, hashes, now)
	if err != nil {
		return nil, err
	}

	return &structs.RecoveryCodesResponse{RecoveryCodes: recoveryCodes}, nil
}

func (s *mfaService) DisableTwoFactor(ctx context.Context, password string, code string, recoveryCode string) error {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return errors.New("user not authenticated")
	}

	verifyResult := s.usw.VerifyPassword(ctx, userID, password)
	switch v := verifyResult.(type) {
	case userService.VerifyPasswordResult:
		if !v.Valid {
			return errors.New(v.Error)
		}
	case error:
		return v
	default:
		return errors.New("password verification failed")
	}

	if err := s.verifyUserMFA(ctx, userID, code, recoveryCode); err != nil {
		return err
	}

	_, err := s.mfaRepo.Disable(ctx, userID)
	return err
}

func (s *mfaService) RegenerateRecoveryCodes(ctx context.Context, code string) (*structs.RecoveryCodesResponse, error) {
	userID := ctxutil.GetUserID(ctx)
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	row, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil || !row.Enabled {
		return nil, errors.New("2fa is not enabled")
	}

	secret, err := s.decryptSecret(row.TotpSecret)
	if err != nil {
		return nil, err
	}

	ok, err := totp.ValidateCustom(code, secret, time.Now(), totp.ValidateOpts{Period: 30, Skew: 1, Digits: otp.DigitsSix, Algorithm: otp.AlgorithmSHA1})
	if err != nil || !ok {
		return nil, errors.New("invalid authentication code")
	}

	recoveryCodes, hashes, err := generateRecoveryCodes()
	if err != nil {
		return nil, err
	}

	_, err = s.mfaRepo.UpdateRecoveryCodes(ctx, userID, hashes, time.Now().UnixMilli())
	if err != nil {
		return nil, err
	}

	return &structs.RecoveryCodesResponse{RecoveryCodes: recoveryCodes}, nil
}

func (s *mfaService) decryptSecret(ciphertext string) (string, error) {
	if s.enc == nil {
		return "", errors.New("encryption service not available")
	}
	if ciphertext == "" {
		return "", errors.New("totp secret not initialized")
	}
	secret, err := s.enc.Decrypt(ciphertext)
	if err != nil {
		return "", errors.New("failed to decrypt totp secret")
	}
	return secret, nil
}

func (s *mfaService) verifyUserMFA(ctx context.Context, userID string, code string, recoveryCode string) error {
	row, err := s.mfaRepo.GetByUserID(ctx, userID)
	if err != nil || !row.Enabled {
		return errors.New("mfa not enabled")
	}

	now := time.Now().UnixMilli()
	if row.LockedUntil > now {
		return errors.New("too many failed attempts, try again later")
	}

	if recoveryCode != "" {
		return s.verifyRecoveryCode(ctx, row.UserID, row.RecoveryCodeHashes, row.FailedAttempts, recoveryCode)
	}

	if code == "" {
		return errors.New("authentication code is required")
	}

	secret, err := s.decryptSecret(row.TotpSecret)
	if err != nil {
		return err
	}

	valid := totp.Validate(code, secret)
	if !valid {
		return s.recordFailedAttempt(ctx, userID, row.FailedAttempts)
	}

	_, err = s.mfaRepo.ResetFailedAttempts(ctx, userID, now)
	return err
}

func (s *mfaService) verifyRecoveryCode(ctx context.Context, userID string, recoveryCodeHashes []string, failedAttempts int, recoveryCode string) error {
	list := recoveryCodeHashes
	if len(list) == 0 {
		return errors.New("invalid recovery code")
	}

	normalized := normalizeRecoveryCode(recoveryCode)
	sum := sha256.Sum256([]byte(normalized))
	target := base64.StdEncoding.EncodeToString(sum[:])

	idx := -1
	for i, h := range list {
		if h == target {
			idx = i
			break
		}
	}
	if idx == -1 {
		return s.recordFailedAttempt(ctx, userID, failedAttempts)
	}

	remaining := append([]string{}, list[:idx]...)
	remaining = append(remaining, list[idx+1:]...)

	now := time.Now().UnixMilli()
	_, err := s.mfaRepo.UpdateRecoveryCodesAndReset(ctx, userID, remaining, 0, now)
	return err
}

func (s *mfaService) recordFailedAttempt(ctx context.Context, userID string, current int) error {
	next := current + 1

	if next >= mfaMaxFailedAttempts {
		lockedUntil := time.Now().Add(mfaLockDuration).UnixMilli()
		s.mfaRepo.LockAccount(ctx, userID, lockedUntil)
	} else {
		s.mfaRepo.IncrementFailedAttempts(ctx, userID, next)
	}

	return errors.New("verification failed")
}

func generateRecoveryCodes() ([]string, []string, error) {
	codes := make([]string, 0, mfaRecoveryCodesCount)
	hashes := make([]string, 0, mfaRecoveryCodesCount)

	for i := 0; i < mfaRecoveryCodesCount; i++ {
		raw, err := randomBase32(mfaRecoveryCodeBytes)
		if err != nil {
			return nil, nil, err
		}
		code := formatRecoveryCode(raw)
		codes = append(codes, code)

		normalized := normalizeRecoveryCode(code)
		sum := sha256.Sum256([]byte(normalized))
		hashes = append(hashes, base64.StdEncoding.EncodeToString(sum[:]))
	}

	return codes, hashes, nil
}

func formatRecoveryCode(input string) string {
	s := strings.ToUpper(strings.ReplaceAll(input, "=", ""))
	s = strings.ReplaceAll(s, " ", "")

	parts := make([]string, 0, (len(s)+mfaRecoveryCodeGroupSize-1)/mfaRecoveryCodeGroupSize)
	for i := 0; i < len(s); i += mfaRecoveryCodeGroupSize {
		end := i + mfaRecoveryCodeGroupSize
		if end > len(s) {
			end = len(s)
		}
		parts = append(parts, s[i:end])
	}
	return strings.Join(parts, "-")
}

func normalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(code, "-", ""))
}

func randomBase32(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b), nil
}
