package issue6_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/twpayne/go-geos"

	pgxgeos "github.com/twpayne/pgx-geos"
)

var (
	pool       *pgxpool.Pool
	resource   *dockertest.Resource
	url        string
	dockerPool *dockertest.Pool
)

func TestMain(m *testing.M) {
	log.Println("Starting PostgreSQL container")
	{
		// Initialize the Docker pool
		var err error
		dockerPool, err = dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to Docker: %s", err)
		}

		// Using a plain PostgreSQL container
		r, err := dockerPool.RunWithOptions(&dockertest.RunOptions{
			Repository: "postgres",
			Tag:        "16",
			Cmd: []string{
				"/bin/sh", "-c", "apt-get update && apt-get install -y postgresql-16-postgis-3 libgeos-dev && docker-entrypoint.sh postgres",
			},
			Env: []string{
				"POSTGRES_USER=postgres",
				"POSTGRES_PASSWORD=secret",
				"POSTGRES_DB=testdb",
			},
		}, func(hostConfig *docker.HostConfig) {
			hostConfig.AutoRemove = true
		})
		if err != nil {
			log.Fatalf("Could not start Docker resource: %s", err)
		}
		resource = r
	}

	log.Println("Waiting for PostgreSQL to be ready")
	{
		// Wait for PostgreSQL to be ready
		err := dockerPool.Retry(func() error {
			pgURL := fmt.Sprintf("postgres://postgres:secret@localhost:%s/testdb?sslmode=disable", resource.GetPort("5432/tcp"))
			config, err := pgxpool.ParseConfig(pgURL)
			if err != nil {
				return err
			}

			// Register the GEOS extension
			config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
				// Create the PostGIS extension
				if _, err := conn.Exec(ctx, "CREATE EXTENSION IF NOT EXISTS postgis;"); err != nil {
					log.Printf("Failed to create PostGIS extension: %s", err.Error())
					return fmt.Errorf("failed to create PostGIS extension: %w", err)
				}

				// Register GEOS
				if err := pgxgeos.Register(ctx, conn, geos.NewContext()); err != nil {
					log.Printf("Failed to register GEOS: %s", err.Error())
					return fmt.Errorf("failed to register GEOS: %w", err)
				}
				return nil
			}

			pool, err = pgxpool.NewWithConfig(context.Background(), config)
			if err != nil {
				return err
			}
			url = pgURL

			return pool.Ping(context.Background())
		})
		if err != nil {
			log.Fatalf("Database ping failed: %s", err)
		}
	}

	// Run tests
	code := m.Run()

	// Cleanup resources
	cleanup()

	// Exit with the test result code
	os.Exit(code)
}

func cleanup() {
	if resource != nil {
		if err := dockerPool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
}

func TestFoo(t *testing.T) {
	ctx := context.Background()
	assert.NoError(t, pool.AcquireFunc(ctx, func(conn *pgxpool.Conn) error {
		box2D := geos.NewBox2D(1, 2, 3, 4)
		var actual geos.Box2D
		if err := conn.QueryRow(ctx, "select $1::box2d", box2D).Scan(&actual); err != nil {
			return err
		}
		assert.Equal(t, *box2D, actual)
		return nil
	}))
}
