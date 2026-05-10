package saga

import (
	"encoding/json"
	"time"

	"github.com/samber/mo"
	"gorm.io/gorm"
)

type SagaInstanceProjection struct {
	ID          string    `gorm:"type:uuid;primaryKey"`
	SagaName    string    `gorm:"type:varchar(100);not null;index:idx_saga_instances_name"`
	State       string    `gorm:"type:varchar(50);not null;index:idx_saga_instances_state"`
	CurrentStep int       `gorm:"not null"`
	Payload     string    `gorm:"type:jsonb"`
	Error       string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt   time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (SagaInstanceProjection) TableName() string {
	return "saga_instances"
}

type GORMStore struct {
	db *gorm.DB
}

func NewGORMStore(db *gorm.DB) *GORMStore {
	return &GORMStore{db: db}
}

func (s *GORMStore) Save(instance *Instance) mo.Result[struct{}] {
	proj := toProjection(instance)
	if err := s.db.Create(&proj).Error; err != nil {
		return mo.Err[struct{}](err)
	}
	return OkStep()
}

func (s *GORMStore) FindByID(id string) mo.Option[*Instance] {
	var proj SagaInstanceProjection
	if err := s.db.Where("id = ?", id).First(&proj).Error; err != nil {
		return mo.None[*Instance]()
	}
	inst, err := toInstance(proj)
	if err != nil {
		return mo.None[*Instance]()
	}
	return mo.Some(inst)
}

func (s *GORMStore) FindByState(state InstanceState) mo.Result[[]*Instance] {
	var projs []SagaInstanceProjection
	if err := s.db.Where("state = ?", string(state)).Find(&projs).Error; err != nil {
		return mo.Err[[]*Instance](err)
	}
	result := make([]*Instance, 0, len(projs))
	for _, proj := range projs {
		inst, err := toInstance(proj)
		if err != nil {
			return mo.Err[[]*Instance](err)
		}
		result = append(result, inst)
	}
	return mo.Ok(result)
}

func (s *GORMStore) Update(instance *Instance) mo.Result[struct{}] {
	instance.UpdatedAt = time.Now()
	proj := toProjection(instance)
	proj.UpdatedAt = instance.UpdatedAt
	if err := s.db.Where("id = ?", instance.ID).Updates(map[string]interface{}{
		"state":        proj.State,
		"current_step": proj.CurrentStep,
		"payload":      proj.Payload,
		"error":        proj.Error,
		"updated_at":   proj.UpdatedAt,
	}).Error; err != nil {
		return mo.Err[struct{}](err)
	}
	return OkStep()
}

func toProjection(inst *Instance) SagaInstanceProjection {
	payloadJSON, _ := json.Marshal(inst.Payload)
	return SagaInstanceProjection{
		ID:          inst.ID,
		SagaName:    inst.SagaName,
		State:       string(inst.State),
		CurrentStep: inst.CurrentStep,
		Payload:     string(payloadJSON),
		Error:       inst.Error,
		CreatedAt:   inst.CreatedAt,
		UpdatedAt:   inst.UpdatedAt,
	}
}

func toInstance(proj SagaInstanceProjection) (*Instance, error) {
	var payload map[string]interface{}
	if proj.Payload != "" {
		if err := json.Unmarshal([]byte(proj.Payload), &payload); err != nil {
			return nil, err
		}
	}
	return &Instance{
		ID:          proj.ID,
		SagaName:    proj.SagaName,
		State:       InstanceState(proj.State),
		CurrentStep: proj.CurrentStep,
		Payload:     payload,
		Error:       proj.Error,
		CreatedAt:   proj.CreatedAt,
		UpdatedAt:   proj.UpdatedAt,
	}, nil
}
