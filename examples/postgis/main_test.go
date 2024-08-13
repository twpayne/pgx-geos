package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestIntegration(t *testing.T) {
	ctx := context.Background()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found in $PATH")
	}

	var (
		database = "pgxgeosdatabase"
		user     = "pgxgeosuser"
		password = "pgxgeospassword"
	)

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("docker.io/postgis/postgis:16-3.4"),
		postgres.WithDatabase(database),
		postgres.WithUsername(user),
		postgres.WithPassword(password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
		),
	)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		assert.NoError(t, pgContainer.Terminate(ctx))
	})

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err)

	conn, err := pgx.Connect(ctx, connStr)
	assert.NoError(t, err)
	assert.NoError(t, registerGeos(ctx, conn))

	assert.NoError(t, createDB(ctx, conn))

	assert.NoError(t, populateDB(ctx, conn))

	r := bytes.NewBufferString(`{"geometry":{"type":"Point","coordinates":[2.3508,48.8567]},"properties":{"name":"Paris"}}`)
	assert.NoError(t, readGeoJSON(ctx, conn, r))

	w := &strings.Builder{}
	assert.NoError(t, writeGeoJSON(ctx, conn, w))
	assert.Equal(t, strings.Join([]string{
		`{"id":1,"geometry":{"type":"Point","coordinates":[0.1275,51.50722]},"properties":{"name":"London"}}`,
		`{"id":2,"geometry":{"type":"Point","coordinates":[13.405,52.52]},"properties":{"name":"Berlin"}}`,
		`{"id":3,"geometry":{"type":"Point","coordinates":[2.3508,48.8567]},"properties":{"name":"Paris"}}`,
	}, "\n")+"\n", w.String())
}
