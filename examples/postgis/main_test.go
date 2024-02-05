package main

import (
	"bytes"
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/ory/dockertest/v3"
	"github.com/twpayne/go-geos"

	pgxgeos "github.com/twpayne/pgx-geos"
)

func TestMain(t *testing.T) {
	ctx := context.Background()

	if _, err := exec.LookPath("docker"); err != nil {
		t.Skip("docker not found in $PATH")
	}

	var (
		database = "pgxgeosdatabase"
		user     = "pgxgeosuser"
		password = "pgxgeospassword"
	)

	pool, err := dockertest.NewPool("")
	assert.NoError(t, err)

	resource, err := pool.Run("postgis/postgis", "16-3.4-alpine", []string{
		"POSTGRES_DB=" + database,
		"POSTGRES_PASSWORD=" + password,
		"POSTGRES_USER=" + user,
	})
	assert.NoError(t, err)
	defer func() {
		assert.NoError(t, pool.Purge(resource))
	}()

	var conn *pgx.Conn
	assert.NoError(t, pool.Retry(func() error {
		config, err := pgx.ParseConfig(strings.Join([]string{
			"database=" + database,
			"host=localhost",
			"password=" + password,
			"port=" + resource.GetPort("5432/tcp"),
			"user=" + user,
		}, " "))
		assert.NoError(t, err)
		conn, err = pgx.ConnectConfig(ctx, config)
		if err != nil {
			return err
		}
		if err := conn.Ping(ctx); err != nil {
			return err
		}
		if err := pgxgeos.Register(ctx, conn, geos.NewContext()); err != nil {
			return err
		}
		return nil
	}))

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
