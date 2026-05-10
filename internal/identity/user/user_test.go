package user

import (
	"cfk/pkg/event"
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/samber/mo"
)

type mockUserRepo struct {
	users   map[string]User
	byEmail map[string]User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{
		users:   make(map[string]User),
		byEmail: make(map[string]User),
	}
}

func (r *mockUserRepo) AppendEvent(eventType string, aggregateID string, payload any, metadata map[string]interface{}) error {
	return nil
}

func (r *mockUserRepo) FindByID(id string) mo.Option[User] {
	if u, ok := r.users[id]; ok {
		return mo.Some(u)
	}
	return mo.None[User]()
}

func (r *mockUserRepo) FindByEmail(tenantID, email string) mo.Option[User] {
	key := tenantID + ":" + email
	if u, ok := r.byEmail[key]; ok {
		return mo.Some(u)
	}
	return mo.None[User]()
}

func setupTestBus(t *testing.T) (*event.Bus, *[]event.Event) {
	t.Helper()
	recorded := &[]event.Event{}
	bus := event.NewBus(event.WithWorkerPool(1), event.WithBufferSize(64))
	bus.Subscribe("*", func(evt event.Event) mo.Result[struct{}] {
		*recorded = append(*recorded, evt)
		return event.OkHandle()
	})
	ctx := context.Background()
	bus.Start(ctx)
	t.Cleanup(func() {
		bus.Stop()
	})
	return bus, recorded
}

func waitForEvents() {
	time.Sleep(50 * time.Millisecond)
}

func TestRegisterUser_Valid(t *testing.T) {
	repo := newMockUserRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	u, err := svc.RegisterUser("t-1", "testuser", "test@example.com", "password123", "First", "Last", "0801234567", "user").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", u.Username)
	}
	if u.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", u.Email)
	}
	if u.Role != "user" {
		t.Errorf("expected role 'user', got '%s'", u.Role)
	}
	if !u.IsActive {
		t.Error("expected IsActive to be true")
	}
	if u.TenantID != "t-1" {
		t.Errorf("expected tenant ID 't-1', got '%s'", u.TenantID)
	}
	if u.HashedPassword == "" {
		t.Error("expected hashed password to be set")
	}
	if u.ID == "" {
		t.Error("expected non-empty ID")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventRegistered {
		t.Errorf("expected event type '%s', got '%s'", EventRegistered, (*recorded)[0].EventType)
	}
	if (*recorded)[0].AggregateType != "user" {
		t.Errorf("expected aggregate type 'user', got '%s'", (*recorded)[0].AggregateType)
	}
}

func TestRegisterUser_AllRoles(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	roles := []string{"user", "premium", "admin"}
	for _, role := range roles {
		_, err := svc.RegisterUser("t-1", "user-"+role, role+"@example.com", "pass123", "F", "L", "", role).Get()
		if err != nil {
			t.Errorf("role '%s': expected no error, got %v", role, err)
		}
	}
}

func TestRegisterUser_EmptyUsername(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RegisterUser("t-1", "", "test@example.com", "password123", "F", "L", "", "user").Get()
	if !errors.Is(err, ErrInvalidUsername) {
		t.Errorf("expected ErrInvalidUsername, got %v", err)
	}
}

func TestRegisterUser_EmptyEmail(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RegisterUser("t-1", "testuser", "", "password123", "F", "L", "", "user").Get()
	if !errors.Is(err, ErrInvalidEmail) {
		t.Errorf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegisterUser_EmptyPassword(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RegisterUser("t-1", "testuser", "test@example.com", "", "F", "L", "", "user").Get()
	if !errors.Is(err, ErrInvalidPassword) {
		t.Errorf("expected ErrInvalidPassword, got %v", err)
	}
}

func TestRegisterUser_InvalidRole(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.RegisterUser("t-1", "testuser", "test@example.com", "password123", "F", "L", "", "superadmin").Get()
	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestRegisterUser_DuplicateEmail(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-existing",
		TenantID: "t-1",
		Username: "existing",
		Email:    "duplicate@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo.users["u-existing"] = existing
	repo.byEmail["t-1:duplicate@example.com"] = existing

	_, err := svc.RegisterUser("t-1", "newuser", "duplicate@example.com", "password123", "F", "L", "", "user").Get()
	if !errors.Is(err, ErrEmailAlreadyExists) {
		t.Errorf("expected ErrEmailAlreadyExists, got %v", err)
	}
}

func TestRegisterUser_SameEmailDifferentTenant(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-existing",
		TenantID: "t-1",
		Username: "existing",
		Email:    "same@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo.users["u-existing"] = existing
	repo.byEmail["t-1:same@example.com"] = existing

	_, err := svc.RegisterUser("t-2", "newuser", "same@example.com", "password123", "F", "L", "", "user").Get()
	if err != nil {
		t.Errorf("same email in different tenant should be allowed, got %v", err)
	}
}

func TestGetUserByEmail_Found(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-1",
		TenantID: "t-1",
		Username: "testuser",
		Email:    "find@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo.users["u-1"] = existing
	repo.byEmail["t-1:find@example.com"] = existing

	u, err := svc.GetUserByEmail("t-1", "find@example.com").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Email != "find@example.com" {
		t.Errorf("expected email 'find@example.com', got '%s'", u.Email)
	}
}

func TestGetUserByEmail_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.GetUserByEmail("t-1", "nonexistent@example.com").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestActivateUser(t *testing.T) {
	repo := newMockUserRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-1",
		TenantID: "t-1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		IsActive: false,
	}
	repo.users["u-1"] = existing

	u, err := svc.ActivateUser("u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !u.IsActive {
		t.Error("expected IsActive to be true after activation")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventActivated {
		t.Errorf("expected event type '%s', got '%s'", EventActivated, (*recorded)[0].EventType)
	}
}

func TestActivateUser_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ActivateUser("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDeactivateUser(t *testing.T) {
	repo := newMockUserRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-1",
		TenantID: "t-1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo.users["u-1"] = existing

	u, err := svc.DeactivateUser("u-1").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.IsActive {
		t.Error("expected IsActive to be false after deactivation")
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventDeactivated {
		t.Errorf("expected event type '%s', got '%s'", EventDeactivated, (*recorded)[0].EventType)
	}
}

func TestDeactivateUser_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.DeactivateUser("nonexistent").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestChangeRole(t *testing.T) {
	repo := newMockUserRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-1",
		TenantID: "t-1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo.users["u-1"] = existing

	u, err := svc.ChangeRole("u-1", "admin").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.Role != "admin" {
		t.Errorf("expected role 'admin', got '%s'", u.Role)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventRoleChanged {
		t.Errorf("expected event type '%s', got '%s'", EventRoleChanged, (*recorded)[0].EventType)
	}
}

func TestChangeRole_InvalidRole(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:       "u-1",
		TenantID: "t-1",
		Username: "testuser",
		Email:    "test@example.com",
		Role:     "user",
		IsActive: true,
	}
	repo.users["u-1"] = existing

	_, err := svc.ChangeRole("u-1", "superadmin").Get()
	if !errors.Is(err, ErrInvalidRole) {
		t.Errorf("expected ErrInvalidRole, got %v", err)
	}
}

func TestChangeRole_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.ChangeRole("nonexistent", "admin").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestUpdateProfile(t *testing.T) {
	repo := newMockUserRepo()
	bus, recorded := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:        "u-1",
		TenantID:  "t-1",
		Username:  "testuser",
		Email:     "test@example.com",
		Role:      "user",
		IsActive:  true,
		FirstName: "Old",
		LastName:  "Name",
		Phone:    "0800000000",
	}
	repo.users["u-1"] = existing

	u, err := svc.UpdateProfile("u-1", "New", "Name", "0812345678").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if u.FirstName != "New" {
		t.Errorf("expected first name 'New', got '%s'", u.FirstName)
	}
	if u.Phone != "0812345678" {
		t.Errorf("expected phone '0812345678', got '%s'", u.Phone)
	}

	waitForEvents()
	if len(*recorded) != 1 {
		t.Fatalf("expected 1 event, got %d", len(*recorded))
	}
	if (*recorded)[0].EventType != EventProfileUpdated {
		t.Errorf("expected event type '%s', got '%s'", EventProfileUpdated, (*recorded)[0].EventType)
	}
}

func TestUpdateProfile_NotFound(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.UpdateProfile("nonexistent", "New", "Name", "0812345678").Get()
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestLogin_Valid(t *testing.T) {
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")

	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	registered, _ := svc.RegisterUser("t-1", "testuser", "login@example.com", "password123", "F", "L", "", "user").Get()
	repo.users[registered.ID] = registered
	repo.byEmail["t-1:login@example.com"] = registered

	tokenResp, err := svc.Login("t-1", "login@example.com", "password123").Get()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if tokenResp.Token == "" {
		t.Error("expected non-empty token")
	}
	if tokenResp.ExpiresAt == 0 {
		t.Error("expected non-zero expires_at")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	registered, _ := svc.RegisterUser("t-1", "testuser", "login2@example.com", "password123", "F", "L", "", "user").Get()
	repo.users[registered.ID] = registered
	repo.byEmail["t-1:login2@example.com"] = registered

	_, err := svc.Login("t-1", "login2@example.com", "wrongpassword").Get()
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	_, err := svc.Login("t-1", "nonexistent@example.com", "password123").Get()
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestLogin_DeactivatedUser(t *testing.T) {
	repo := newMockUserRepo()
	bus, _ := setupTestBus(t)
	svc := NewService(repo, bus)

	existing := User{
		ID:             "u-1",
		TenantID:       "t-1",
		Email:          "deactivated@example.com",
		Role:           "user",
		IsActive:       false,
		HashedPassword: "$2a$10$dummy",
	}
	repo.users["u-1"] = existing
	repo.byEmail["t-1:deactivated@example.com"] = existing

	_, err := svc.Login("t-1", "deactivated@example.com", "password123").Get()
	if !errors.Is(err, ErrUserDeactivated) {
		t.Errorf("expected ErrUserDeactivated, got %v", err)
	}
}
