package service

import (
	"context"

	"github.com/SteeperMold/Emergency-Notification-System/sender-service/internal/domain"
)

// HealthCheckService provides methods to check the health of system dependencies.
// It verifies connectivity to the database and Kafka.
type HealthCheckService struct {
	db domain.DBConn
	kf domain.KafkaFactory
}

// NewHealthCheckService creates and returns a new HealthCheckService.
func NewHealthCheckService(db domain.DBConn, kf domain.KafkaFactory) *HealthCheckService {
	return &HealthCheckService{
		db: db,
		kf: kf,
	}
}

// HealthCheck runs health checks for the database and Kafka.
// It returns an error if any of the checks fail.
func (s *HealthCheckService) HealthCheck(ctx context.Context) error {
	err := s.db.Ping(ctx)
	if err != nil {
		return err
	}

	err = s.kf.Ping(ctx)
	if err != nil {
		return err
	}

	return nil
}
