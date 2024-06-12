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

package grpc

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/madsrc/sophrosyne"
)

func Test_newGetProfileResponseFromProfile(t *testing.T) {
	now := time.Now().UTC()
	checks := []string{"testCheck1", "testCheck2"}
	input := sophrosyne.Profile{
		ID:   "someID",
		Name: "test",
		Checks: []sophrosyne.Check{
			{
				Name: "testCheck1",
			},
			{
				Name: "testCheck2",
			},
		},
		CreatedAt: now,
		UpdatedAt: now,
		DeletedAt: &now,
	}

	output := newGetProfileResponseFromProfile(&input)

	require.Equal(t, input.Name, output.Name)
	require.Equal(t, checks, output.Checks)
	require.Equal(t, input.CreatedAt, output.CreatedAt.AsTime())
	require.Equal(t, input.UpdatedAt, output.UpdatedAt.AsTime())
	delAt := output.DeletedAt.AsTime()
	require.Equal(t, input.DeletedAt, &delAt)
}
