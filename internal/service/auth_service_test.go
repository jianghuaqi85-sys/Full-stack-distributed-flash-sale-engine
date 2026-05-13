package service

import (
	"testing"

	"order-system/internal/pkg/db"
	"order-system/internal/repository"
)

type mockUserRepo struct {
	users map[string]*db.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*db.User)}
}

func (m *mockUserRepo) FindByID(id uint) (*db.User, error) {
	for _, u := range m.users {
		if u.ID == id {
			return u, nil
		}
	}
	return nil, nil
}

func (m *mockUserRepo) FindByUsername(username string) (*db.User, error) {
	if u, ok := m.users[username]; ok {
		return u, nil
	}
	return nil, nil
}

func (m *mockUserRepo) Create(user *db.User) error {
	if _, ok := m.users[user.Username]; ok {
		return nil
	}
	m.users[user.Username] = user
	return nil
}

var _ repository.UserRepository = (*mockUserRepo)(nil)

func TestAuthService_Register(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 3600)

	err := svc.Register(RegisterInput{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	if _, ok := repo.users["testuser"]; !ok {
		t.Error("user not created in repository")
	}
}

func TestAuthService_Login(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 3600)

	err := svc.Register(RegisterInput{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	output, err := svc.Login("testuser", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	if output.AccessToken == "" {
		t.Error("access token is empty")
	}
	if output.TokenType != "Bearer" {
		t.Errorf("token type = %q, want Bearer", output.TokenType)
	}
	if output.ExpiresIn != 3600 {
		t.Errorf("expires_in = %d, want 3600", output.ExpiresIn)
	}
}

func TestAuthService_Login_InvalidCredentials(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 3600)

	_, err := svc.Login("nonexistent", "password123")
	if err == nil {
		t.Error("Login() expected error for nonexistent user")
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	svc := NewAuthService(repo, "test-secret", 3600)

	err := svc.Register(RegisterInput{
		Username: "testuser",
		Password: "password123",
		Email:    "test@example.com",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	_, err = svc.Login("testuser", "wrongpassword")
	if err == nil {
		t.Error("Login() expected error for wrong password")
	}
}
