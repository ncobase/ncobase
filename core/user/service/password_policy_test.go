package service

import (
	"context"
	"strings"
	"testing"

	ent "ncobase/core/user/data/ent"
	"ncobase/core/user/structs"

	"github.com/ncobase/ncore/security/crypto"
	"github.com/ncobase/ncore/types"
)

func TestValidatePasswordPolicyAcceptsDefaultPolicy(t *testing.T) {
	if err := ValidatePasswordPolicy(context.Background(), nil, "StrongPass1"); err != nil {
		t.Fatalf("expected strong password to pass, got %v", err)
	}
}

func TestValidatePasswordPolicyRejectsWeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     string
	}{
		{name: "too short", password: "Aa1", want: "at least 8"},
		{name: "missing uppercase", password: "strongpass1", want: "uppercase"},
		{name: "missing lowercase", password: "STRONGPASS1", want: "lowercase"},
		{name: "missing number", password: "StrongPass", want: "number"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordPolicy(context.Background(), nil, tt.password)
			if err == nil {
				t.Fatal("expected password policy error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %q", tt.want, err.Error())
			}
		})
	}
}

func TestSetPasswordByIDHashesAcceptedPassword(t *testing.T) {
	repo := &passwordSetRepository{}
	svc := &userService{user: repo}

	if err := svc.SetPasswordByID(context.Background(), "user-id", "StrongPass1"); err != nil {
		t.Fatalf("expected password set to pass, got %v", err)
	}
	if repo.userID != "user-id" {
		t.Fatalf("expected repository call for user-id, got %q", repo.userID)
	}
	if repo.hashedPassword == "" || repo.hashedPassword == "StrongPass1" {
		t.Fatalf("expected hashed password, got %q", repo.hashedPassword)
	}
	if !crypto.ComparePassword(repo.hashedPassword, "StrongPass1") {
		t.Fatal("expected stored hash to match original password")
	}
}

func TestSetPasswordByIDRejectsWeakPasswordBeforeRepositoryCall(t *testing.T) {
	repo := &passwordSetRepository{}
	svc := &userService{user: repo}

	err := svc.SetPasswordByID(context.Background(), "user-id", "weak")
	if err == nil {
		t.Fatal("expected weak password to be rejected")
	}
	if repo.userID != "" || repo.hashedPassword != "" {
		t.Fatal("expected repository to remain untouched for weak password")
	}
}

type passwordSetRepository struct {
	userID         string
	hashedPassword string
}

func (r *passwordSetRepository) Create(context.Context, *structs.UserBody) (*ent.User, error) {
	return nil, nil
}

func (r *passwordSetRepository) Update(context.Context, string, types.JSON) (*ent.User, error) {
	return nil, nil
}

func (r *passwordSetRepository) GetByID(context.Context, string) (*ent.User, error) {
	return nil, nil
}

func (r *passwordSetRepository) Delete(context.Context, string) error {
	return nil
}

func (r *passwordSetRepository) List(context.Context, *structs.ListUserParams) ([]*ent.User, error) {
	return nil, nil
}

func (r *passwordSetRepository) Find(context.Context, *structs.FindUser) (*ent.User, error) {
	return nil, nil
}

func (r *passwordSetRepository) FindUser(context.Context, *structs.FindUser) (*ent.User, error) {
	return nil, nil
}

func (r *passwordSetRepository) UpdatePassword(context.Context, *structs.UserPassword) error {
	return nil
}

func (r *passwordSetRepository) UpdatePasswordByID(_ context.Context, userID, hashedPassword string) error {
	r.userID = userID
	r.hashedPassword = hashedPassword
	return nil
}

func (r *passwordSetRepository) CountX(context.Context, *structs.ListUserParams) int {
	return 0
}
