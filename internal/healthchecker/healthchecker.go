// Sophrosyne
//   Copyright (C) 2024  Mads R. Havmand
//
// This program is free software: you can redistribute it and/or modify
//   it under the terms of the GNU Affero General Public License as published by
//   the Free Software Foundation, either version 3 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU Affero General Public License for more details.
//
//   You should have received a copy of the GNU Affero General Public License
//   along with this program.  If not, see <http://www.gnu.org/licenses/>.

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
