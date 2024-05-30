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

package sophrosyne

// PerformScanRequest is the parameter payload expected to be provided when calling PerformScan via the API.
//
// Fields "image" and "text" are mutually exclusive.
type PerformScanRequest struct {
	// Name of the profile to use. If empty, the default profile for the user making the request will be used.
	Profile string `json:"profile"`
	// Text to scan.
	Text string `json:"text"`
	// Base64-encoded image to scan.
	Image string `json:"image",validate:"base64"`
}

// Validate the PerformScanRequest struct. The image and text fields are mutually exclusive, and
// that kind of validation check isn't supported by go-validator/v10. Therefore, we have to do it
// manually by implementing the Validate method.
func (p PerformScanRequest) Validate(interface{}) error {
	if p.Text == "" && p.Image == "" {
		return NewValidationError("no text or image specified")
	}
	return nil
}
