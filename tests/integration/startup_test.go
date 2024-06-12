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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	googlegrpc "google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	v0 "github.com/madsrc/sophrosyne/internal/grpc/sophrosyne/v0"

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

	healthClient         healthpb.HealthClient
	scanServiceClient    v0.ScanServiceClient
	userServiceClient    v0.UserServiceClient
	checkServiceClient   v0.CheckServiceClient
	profileServiceClient v0.ProfileServiceClient
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

func (te testEnv) getAuthContext(ctx context.Context) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{"authorization": fmt.Sprintf("Bearer %s", te.rootToken)}))
}

func newClients(t *testing.T, te *testEnv) {
	conn, err := googlegrpc.NewClient(te.endpoint, googlegrpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{
		InsecureSkipVerify: true, //nolint:gosec // this is just for testing
	})))
	if err != nil {
		t.Fatalf("could not create client: %s", err)
	}

	te.healthClient = healthpb.NewHealthClient(conn)
	te.scanServiceClient = v0.NewScanServiceClient(conn)
	te.userServiceClient = v0.NewUserServiceClient(conn)
	te.checkServiceClient = v0.NewCheckServiceClient(conn)
	te.profileServiceClient = v0.NewProfileServiceClient(conn)

	t.Cleanup(func() {
		_ = conn.Close()
	})
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
		WaitingFor:   wait.ForLog("starting server"),
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

	newClients(t, &te)

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

	// When a client terminates the TLS handshake due to a bad certificate, in this case because it doesn't trust the
	// certificate, the server logs a remote error. This tests ensures that when that happens, it is logged. Because
	// of the LogConsumer added to setupEnv, if this log cannot be unmarshalled as JSON, it fails. Thus, this test
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

	t.Run("Health endpoint is available and API is serving", func(t *testing.T) {
		resp, err := te.healthClient.Check(ctx, &healthpb.HealthCheckRequest{Service: ""})
		require.NoError(t, err)
		require.Equal(t, healthpb.HealthCheckResponse_SERVING, resp.Status)
	})

	t.Run("Create dummycheck", func(t *testing.T) {
		dummyIP, err := te.dummycheck.ContainerIP(ctx)
		require.NoError(t, err)

		resp, err := te.checkServiceClient.CreateCheck(te.getAuthContext(ctx), &v0.CreateCheckRequest{
			Name:             "dummycheck",
			Profiles:         []string{"default"},
			UpstreamServices: []string{fmt.Sprintf("http://%s:11432", dummyIP)},
		})

		require.NoError(t, err)
		require.Equal(t, "dummycheck", resp.Name)
		require.Equal(t, "default", resp.Profiles[0])
		require.Equal(t, []string{fmt.Sprintf("http://%s:11432", dummyIP)}, resp.UpstreamServices)
	})

	t.Run("Perform scan using default profile", func(t *testing.T) {
		img, err := base64.StdEncoding.DecodeString("iVBORw0KGgoAAAANSUhEUgAABKgAAABQCAMAAAAKjb1wAAAAkFBMVEUzKTqus883LD6XmbObnbY/NUesscw5L0FFPVBBOEqgpL5zcYhYUmaEhJxqZ32Ulq+qr8p3dYyeoLo7MkR+fpVlYXZMRVlUTmJRS19vbYOBgZmorMeJiqJ8e5JDO01bVmqlqsSRkqttaoBoZXp6eI9eWW6ipsFJQlWOkKhgXHGMjaV2dIpPSFxiXnNHQFOHh5/Ud/GvAAAfWElEQVR42uza226CQBAG4BlaRA5qARVMKkQpwQOt7/92FavxYsQlazZR+n93m2V3/gywV0tEdh0E85haxXUSVA51FeffFaXMn9QiSxPn7v4WUdCsF893ZsWu62jkFfUFuyo95jUBwOs7//jaDswhmSbzqut/8NGgIgB4fdoH1W61dWnoM/MPmSbzqusPmDnaEgD0gPZBNeGzKRkn86rrJ0GdWQQAfaB9UK34j2+ReTLvtT5OI4Dei8fjhHRY85m/WOQZKRjKe60PAAAAAAAAAAAAAAAA8GzqdOiQAWlZWvSgoMwNr398P7csh7fGedn4XxfVFf3RkCXHJrrUQ6n+tyn7+zTMvK/rheqSuojDMCzuxiwsunhn7h5Yrm84zDw0ul5Nud+M2b815pMRmeOHM3L2Ya6aF/0RTPdHO++GGwWd5NOJt9oHdB6FFyndZhej6O1LlBPzLeNkuYu8zbKmVvP1IfImy4RkPnXeiNlu309N9leyf8m50u40gSg6l1XcIIBoRY02bnGp///fNQ+HYrnQsaXt6Tm9nzJ5vO2+mZsMNv3I5rbH11UlhW6+VKB4TXz0runG6YfzneJ5dZpH236IACBaqSewkyf9dnsKuF2ESvy7CI34/zNC5f1xoQJs1aMUbCd+CH+an1+sVx1RbXx/hDu2d9sNGm3nNHBQYKnMdl73+tAYte34Me7oK6rPUC8JFcX7LULl2wBmP4of5AAO8tUZD1AE5sOBhr3ieXWfB++HILvun/yj/4L6Ybvd6SZU2r+CBWDV3b87OJ5ZqPwk2f1jQqX5IfwRfroLlSywnR4OvixeAORODuDlKaFKYnHwJITZzusLBBEAzFUjlhCs5eBTfbQ2CRXH6y5U1glaqFrjL1AK1QAVYkVgPhxZ5rK2LZpXt3m07u+xyJgZVlFW+veESvWyQ3f/7uB4ZqESTP8zoSJ+OgrVEHCUxioCbr6yUiBK9MEfZwWStgOMq/IXgOca7by+IH/JfJXZALIm/y8AnJmvgk8p1Weol4SK4v0WoZpDMPtB/Fd8E6ok04gaf2thPuYjeRflpgDeaV6d58H7QSe4KTOmhR7vf0qogpVV2WVZcwkC8v8RrNV38ZJeon7Kf+Vy/na4YvR71g8fkI1CaxIq6p/B/QisXkD+nYWK43N+5ttcP/PTSaiuwPJhH6993cREH/xp5cv1hUAqLfWBN+qH7LRO3hJ9OWrxT4FQs8X1mevVQtXzFcXjfig/88v2AQrM2uMne+xLoSqRAQgoP/FRYQNsaV4mvs3z4v3w/N8Jb4FBDFx04+G97ywMRYE/27YNILY/MNUD6i1jwEvLUqdHD4jCV6339sZPY2D9Inb2Vwtb8HCy/M8xAM958wvN+RRGANaLg3rSfycO+0WP8xPEe2QDTu8WIZ80xrOWfWCfvgGD2pqEivu37W3Vl23f6v1IubZ9ORwBRGlC/j9z8Jkfik98Et+Un2Dgp/En5C3V2dj+EWsNRBJTMp7LK8cY2NDB5/p6kJ06yWTX2tQP2/l5jSOwbPBPoof8XJ+5XjlJpxxeOFaqFo/6ofzEL9tnANK+CFUVny+v7xFwqH3vRPmZjwovwILmZeDbNK+OQtUHLkvg/P3VbnifyBwVJtruQFDyOITGsvTfQFC4s7/aQ9CrXbkF73pnaIyf819qa/7K+RlrPCBoiOeGKDGorUmouH9g/0h/WO9HpAUY71EgI/+fEirih+ITn8Q35ScY+Gl5BxW3CRUE1cvZURnmFdiXBz8IVAu/slpbM8TyRU79sJ2f17AlMftfitItXQDVZ643BiIUOKtaPOqH8hO/bPcdnCwtVBxff2ujakLl58CO8hMfJFw0LwPfhnm1C9VJGXGQUPoaykL1ulwuPWC0/ECm7UCY9suPFL/kQL5I5duvpd27LfuAtyJ/wac0jR6FJoUE2J4iLVQS/XyMgDx4xn8HIF7OpYqE8jPkkmsXLeTAtSHetmiveGRQW5NQcf97YU9jDLxQPyIkgvgURsjIX+AhVKtGnSU78UPxa/mJb8pPMPDTUO+4FMkm+0edoUzsA5e70J7LOXrVy+k8HFtN/KqrnMIBsAoA+NQP2el5jcQDpsyH6E96PUb6NwKqz1xvDCA6zftSby0e9UP5iV+2q9XcV5VQUXxlOfAOdaG6yiHn/MxHudyKneZl4ts0r5b9nYnaGvEJuCnfA3okVC3vqIQkd38/51YMOMm9fkfbo2lxT8aE/DVCSVZRiLMlHW6GBfPbXvkR7Nnkr40bX6lLrgdL+UmoAvHfSoMvHM+N7nHmAAa1NQkV9z96GP9NvuZ+REjyney4rUv+gsT1P/K6liKQnfih+LX8xDflZ7Tz01LPEECUtNd7Fear3bco71hwC9Y07ID51c+PgKEPYEX9kJ2er/Q2TxT7vwF73LHk+sz1qvh+Qi0bOHE85vsxP/FLdg0tVBRfV7xVdaEKxcz5iQ9BsgljD8CI52Xi2zSvlv29AjBUJtiF4myA8bNClep3ep+kXCC676g9sLrbl9o+eEZoRsDGUoyddGj2zx4+h40V52eh2he3xZ0KgJTjTYC+lOOL8NXWJFTc/1shB9fRaFXsJpf7KYTkXbX4/xJYqN6b+SS+TflN/DTDWkT5J0VoFKp3wJPYyRrAl+Lg5+HJ8QCcmurbyrY8SocecKF+yE7PV1r6mfefvimvR/M9gCnVZ65XhGqkN6bnUjzmu8rP/FJ9JFQUX/UirJOaUElWZA3zJj60bAjShnmZ+DbPS4P7iZavF1/9AAGAoNjem2eF6lXvMGllVw5ILYDp3T6t7GahijXnFdzp5LzdnoCN2V/y2+WLEU9xfhaquJjuUCXAjeMNxE23M6itSai4/6yo2gbehNiY+xGm5fvN/h2FiuJTfuLbkN/ATxssSz0pVP4e6A9XUwf64A939x/XkH3G9aWS/whM5CjOqJ+andfV/esoeZr84cjn8zZgU33meiXeu95puFA85rvKz/xSfSRUFF9tgKuqC9UL4DTOm/mQy9Iagi3Py8S3cV4tmHoQTAz/XsPRPwCSJ4XqUNmFAngFdCIH+FLZzUIFIHm0JS8RNMIn/MelaFgeEFB+wlrG+Uk4tIAFxxvpd4byzKC2JqHi/q0cueVCpvUKpE39fK68yb+jUFF8yk98G/Ib+Plp8MafoEJVWHF13zbUN5DubsDOAnCgfshOzwuCPhALZc18DPX+gUX1GetVcZnmCOwoHvNdhWF+qT4SKoo/kTHXhcpaA+OmeTMf2iG7iZnmZeLbOK8WXCKzUJ0ALD8A4N0oVGxf4AFvZH9GqHL+L9SjOF6bhUoTvy0lCBnlbxWqWYtQHYUGrYADWpNQUf9H4PAuHVhn4NrUz+fq5Rv5dxQqik/5iW9DfhM/3YVKTXIIUiCqfTx+bKjvDTiKbZYAcKmfmp3XAtcB9qvm/Tcuy7AA9Kg+Y70iVKuS2QnFY76r/Mwv1UdCRfHXwCT7AIBrFlS/Nntu07yZj0fF7NO8THyb58XQpEXpbuardvhRFff2KES754RKrqzjEr1fEir4tQ+j52IcPidUn76l8YBed6EaAZMy8IDWJFTU/ydgMkcMTI9FUO7nc/Fyr9m/o1BRfMpPfBvyG/npLlQq2H1l70ob1ISBKAPIoa4oIKz3hcdaq///35WErEN9S1Ptdttt8z60pJO5k2dFJZ1i25rfvGMdijHGtyVyxR+tPZEH+YAcxiX8iGiSNay/l+v+XBLlEJ8uXrnnzvw/KrAH9Wb/WF+ID4gK7FMdU2a/y5v9hnow9uyY+6Wpt7ZfDZiTqNWPsSDGxBb31lWhe98RTdhEVCN+/QY56vPCvxZmx9eqP5Gl/DsafdWo1fVem30vUaG94rUpY6IOjiVy/toH58+ydEc9h4oJ7SAfJBKpr4XdJMD6oH32D/XW+dfVpwHZrBPqiOoWHdFIWAgQ3zMRtZLVWH7cBfmAHOdb9oFocrIkUD8vhdcXvj3Ep4tXEtWCyQTscT7gH+sLciCqG/tAVPChGvjneuC96wD6pam3rl8NyDQ/hFT1uCQCoVdlc1BJTXmjO9wyIIJt6cPXEJXSZxz5XQb74a/Jv26OWKOvCuFJfz1RiHuJCu1tiNzXhnRwLJFdFwfm73tU4tyhEhfIh4kE9ZvhH7zJulnM9QH74B/qrfGvq48CrnH3TqKyq6+78NghSjE+ebvlRcWUQj4gx/n2lPcl6st3GJlqMvkQny5eSVSz12+9hmCP80H/WF+QA1GB/c26QjlK190rp+7spn5zPfCRKi3ol6beun4h9F/4ZDvD6nJVVTgluigmHHNsvSYiSkoDqS1NLdYgB33+Pj98mu5/ySq+OSjmp91P6PvLyk/YVv7uIiq0J/j6qwqrg2OJhH9qhflHJBjhRCV6kA8QCeg3/a6zb1sIqA/YR//1emv96+qDUJuNqVNHVP5JliFS30g8jUIZzICIhhifjGXXqt4JPEM+IMfxE1Gf9yXqj8tc7Wr/xhCfLl51u2VfNSUCe5AP+If61uVAVGBfQd1MZ04dNKw3qMdpdrYEzqWSC/3S1FvXr0fPBFV8FF4j2FVR0eXrl36NqL4QeePecDgHIlDznVJYtKmNctRXMjqM1mtf5BOV0tVoM2rL/PNStBqNjh6VaOn0VeGmvV5czgnuJyq0dyTqF73UUwsFxoqMlsV6nUP+6mOVmXwBoAzyASIB/YaN39BlqA/YB//1emv96+rTfDvBa2mIim27xehp92puRN6hWBdxOXbtt+LLhKQoJnJf3uQDchznVGJVoQP6akdEvd5BEA/Gp49XXLU7o6nQB3s3+YB/qC/KrbSMvBTG5V8h2AeiUt07N6w3qMewlBa90UX470G/NPXW9etuosIvWCiLmeRfiQlv9NaEf0uGRJCSAhAV6N8+iesspTEpvNQ/lnBLpcVP6PsR/zTqfqJCe/M2sUEYq8J6cnyB/NUu7VYvVDbkA0QC+m8h7Dd8Igf1BfvgH+qt8a+pD0CtZa+wfpqoFAq72vivWGZv12ejRrsympt8QI7jF2IcUL8ewQri+5l4Y75DZIM9yAf8Q30hPnZAe7CPRMX/t7PQP9ZjSBw/9EtTb22/mtDVE9WxFsOucn0WgXjTc+1Xi2FnR30VWKReLPOrvFtF7h1BjvoKdrHzrjfr5pK+qf80F6MklYNpUhBttPryui/XSbkLwT8ippWsXWbZfXp6y14gmS9aiP0OY4VT1PdIqGP+reppqVu1DzCfDecF+g1ojRXJILA+bB/9Q701/nX1aULY0jxRllWzyns8VMGoferNkqb6DHcyABHMTT4gx/EC7jWD/qYt3XVsiO9n4nVomS+lOdsCe5AP+If6gtzlBAKwf8WylPLlsGG9YT2y6YQk4i30S1dvfb+a3zKMrfth7/Pch3+0fb9RIznl3f2NWKfvh7yaW6duN0uuozzP7Dv0RcDd0HoUaC/Jc3XJY4gnCUNfkz/k83D9Urzno+8P+od6g/876vNuCBZ5nvAwfF7k3cD+UX3mz7mP+YAcxoAG/TDPzzbEp4+Xp+TPtgX2IB/wj/VFOQPt66Hvtx10t/kptBD6euv7hWjtt22itWXw+dGdkGN9DM77Gv7JY2IM/i70qYT3fx3r9I/ii/zB6ceA6jhYBgbvDzwu68Uy+PwganetD8KyjifLwOA3I++e5rZl8A9gVfiWgYGBgYGBgYGBgYGBgYGBgYGBgYGBgYGBgYGBgYGBgXju4TawDAw0eN5u7f95PZX5m28N/UnwWfmdwdDyvwwWOAfk74T5pqH3D8bTGgwGLbWuBowqv3x2jNuH2clCoBz1Ea3NU+wUGB7E04RsfYnb0SxvmH9KV2132kkwHtB/9/5hf1yi8Heupy9Vai8n9t67xE769RrCeXCM49XTZl5ebwffIfjo/B/wxz1ea9cLztfj8f3zOcAL644jyd8HadNj4x+Mx3eofm4aoxwHbVJ48sESyFEfEboNJ6pDPA3oUYX22/NTPvIe4gH939K/FJ65/xvXE73C7SoaXpLEQfVr4JHCXD0Eh7H96Pwf99chijTrBefr8fj++Sw/Y/iDROW+L1HZR+JGd4gRy4P9BfpUYgyWQA76iCQWJCI2T4FCjAcxI4GlIhqYPxLhRK5wEkA8oP9b+uf+EaJSv4/cT0QBRH2jaz1ot6MSwQcRlfsbiMpfEi006wXmG6L6p4hqTLVGJ5lCn2gtiWgyyHwrc4jwofsgB/0360ZfLX+KD8WEeBDqEFz32bfCUfrm/B1RFFZPciwgnu/1/xWiyoJ994nUkcMpkbewk5kirhYRHcRBot2jIKowE3CJZvLC/jRE1YPH0sN6wfmfjqjsMPCtOsI5TAlgOTXPb4UPLyz0h/ZbQuQH9cPqW7AQIL5HC92htxqdUbXAkk1SeZgQPnQT5KCPiNQ5pW2wB/E0va+Ikub5AamTtFdEK4iH9R+sF/YH648bVUzg2dDPR9cTH+ekypBVR6P3xPpayRn1wwfO9QeYbn5pY9rf76ckSPT5/xIR2G3x8qZfLzif9xvGz+P5A/G9O9+8yJPyJ5dcjfNDOe5HCzXcOs5pfyCifpqUPXQcZWHoOOu35tuzNtEu3dy/sNAf2p86Tw6RG1z6NBkqhahfOpwGcrB2nFJOsVMih3wa4ule0j3Gw0+RT9vcaDjfgBf3zAKAHPUZTCTDzCqIHJBCPBB/0ifKm+fz8WxrIhfiAf17+of9wfpjf1yiYBYTealtQT+164nz1xGV7clHnr6QYolh9eTVDrxl/0WiujjH82FCnvOiOHck0qHldK/P/5H8+Dg/oY/7k/sP83G/QfwC/lo+49fd+I/UI42q170sitZVP+/imzpIQVnuksLseshDb0cSmXzB71kSjugmzm9Fang/UaE/tL+kGkK1xWo3h8fEGGI+6E+RQ9wUj+/S0YZG+5PbIxIdkS8A5KjPUNEu7WeKxcUEpBAPxH+Sl3bYNN8nor26iz6FeED/rv5hf6D+2B+XyOXlAv3E9aTvHxKVimwtb0k5V6VAjqN3Jao2karB6ObTl542/3vzY7jKPO5P7j/Mx/2G8ctbXAovD9QDzywY38E3t4bjdJS2leJ5QjSZpqJ4C6UoEB+jvlDcvDY2kDdRcH5B5Yxv7J3pttIwFIWbaAvIVJkqcJm0DPfChfd/O5M05aBftQbUpUv3DxVzTnLG3TYtdPTmJvKJ0VhaMwGMYz3Mb4K4c5P358XJ69YKL2ZWyhLwarFYJCYIC4MJ/Km2p3cTFI4vZ40IiX4t3o4kGCZyJgJgHPq3Axt7mF+2+Epc2kP7V0qNXg9peYSi/MEscI50x4RlDHugH5Y/5AfxZ37eOmF7zE+XEfKJegrMnxBVXDDD4togiY2JW24WVxJVuP+eqFRyyd+WJh2t+cdDavxp1fsf7l+Z9bXzkv0p+ac8+432RyNlE5i/pOo5PB4kqpMK4ptbFCeBeuX6SA8Mww8L+95eFedb63FuHI/9m9LtoeilQj5Oi4KaSWENYxOTONYRgHGsh/nX1u++Urn1/H0RiI3RP83NetgDgD7X85yeDmmPAETV//pCL/eNRXBc9ImpPdHpKtVpuPvlhNgD+32h7m6OUJQ/2+Hd3O4ha9gD/bD8MT+MP/donG2xe4En8ol6CsufEJXuFqeSF+eXt3Tr3jtnsGl/qCCqQP+FqFbFy5AujhPyZvmIwLHO/7vrs++Dg/68zT/l2W+0/1UZw7UtnE0nPB4kqqkK4ptbFBPITaO0Vb6mcClvGLx93dLUG/BcIT/2djSEOAKA9TD/Wu3cYXEbtdyu88TWn98zHKAQoF8J/ZTOjU8EEy2zCt37NOAmHsapT+Q2nwcbgUSpU509tP+oDNbdmWWbDPJ+bYeLhj3QDwXzw/izUUc+f9MI+UQ9BefPVtNqO7V8MPiiRQcFG61S5fBuqklUBMHUPJUvDNSRYGv/v8b/YP9kHzKNK/oT+Yc8+g32dw2N6+hekKiO6gf5hlDptHkbz5erx5lXHNwGvGC+vduLpHzbxt1/DCcqrIf512rgnO1EQ0f525s9hwSFQP1q6OpUMNF8/eH1/PWAOTAOfWJk43ewVymprFlrj9aib6Y2mY/fuMBQvjlQHl0Ne6AfCOaH8Wejrvz27ihCPlFPwflTV6STYudmeiWqnlvo4qlq0/hJRLX98gVVcTY+5vmLvaav8//O+jRzL6r6E/mHPPoN9g9E9acQ1VgF8A1T+e4y1mXJqsTBXtN7xe5X28h7V4vdKvlumePpvUQl63F+eWFo5l8Y2iultRFooRCgfweYaL22RS5omfFBDC2MQ59oW38uplg0XoMOewDvb6c8h9OU1wOT7G2reTTxOMIe6AeC+WH82ah7KWTk8+F6UiW6Tc9B74sBudZqrHLH3u9/ElFN/Ga2a7zh+1R59KN6/8PhSClZVvQn6gXy6Dfab+we/kyias1/mG9YWw5vJp7XBJ+84vGr58fazp9OlfyhPG/r3UtUsh7nx5uNraP51Y0JCgH6P4Wotl8+jRmbBXfLiOA49fnuxYN16sNQiRcBRGXCnkYWluiakHdLn33jpw3YA/0wMD+Mv+SHhYx8Pl5PdtnX19VkeK3e0fW1TB0ROw2M4/rnEJVLdWn5Rhmkg8GaREX/70PXtTb7E/UCefQb7cfN5zDQv04g3wj01u25q4F3YtMr0ax65XjHSp7sXSLKu8/jx86oZD3Oz0aYXvNrdxFRCNB/jKikkC/yye47zifQwTj0AX/mbv+I7WHvx+zBAzLXRs8o3y5PtS0R7mEP9B8jKsS/plGRz0frSe76yRnn0/U5jclXu4jLn0NU+3LDulNMO2va+X8ZUS0TvyT7k/kXefYb7Uf87ieqrfUvnG94z9sfQKfgM1GU64RT7q9yIW8G3guxfwv6h4iK87MRnstL/5a/VJHYiH4dJot2K4Colur2cKUPhodO1OE49auf54yHm5m7/VZrD+3PrkdAW46Un90SQQZ7oB+WP+aH8Zf8sFGRT9RTWP7YaD3jn/aOihX+UvMEogr1X05l7D/O8pyWXfhNnf+B/vHZYfan5J/y7DfaH+2Ez++Kxxszjfd/Fs43hDYGPctXCioVhYvync8q5D+V27K7bxdW45DMP9YTFeZnI7jGTlo+EDuJzZ761Sh74m0AUbXNUlpuyZCnmtNnjkOf8r7Mnv2G5wjjsAf2N1JfWBPboZQ/usbx7LSHPdAPyx/zw/hLftioyCfqKSx/JKq9cpUr9w+ujgtxgahC6ved9UO+ApCXZveVGtT5f1d9xqkUA/tT8k959hvtj55gV1g8DrJHNAvnG9JhY+AmdM+VjLT779XHasWTfNWe8q1EqVendVtY/N5tquuJivOzERrrIgAtH2Afk161PuDJ1TVtPVFJjN/fXpumJ34N5gXj0Ke8P97tYvdInvqAcdhD+2dGVlt+9CniHpVaebsSDXugH5Y/5ofxl/ywUZFP1FNo/oSohCX6DcPDaWFENm8PI/+c1Rqb6SH+S6jTib+mKfj2UP5Iw67O/0D/+Hst7E/JP+XZb7RfnhZoHCf3xGNUbCt8UN6/0yqQbwS798+r49tyb/XV7nP1xuP8nd+soOJASdVA/sVMlPdGSV1hfaglKs6PRvCBfOr1jE1JU54lSmZGZQn9aqyUbCYTo81mYxNv/mpddzPO5WimDDYFSn+nyiCuGoc+5N2ZzNs8n5sVMQ57aL+riH6vd3A3USgf263Mxaex/a8u7IF+aP6YH8Zf8oNGRT5RT4H5I1F9svYsFmun4padX9rjqa3pEYiKoP8kKrX+2Lv4y+nMpn86fUlcDmv8D/KPv9fC/mT+RZ79RvsNdN8svpl+mr5Tz2HxkP69vB7TkqgWKpBvBLLn7inQA4rCyVLakF++Ux7fPlVP3WoE1sP8bIRGH8vFc+Uwhj7hc5HkkYBp99iXB8T+zea14CD7IC4LHIc+5F3aHXZnjtMe2j9VHptK+ZWEYwh7oB+aP+aH8Zf8oFGRT9RTYP5IVHqmCqQrv2yJQYNEFV6/Yu7FR7fAW+P0qsb/AP/4ey3sT+Yf8ug32B/F1wme74mH3imH+ZWoAvlGkBRp2kYencKy5KUw7BOMaCrZGKR801VafyXrAfHMJo3Aepx/oDZOZBLptDgl0Hlqh9c3E7baO5XawFK/Gq34u7ctSjR9L6rxzdFO8FS6ZyTyynHoQ95gvHMBbGKc9lTa/8nVWtLW1fKnTRGOfBjBHuiH5Q/5QfyZn74/Wcjs9i/yiXoKzl+Cq5L22tX7h8KVj33fSFPhs6ficjPMf2n017l1119ND10jpk/D3NR1nf/h/umd2Ir+RP5Fnv1G+z2Wl8S50F3e1c/ngZ3t6ez9W20C+UbQmHSyrKkjwfCUdfaStxpAfphlCCwuXPfRnfMTep91WniWt9GA/u+Cnpwfkl9+yBqPzNfKsrPkE4hNOM6NYP3H88f81OeT9fQ4mh9uJxxK/RP1/pOohtF+e9IS7iybyCf4/xhe5fdaHpWn/VIwnc5kGBwPSWfW+BG++fPQmRvG/4+/Fv96/j63c8cqCMQwAEDbSRDXzqcgOIj+/+cJHcx0cHBHSbn35gwJSQMd2qh/5aAPE/+17I+P/BPNw/f1V4Z79wevzOrs/ev1J1lUS2uXg+Ij/0TzUEMZrtbrrTCts/ev159kUR0o8k80D/dQhns8M95H0b/t9VtU5gHmtbRPmdns+QMAAAAAAAAAAAAAwJofhnNJW1blsQYAAAAASUVORK5CYII=")
		require.NoError(t, err)

		resp, err := te.scanServiceClient.Scan(te.getAuthContext(ctx), &v0.ScanRequest{Kind: &v0.ScanRequest_Image{
			Image: img,
		}})
		require.NoError(t, err)
		require.True(t, resp.Result)
		require.Len(t, resp.Checks, 1)
		require.Equal(t, "dummycheck", resp.Checks[0].Name)
	})

	t.Run("Text param value is passed along to the upstream check", func(t *testing.T) {
		resp, err := te.scanServiceClient.Scan(te.getAuthContext(ctx), &v0.ScanRequest{Kind: &v0.ScanRequest_Text{
			Text: "false",
		}})
		require.NoError(t, err)
		require.False(t, resp.Result)
		require.Len(t, resp.Checks, 1)
		require.Equal(t, "dummycheck", resp.Checks[0].Name)
	})
}
