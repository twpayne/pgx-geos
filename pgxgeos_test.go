package pgxgeos_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxtest"
	"github.com/twpayne/go-geos"

	pgxgeos "github.com/twpayne/pgx-geos"
)

var defaultConnTestRunner pgxtest.ConnTestRunner

func init() {
	defaultConnTestRunner = pgxtest.DefaultConnTestRunner()
	defaultConnTestRunner.AfterConnect = func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		_, err := conn.Exec(ctx, "create extension if not exists postgis")
		assert.NoError(t, err)
		pgxgeos.Register(ctx, conn, geos.NewContext())
	}
}

func TestCodecDecodeValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		original, err := geos.NewGeomFromWKT("POINT(1 2)")
		assert.NoError(t, err)

		rows, err := conn.Query(ctx, "select $1::geometry", original)
		assert.NoError(t, err)

		for rows.Next() {
			values, err := rows.Values()
			assert.NoError(t, err)

			assert.Equal(t, 1, len(values))
			v0, ok := values[0].(*geos.Geom)
			assert.True(t, ok)
			assert.True(t, original.Equals(v0))
		}

		assert.NoError(t, rows.Err())
	})
}

func TestCodecDecodeNullValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		rows, err := conn.Query(ctx, "select $1::geometry", nil)
		assert.NoError(t, err)

		for rows.Next() {
			values, err := rows.Values()
			assert.NoError(t, err)
			assert.Equal(t, []any{nil}, values)
		}

		assert.NoError(t, rows.Err())
	})
}

func TestCodecScanValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, t testing.TB, conn *pgx.Conn) {
		var geom *geos.Geom
		err := conn.QueryRow(ctx, "select 'POINT(1 2)'::geometry").Scan(&geom)
		assert.NoError(t, err)
		assert.Equal(t, mustNewGeomFromWKT(t, "POINT(1 2)").ToEWKBWithSRID(), geom.ToEWKBWithSRID())
	})
}

func mustNewGeomFromWKT(t testing.TB, wkt string) *geos.Geom {
	t.Helper()
	geom, err := geos.NewGeomFromWKT(wkt)
	assert.NoError(t, err)
	return geom
}
