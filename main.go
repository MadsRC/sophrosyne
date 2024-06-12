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

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type OtelOutput string

const (
	OtelOutputStdout OtelOutput = "stdout"
	OtelOutputHTTP   OtelOutput = "http"
)

type HttpService interface {
	Start() error
}

type Validator interface {
	Validate(interface{}) error
}

func ExtractUser(ctx context.Context) *User {
	v := ctx.Value(UserContextKey{})
	u, ok := v.(*User)
	if ok {
		return u
	}
	return nil
}

type MetricService interface {
	RecordPanic(ctx context.Context)
}

type Span interface {
	End()
}

type TracingService interface {
	StartSpan(ctx context.Context, name string) (context.Context, Span)
	GetTraceID(ctx context.Context) string
	NewHTTPHandler(route string, h http.Handler) http.Handler
	WithRouteTag(route string, h http.Handler) http.Handler
}

func NewToken(source io.Reader) ([]byte, error) {
	b := make([]byte, 64)
	_, err := source.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// ProtectToken applies a Keyed-Hash Message Authentication Code (HMAC) to the
// token using the site key, salt and SHA-256.
//
// If, for any reason, the HMAC fails, the function will panic.
func ProtectToken(token []byte, config *Config) []byte {
	h := hmac.New(sha256.New, config.Security.SiteKey)
	n, err := h.Write(token)
	if err != nil {
		panic(err)
	}
	if n != len(token) {
		panic(fmt.Errorf("failed to write all bytes (token) to HMAC"))
	}
	n, err = h.Write(config.Security.Salt)
	if err != nil {
		panic(err)
	}
	if n != len(config.Security.Salt) {
		panic(fmt.Errorf("failed to write all bytes (salt) to HMAC"))
	}

	var out []byte
	out = h.Sum(out)
	return out
}

var TimeFormatInResponse = time.RFC3339

var xidRegex *regexp.Regexp = regexp.MustCompile("^[0-9a-v]{20}$")

func IsValidXID(s string) bool {
	return xidRegex.MatchString(s)
}

const DatabaseCursorSeparator = "::"

type DatabaseCursor struct {
	OwnerID  string
	Position string
}

func NewDatabaseCursor(ownerID, position string) *DatabaseCursor {
	return &DatabaseCursor{
		OwnerID:  ownerID,
		Position: position,
	}
}

func (c DatabaseCursor) String() string {
	if c.OwnerID == "" || c.Position == "" {
		return ""
	}
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s%s%s", c.OwnerID, DatabaseCursorSeparator, c.Position)))
}

func (c *DatabaseCursor) Reset() {
	c.Position = ""
}

func (c *DatabaseCursor) Advance(position string) {
	c.Position = position
}

func (c *DatabaseCursor) LogValue() slog.Value {
	return slog.GroupValue(slog.String("owner_id", c.OwnerID), slog.String("last_read", c.Position))
}

var errInvalidCursor = errors.New("invalid cursor")

// DecodeDatabaseCursorWithOwner decodes a cursor string into a DatabaseCursor and verifies that the
// ownerID matches.
//
// It is sugar around [DecodeDatabaseCursor] and carries with it the same caveats as
// [DecodeDatabaseCursor] does.
//
// In addition to the possible errors returned by [DecodeDatabaseCursor], this function returns
// an error if the ownerID does not match.
func DecodeDatabaseCursorWithOwner(s string, ownerID string) (*DatabaseCursor, error) {
	cursor, err := DecodeDatabaseCursor(s)
	if err != nil {
		return nil, err
	}
	if cursor.OwnerID != ownerID {
		return nil, errInvalidCursor
	}
	return cursor, nil
}

// DecodeDatabaseCursor decodes a cursor string into a DatabaseCursor.
//
// A database cursor is a base64 encoded string of the form:
//
//	<ownerID>::<lastReadPosition>
//
// The ownerID and lastReadPositions must be textual representations of an XID.
func DecodeDatabaseCursor(s string) (*DatabaseCursor, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(string(b), DatabaseCursorSeparator)
	if len(parts) != 2 {
		return nil, errInvalidCursor
	}

	if !IsValidXID(parts[0]) || !IsValidXID(parts[1]) {
		return nil, errInvalidCursor
	}

	return &DatabaseCursor{
		OwnerID:  parts[0],
		Position: parts[1],
	}, nil
}

type AuthorizationProvider interface {
	IsAuthorized(ctx context.Context, req AuthorizationRequest) bool
}

type AuthorizationEntity interface {
	EntityType() string
	EntityID() string
}

type AuthorizationAction string

func (a AuthorizationAction) EntityType() string {
	return "Action"
}

func (a AuthorizationAction) EntityID() string {
	return string(a)
}

type AuthorizationRequest struct {
	Principal AuthorizationEntity
	Action    AuthorizationEntity
	Resource  AuthorizationEntity
	Context   map[string]interface{}
}

type RPCServer interface {
	HandleRPCRequest(ctx context.Context, req []byte) ([]byte, error)
}

type HealthCheckService interface {
	UnauthenticatedHealthcheck(ctx context.Context) bool
	AuthenticatedHealthcheck(ctx context.Context) ([]byte, error)
}

type HealthChecker interface {
	Health(ctx context.Context) (bool, []byte)
}
