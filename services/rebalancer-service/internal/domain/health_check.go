package domain

import "context"

// HealthCheckService defines the contract for performing application health checks.
type HealthCheckService interface {
	HealthCheck(ctx context.Context) error
}
