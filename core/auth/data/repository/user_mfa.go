package repository

import (
	"context"
	"ncobase/core/auth/data"
	"ncobase/core/auth/data/ent"
	userMFAEnt "ncobase/core/auth/data/ent/usermfa"
)

// UserMFARepositoryInterface defines user MFA repository operations
type UserMFARepositoryInterface interface {
	GetByUserID(ctx context.Context, userID string) (*ent.UserMFA, error)
	Create(ctx context.Context, userID string, totpSecret string) (*ent.UserMFA, error)
	UpdateSetup(ctx context.Context, userID string, totpSecret string) (int, error)
	Enable(ctx context.Context, userID string, verifiedAt int64, recoveryHashes []string, recoveryGeneratedAt int64) (int, error)
	Disable(ctx context.Context, userID string) (int, error)
	UpdateRecoveryCodes(ctx context.Context, userID string, hashes []string, generatedAt int64) (int, error)
	UpdateRecoveryCodesAndReset(ctx context.Context, userID string, hashes []string, failedAttempts int, lastUsedAt int64) (int, error)
	ResetFailedAttempts(ctx context.Context, userID string, lastUsedAt int64) (int, error)
	IncrementFailedAttempts(ctx context.Context, userID string, attempts int) (int, error)
	LockAccount(ctx context.Context, userID string, lockedUntil int64) (int, error)
}

// userMFARepository implements UserMFARepositoryInterface
type userMFARepository struct {
	data *data.Data
}

// NewUserMFARepository creates a new user MFA repository
func NewUserMFARepository(d *data.Data) UserMFARepositoryInterface {
	return &userMFARepository{data: d}
}

// GetByUserID gets a user MFA record by user ID
func (r *userMFARepository) GetByUserID(ctx context.Context, userID string) (*ent.UserMFA, error) {
	return r.data.GetSlaveEntClient().UserMFA.Query().
		Where(userMFAEnt.UserIDEQ(userID)).
		Only(ctx)
}

// Create creates a new user MFA record
func (r *userMFARepository) Create(ctx context.Context, userID string, totpSecret string) (*ent.UserMFA, error) {
	return r.data.GetMasterEntClient().UserMFA.Create().
		SetUserID(userID).
		SetEnabled(false).
		SetTotpSecret(totpSecret).
		Save(ctx)
}

// UpdateSetup updates an existing MFA setup (reset state for re-setup)
func (r *userMFARepository) UpdateSetup(ctx context.Context, userID string, totpSecret string) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetEnabled(false).
		SetTotpSecret(totpSecret).
		ClearVerifiedAt().
		ClearRecoveryCodesGeneratedAt().
		SetRecoveryCodeHashes([]string{}).
		SetFailedAttempts(0).
		ClearLockedUntil().
		Save(ctx)
}

// Enable enables MFA for a user with recovery codes
func (r *userMFARepository) Enable(ctx context.Context, userID string, verifiedAt int64, recoveryHashes []string, recoveryGeneratedAt int64) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetEnabled(true).
		SetVerifiedAt(verifiedAt).
		SetRecoveryCodeHashes(recoveryHashes).
		SetRecoveryCodesGeneratedAt(recoveryGeneratedAt).
		SetFailedAttempts(0).
		ClearLockedUntil().
		Save(ctx)
}

// Disable disables MFA for a user
func (r *userMFARepository) Disable(ctx context.Context, userID string) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetEnabled(false).
		ClearTotpSecret().
		ClearVerifiedAt().
		SetRecoveryCodeHashes([]string{}).
		ClearRecoveryCodesGeneratedAt().
		SetFailedAttempts(0).
		ClearLockedUntil().
		Save(ctx)
}

// UpdateRecoveryCodes updates recovery code hashes
func (r *userMFARepository) UpdateRecoveryCodes(ctx context.Context, userID string, hashes []string, generatedAt int64) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetRecoveryCodeHashes(hashes).
		SetRecoveryCodesGeneratedAt(generatedAt).
		Save(ctx)
}

// UpdateRecoveryCodesAndReset updates recovery codes and resets failed attempts
func (r *userMFARepository) UpdateRecoveryCodesAndReset(ctx context.Context, userID string, hashes []string, failedAttempts int, lastUsedAt int64) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetRecoveryCodeHashes(hashes).
		SetFailedAttempts(failedAttempts).
		ClearLockedUntil().
		SetLastUsedAt(lastUsedAt).
		Save(ctx)
}

// ResetFailedAttempts resets failed attempts and updates last used time
func (r *userMFARepository) ResetFailedAttempts(ctx context.Context, userID string, lastUsedAt int64) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetFailedAttempts(0).
		ClearLockedUntil().
		SetLastUsedAt(lastUsedAt).
		Save(ctx)
}

// IncrementFailedAttempts increments failed attempts counter
func (r *userMFARepository) IncrementFailedAttempts(ctx context.Context, userID string, attempts int) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetFailedAttempts(attempts).
		Save(ctx)
}

// LockAccount locks account and resets failed attempts
func (r *userMFARepository) LockAccount(ctx context.Context, userID string, lockedUntil int64) (int, error) {
	return r.data.GetMasterEntClient().UserMFA.Update().
		Where(userMFAEnt.UserIDEQ(userID)).
		SetLockedUntil(lockedUntil).
		SetFailedAttempts(0).
		Save(ctx)
}
