//go:build integration

package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/tls"
	"fmt"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	"net/http"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
)

func TestStartup(t *testing.T) {
	ctx := context.Background()

	nw, err := network.New(ctx,
		network.WithCheckDuplicate(),
		network.WithAttachable(),
		network.WithDriver("bridge"),
	)
	require.NoError(t, err)
	defer func() {
		err := nw.Remove(ctx)
		require.NoError(t, err)
	}()

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

	// Clean up the container
	defer func() {
		if err := postgresContainer.Terminate(ctx); err != nil {
			t.Fatalf("failed to terminate container: %s", err)
		}
	}()

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
  name: users`, pgIP, "5432")))

	req := testcontainers.ContainerRequest{
		Image:        "sophrosyne:0.0.0",
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
	}
	sophC, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Could not start sophrosyne: %s", err)
	}
	defer func() {
		if err := sophC.Terminate(ctx); err != nil {
			t.Fatalf("Could not stop sophrosyne: %s", err)
		}
	}()

	apiEndpoint, err := sophC.Endpoint(ctx, "")
	require.NoError(t, err)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	res, err := httpClient.Get(fmt.Sprintf("https://%s/healthz", apiEndpoint))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)

	res, err = httpClient.Get(fmt.Sprintf("https://%s/v1/rpc", apiEndpoint))
	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, res.StatusCode)
}
