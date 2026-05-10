package user

import (
	"cfk/pkg/event"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/samber/mo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUsername    = errors.New("invalid username")
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidRole        = errors.New("invalid role")
	ErrNotFound           = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

var validRoles = map[string]bool{
	"user":    true,
	"premium": true,
	"admin":   true,
}

type Service struct {
	repo     Repository
	eventBus *event.Bus
}

func NewService(repo Repository, eventBus *event.Bus) *Service {
	return &Service{
		repo:     repo,
		eventBus: eventBus,
	}
}

func (s *Service) RegisterUser(tenantID, username, email, password, firstName, lastName, phone, role string) mo.Result[User] {
	if username == "" {
		return mo.Err[User](ErrInvalidUsername)
	}
	if email == "" {
		return mo.Err[User](ErrInvalidEmail)
	}
	if password == "" {
		return mo.Err[User](ErrInvalidPassword)
	}
	if !validRoles[role] {
		return mo.Err[User](ErrInvalidRole)
	}

	if _, ok := s.repo.FindByEmail(tenantID, email).Get(); ok {
		return mo.Err[User](ErrEmailAlreadyExists)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return mo.Err[User](err)
	}

	id := uuid.New().String()
	now := time.Now()

	eventPayload := map[string]interface{}{
		"id":              id,
		"tenant_id":       tenantID,
		"username":        username,
		"email":           email,
		"hashed_password": string(hashedPassword),
		"first_name":      firstName,
		"last_name":       lastName,
		"phone":           phone,
		"role":            role,
		"is_active":       true,
		"created_at":      now,
		"updated_at":      now,
	}

	evt := event.Event{
		AggregateType: "user",
		AggregateID:   id,
		EventType:     EventRegistered,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[User](r.Error())
	}

	return mo.Ok(User{
		ID:             id,
		TenantID:       tenantID,
		Username:       username,
		HashedPassword: string(hashedPassword),
		FirstName:      firstName,
		LastName:       lastName,
		Phone:          phone,
		Email:          email,
		Role:           role,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	})
}

func (s *Service) ActivateUser(id string) mo.Result[User] {
	return s.updateUserStatus(id, true, EventActivated)
}

func (s *Service) DeactivateUser(id string) mo.Result[User] {
	return s.updateUserStatus(id, false, EventDeactivated)
}

func (s *Service) ChangeRole(id, role string) mo.Result[User] {
	if !validRoles[role] {
		return mo.Err[User](ErrInvalidRole)
	}

	userOpt := s.repo.FindByID(id)
	user, ok := userOpt.Get()
	if !ok {
		return mo.Err[User](ErrNotFound)
	}

	now := time.Now()
	eventPayload := map[string]interface{}{
		"id":         id,
		"role":       role,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "user",
		AggregateID:   id,
		EventType:     EventRoleChanged,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[User](r.Error())
	}

	user.Role = role
	user.UpdatedAt = now
	return mo.Ok(user)
}

func (s *Service) UpdateProfile(id, firstName, lastName, phone string) mo.Result[User] {
	userOpt := s.repo.FindByID(id)
	user, ok := userOpt.Get()
	if !ok {
		return mo.Err[User](ErrNotFound)
	}

	now := time.Now()
	eventPayload := map[string]interface{}{
		"id":         id,
		"first_name": firstName,
		"last_name":  lastName,
		"phone":      phone,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "user",
		AggregateID:   id,
		EventType:     EventProfileUpdated,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[User](r.Error())
	}

	user.FirstName = firstName
	user.LastName = lastName
	user.Phone = phone
	user.UpdatedAt = now
	return mo.Ok(user)
}

func (s *Service) GetUserByEmail(tenantID, email string) mo.Result[User] {
	userOpt := s.repo.FindByEmail(tenantID, email)
	user, ok := userOpt.Get()
	if !ok {
		return mo.Err[User](ErrNotFound)
	}
	return mo.Ok(user)
}

func (s *Service) updateUserStatus(id string, isActive bool, eventType string) mo.Result[User] {
	userOpt := s.repo.FindByID(id)
	user, ok := userOpt.Get()
	if !ok {
		return mo.Err[User](ErrNotFound)
	}

	now := time.Now()
	eventPayload := map[string]interface{}{
		"id":         id,
		"is_active":  isActive,
		"updated_at": now,
	}

	evt := event.Event{
		AggregateType: "user",
		AggregateID:   id,
		EventType:     eventType,
		Version:       1,
		Payload:       eventPayload,
		Metadata: map[string]interface{}{
			"timestamp": now,
		},
	}

	if r := s.eventBus.Publish(evt); r.IsError() {
		return mo.Err[User](r.Error())
	}

	user.IsActive = isActive
	user.UpdatedAt = now
	return mo.Ok(user)
}
