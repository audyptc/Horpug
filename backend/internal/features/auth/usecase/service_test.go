package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	authdomain "apihorpug/internal/features/auth/domain"
	userdomain "apihorpug/internal/features/user/domain"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type fakeUserRepo struct {
	user        userdomain.User
	newPassword string
}

func (f *fakeUserRepo) FindByLogin(context.Context, string) (userdomain.User, error) {
	return f.user, nil
}

func (f *fakeUserRepo) GetByID(context.Context, uuid.UUID) (userdomain.User, error) {
	return f.user, nil
}

func (f *fakeUserRepo) UpdatePassword(_ context.Context, _ uuid.UUID, hashed string) error {
	f.newPassword = hashed
	return nil
}

type fakeTokenRepo struct {
	revokedOthersKeep *string
}

func (f *fakeTokenRepo) SaveRefreshToken(context.Context, authdomain.RefreshToken) error { return nil }
func (f *fakeTokenRepo) FindRefreshToken(context.Context, string) (authdomain.RefreshToken, error) {
	return authdomain.RefreshToken{}, authdomain.ErrRefreshTokenInvalid
}
func (f *fakeTokenRepo) RevokeRefreshToken(context.Context, uuid.UUID) error { return nil }
func (f *fakeTokenRepo) RevokeOtherRefreshTokens(_ context.Context, _ uuid.UUID, keepHash string) error {
	f.revokedOthersKeep = &keepHash
	return nil
}

func newChangePasswordFixture(t *testing.T) (*Service, *fakeUserRepo, *fakeTokenRepo) {
	t.Helper()
	hashed, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	users := &fakeUserRepo{user: userdomain.User{ID: uuid.New(), Password: string(hashed), IsActive: true}}
	tokens := &fakeTokenRepo{}
	return New(users, tokens, nil, "secret", time.Minute, time.Hour), users, tokens
}

func TestChangePasswordRejects(t *testing.T) {
	cases := []struct {
		name            string
		current, newPwd string
		want            error
	}{
		{"wrong current", "nope", "new-password", authdomain.ErrWrongPassword},
		{"too short", "old-password", "short", authdomain.ErrWeakPassword},
		{"blank", "old-password", "          ", authdomain.ErrWeakPassword},
		{"same as old", "old-password", "old-password", authdomain.ErrSamePassword},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			svc, users, tokens := newChangePasswordFixture(t)
			err := svc.ChangePassword(context.Background(), users.user.ID, tc.current, tc.newPwd, "rt", "")
			if !errors.Is(err, tc.want) {
				t.Fatalf("got %v, want %v", err, tc.want)
			}
			if users.newPassword != "" || tokens.revokedOthersKeep != nil {
				t.Fatal("nothing should change on a rejected request")
			}
		})
	}
}

func TestChangePasswordKeepsCurrentDevice(t *testing.T) {
	svc, users, tokens := newChangePasswordFixture(t)
	if err := svc.ChangePassword(context.Background(), users.user.ID, "old-password", "new-password", "my-refresh", ""); err != nil {
		t.Fatal(err)
	}
	if bcrypt.CompareHashAndPassword([]byte(users.newPassword), []byte("new-password")) != nil {
		t.Fatal("stored hash does not match the new password")
	}
	if tokens.revokedOthersKeep == nil || *tokens.revokedOthersKeep != hashToken("my-refresh") {
		t.Fatal("other sessions should be revoked, keeping the current refresh token")
	}
}
