package healthchecker

import (
	"context"
	"github.com/madsrc/sophrosyne"
)

type HealthCheckService struct {
	services []sophrosyne.HealthChecker
}

func NewHealthcheckService(services []sophrosyne.HealthChecker) (*HealthCheckService, error) {
	return &HealthCheckService{
		services: services,
	}, nil
}

func (h HealthCheckService) UnauthenticatedHealthcheck(ctx context.Context) bool {
	for _, service := range h.services {
		ok, _ := service.Health(ctx)
		if !ok {
			return false
		}
	}
	return true
}

func (h HealthCheckService) AuthenticatedHealthcheck(ctx context.Context) ([]byte, error) {
	//TODO implement me
	panic("implement me")
}
