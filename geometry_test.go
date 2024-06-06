package pgxgeos_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

func TestCodecDecodeGeometryValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				original := mustNewGeomFromWKT(t, "POINT(1 2)").SetSRID(4326)
				rows, err := conn.Query(ctx, "select $1::geometry", pgx.QueryResultFormats{format}, original)
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
	})
}

func TestCodecDecodeGeometryNullValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()

		type s struct {
			Geom *geos.Geom `db:"geom"`
		}

		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				tb.Helper()

				rows, err := conn.Query(ctx, "select NULL::geometry AS geom", pgx.QueryResultFormats{format})
				assert.NoError(tb, err)

				value, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[s])
				assert.NoError(t, err)
				assert.Zero(t, value)
			})
		}
	})
}

func TestCodecDecodeGeometryNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		rows, err := conn.Query(ctx, "select $1::geometry", nil)
		assert.NoError(tb, err)

		for rows.Next() {
			values, err := rows.Values()
			assert.NoError(tb, err)
			assert.Equal(tb, []any{nil}, values)
		}

		assert.NoError(tb, rows.Err())
	})
}

func TestCodecGeometryScanValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, format := range []int16{
			pgx.BinaryFormatCode,
			pgx.TextFormatCode,
		} {
			tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) {
				var geom *geos.Geom
				err := conn.QueryRow(ctx, "select ST_SetSRID('POINT(1 2)'::geometry, 4326)", pgx.QueryResultFormats{format}).Scan(&geom)
				assert.NoError(t, err)
				assert.Equal(t, mustNewGeomFromWKT(t, "POINT(1 2)").SetSRID(4326).ToEWKBWithSRID(), geom.ToEWKBWithSRID())
			})
		}
	})
}

func mustNewGeomFromWKT(tb testing.TB, wkt string) *geos.Geom {
	tb.Helper()
	geom, err := geos.NewGeomFromWKT(wkt)
	assert.NoError(tb, err)
	return geom
}
