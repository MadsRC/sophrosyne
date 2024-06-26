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

//go:build integration

package integration

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/testcontainers/testcontainers-go"
)

type testEnv struct {
	t              *testing.T
	database       *postgres.PostgresContainer
	api            testcontainers.Container
	dummycheck     testcontainers.Container
	network        *testcontainers.DockerNetwork
	rootToken      string
	httpClient     *http.Client
	endpoint       string
	healthEndpoint *url.URL
	rpcEndpoint    *url.URL
}

func (te testEnv) Close(ctx context.Context) {
	var err error
	if te.database != nil {
		err = errors.Join(err, te.database.Terminate(ctx))
	}
	if te.api != nil {
		err = errors.Join(err, te.api.Terminate(ctx))
	}
	if te.dummycheck != nil {
		err = errors.Join(err, te.dummycheck.Terminate(ctx))
	}
	if te.network != nil {
		err = errors.Join(err, te.network.Remove(ctx))
	}

	require.NoError(te.t, err, "could not clean up test environment")
}

func doAuthenticatedRequest(t *testing.T, te *testEnv, method string, body []byte) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequest(method, te.rpcEndpoint.String(), bytes.NewBuffer(body))
	require.NoError(t, err, "could not create HTTP request")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", te.rootToken))
	return te.httpClient.Do(req)
}

func compareResponse(t *testing.T, expected []byte, response *http.Response) {
	t.Helper()
	require.NotNil(t, response, "response is nil")
	require.Equalf(t, http.StatusOK, response.StatusCode, "expected status code %d, got %d", http.StatusOK, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err, "could not read response body")
	require.JSONEq(t, string(expected), string(body), "expected response body to be different")
}

func setupEnv(ctx context.Context, t *testing.T) testEnv {
	t.Helper()
	te := testEnv{t: t}

	nw, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		network.WithDriver("bridge"),
	)
	require.NoError(t, err)
	te.network = nw

	dbName := "users"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgres:16-alpine"),
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
		network.WithNetwork(nil, nw),
	)
	if err != nil {
		t.Fatalf("failed to start container: %s", err)
	}

	te.database = postgresContainer

	dummycheckReq := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:       "../../",
			Dockerfile:    "tests/integration/dummycheck.Dockerfile",
			PrintBuildLog: true,
			KeepImage:     true,
		},
		ExposedPorts: []string{"11432/tcp"},
		Networks:     []string{nw.Name},
		WaitingFor:   wait.ForLog("starting server on port 11432"),
	}

	dummycheck, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: dummycheckReq,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Could not start dummycheck: %s", err)
	}
	te.dummycheck = dummycheck

	_, err = postgresContainer.Endpoint(ctx, "")
	require.NoError(t, err)
	pgIP, err := postgresContainer.ContainerIP(ctx)
	require.NoError(t, err)

	siteKey := make([]byte, 64)
	salt := make([]byte, 32)
	_, err = rand.Read(siteKey)
	require.NoError(t, err)
	_, err = rand.Read(salt)
	require.NoError(t, err)

	siteKeyContent := bytes.NewReader(siteKey)
	saltContent := bytes.NewReader([]byte(salt))
	r := bytes.NewReader([]byte(fmt.Sprintf(`database:
  host: %s
  port: %s
  user: user
  password: password
  name: users
logging:
  level: debug`, pgIP, "5432")))

	img := "ghcr.io/madsrc/sophrosyne:latest"
	if os.Getenv("sophrosyne_test_image") != "" {
		img = os.Getenv("sophrosyne_test_image")
	}

	req := testcontainers.ContainerRequest{
		Image:        img,
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForLog("Starting server"),
		Cmd:          []string{"--secretfiles", "/security.salt,/security.siteKey", "run"},
		Files: []testcontainers.ContainerFile{
			{
				Reader:            r,
				ContainerFilePath: "/config.yaml",
				FileMode:          0644,
			},
			{
				Reader:            saltContent,
				ContainerFilePath: "/security.salt",
				FileMode:          0644,
			},
			{
				Reader:            siteKeyContent,
				ContainerFilePath: "/security.siteKey",
				FileMode:          0644,
			},
		},
		Networks: []string{nw.Name},
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Opts:      []testcontainers.LogProductionOption{testcontainers.WithLogProductionTimeout(10 * time.Second)},
			Consumers: []testcontainers.LogConsumer{ensureJSON{t: te.t}},
		},
	}
	sophC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Could not start sophrosyne: %s", err)
	}
	te.api = sophC

	te.rootToken = extractToken(t, ctx, te.api)
	require.NotEmpty(t, te.rootToken, "unable to extract root token")

	te.httpClient = newHTTPClient(t)

	te.endpoint, err = te.api.Endpoint(ctx, "")
	require.NoError(t, err)

	te.healthEndpoint, err = url.Parse(fmt.Sprintf("https://%s/healthz", te.endpoint))
	require.NoError(t, err)

	te.rpcEndpoint, err = url.Parse(fmt.Sprintf("https://%s/v1/rpc", te.endpoint))
	require.NoError(t, err)

	return te
}

func extractToken(t *testing.T, ctx context.Context, c testcontainers.Container) string {
	t.Helper()

	rc, err := c.Logs(ctx)
	require.NoError(t, err)

	var count int
	buf := bufio.NewReader(rc)
	for {
		count = count + 1
		require.Less(t, count, 100, "unable to extract token within first 100 log lines")
		line, err := buf.ReadString('\n')
		require.NoError(t, err)
		var d map[string]interface{}
		err = json.Unmarshal([]byte(line), &d)
		require.NoError(t, err)
		if d["token"] != nil {
			return d["token"].(string)
		}
	}
}

type ensureJSON struct {
	t *testing.T
}

func (e ensureJSON) Accept(l testcontainers.Log) {
	e.t.Helper()
	var cnt map[string]interface{}
	err := json.Unmarshal(l.Content, &cnt)
	require.NoError(e.t, err, "could not unmarshal log: '%s'", string(l.Content))
}

func newHTTPClient(t *testing.T) *http.Client {
	t.Helper()
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true, //nolint:gosec
			},
		},
	}
}

func outputAPILogs(t *testing.T, ctx context.Context, te *testEnv) {
	t.Helper()
	logReader, err := te.api.Logs(ctx)
	require.NoError(t, err)
	l, err := io.ReadAll(logReader)
	require.NoError(t, err)
	t.Log(string(l))
}

func TestStartup(t *testing.T) {

	ctx := context.Background()

	te := setupEnv(ctx, t)
	t.Cleanup(func() {
		outputAPILogs(t, ctx, &te)
		te.Close(ctx)
	})

	t.Run("API served via TLS", func(t *testing.T) {
		conf := &tls.Config{
			InsecureSkipVerify: true, //nolint:gosec
		}
		tlsConn, err := tls.Dial("tcp", te.endpoint, conf)
		require.NoError(t, err)
		_, err = fmt.Fprintf(tlsConn, "GET / HTTP/1.0\r\n\r\n")
		require.NoError(t, err)
		status, err := bufio.NewReader(tlsConn).ReadString('\n')
		require.NoError(t, err)
		require.Equal(t, "HTTP/1.0 404 Not Found\r\n", status)
		require.NoError(t, tlsConn.Close())
	})

	// The Go default HTTP server responds with `HTTP/1.0 400 Bad Request
	//
	//Client sent an HTTP request to an HTTPS server.` when receiving an HTTP request on an HTTPS listener.
	t.Run("API not served via plaintext", func(t *testing.T) {
		rawConn, err := net.Dial("tcp", te.endpoint)
		require.NoError(t, err)
		_, err = fmt.Fprintf(rawConn, "GET / HTTP/1.0\r\n\r\n")
		require.NoError(t, err)
		status, err := bufio.NewReader(rawConn).ReadString('\n')
		require.NoError(t, err)
		require.Equal(t, "HTTP/1.0 400 Bad Request\r\n", status)
		require.NoError(t, rawConn.Close())
	})

	// When a client terminates the TLS handshake due to a bad certificate, in this case because it doesn't trust the
	// certificate, the server logs a remote error. This tests ensures that when that happens, it is logged. Because
	// of the LogConsumer added to setupEnv, if this log cannot be unmarshalled as JSON, it fails. Thus this test
	// ensures that it is logged as JSON.
	t.Run("client remote error logged as non-json", func(t *testing.T) {
		tlsConn, err := tls.Dial("tcp", te.endpoint, &tls.Config{MinVersion: tls.VersionTLS13})
		require.Error(t, err)
		require.Nil(t, tlsConn)
	})

	t.Run("TLS1.3 or better required", func(t *testing.T) {
		tlsConn, err := tls.Dial("tcp", te.endpoint, &tls.Config{MinVersion: tls.VersionTLS12, MaxVersion: tls.VersionTLS12}) //nolint:gosec
		require.Error(t, err)
		require.Nil(t, tlsConn)
	})

	t.Run("Health endpoint is available", func(t *testing.T) {
		res, err := te.httpClient.Get(te.healthEndpoint.String())
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("RPC endpoint is available", func(t *testing.T) {
		res, err := te.httpClient.Get(te.rpcEndpoint.String())
		require.NoError(t, err)
		require.Equal(t, http.StatusUnauthorized, res.StatusCode)
	})

	t.Run("send invalid json body in request", func(t *testing.T) {
		res, err := doAuthenticatedRequest(t, &te, "POST", []byte(`{ this is not json }`))
		require.NoError(t, err)
		expected := []byte(`{"jsonrpc":"2.0","error":{"code":-32700,"message":"Parse error"},"id":null}`)
		compareResponse(t, expected, res)
	})

	t.Run("Create dummycheck", func(t *testing.T) {
		dummyIP, err := te.dummycheck.ContainerIP(ctx)
		require.NoError(t, err)
		rawPayload := []byte(
			fmt.Sprintf(
				`{"jsonrpc":"2.0","id":"dummycheck","method":"Checks::CreateCheck","params":{"name":"dummycheck","profiles":["default"],"upstream_services":["http://%s:11432"]}}`,
				dummyIP,
			),
		)
		res, err := doAuthenticatedRequest(t, &te, "POST", rawPayload)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, res.StatusCode)
	})

	t.Run("Perform scan using default profile", func(t *testing.T) {
		res, err := doAuthenticatedRequest(t, &te, "POST", []byte(`{"jsonrpc":"2.0","id":"1234","method":"Scans::PerformScan","params":{}}`))
		require.NoError(t, err)
		expected := []byte(`{"jsonrpc":"2.0","result":{"result":true,"checks":{"dummycheck":{"status":true,"detail":"this was true"}}},"id":"1234"}`)
		compareResponse(t, expected, res)
	})
}
