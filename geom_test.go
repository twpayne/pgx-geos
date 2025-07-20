package pgxgeos_test

import (
	"context"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

var geomTypes = []string{"geography", "geometry"}

func TestGeometryCodecNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, geomType := range geomTypes {
			t.Run(geomType, func(t *testing.T) {
				for _, format := range []int16{
					pgx.BinaryFormatCode,
					pgx.TextFormatCode,
				} {
					tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) { //nolint:forcetypeassert
						var actual *geos.Geom
						assert.NoError(t, conn.QueryRow(ctx, "select NULL::"+geomType, pgx.QueryResultFormats{format}).Scan(&actual))
						assert.Zero(t, actual)
					})
				}
			})
		}
	})
}

func TestGeometryCodecPointer(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, geomType := range geomTypes {
			t.Run(geomType, func(t *testing.T) {
				for _, format := range []int16{
					pgx.BinaryFormatCode,
					pgx.TextFormatCode,
				} {
					tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) { //nolint:forcetypeassert
						geom := mustNewGeomFromWKT(t, "POINT(1 2)").SetSRID(4326)
						var actual *geos.Geom
						assert.NoError(t, conn.QueryRow(ctx, "select $1::"+geomType, pgx.QueryResultFormats{format}, geom).Scan(&actual))
						assert.Equal(t, geom.ToEWKBWithSRID(), actual.ToEWKBWithSRID())
					})
				}
			})
		}
	})
}

func TestGeometryCodecEncode(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, geomType := range geomTypes {
			t.Run(geomType, func(t *testing.T) {
				for _, format := range []int16{
					pgx.BinaryFormatCode,
					pgx.TextFormatCode,
				} {
					tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) { //nolint:forcetypeassert
						geom := mustNewGeomFromWKT(t, "POINT(1 2)").SetSRID(4326)
						var actual *geos.Geom
						assert.NoError(t, conn.QueryRow(ctx, "select $1::"+geomType, pgx.QueryResultFormats{format}, geom).Scan(&actual))
						assert.Equal(t, geom.ToEWKBWithSRID(), actual.ToEWKBWithSRID())
					})
				}
			})
		}
	})
}

func TestGeometryCodecEncodeNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, geomType := range geomTypes {
			t.Run(geomType, func(t *testing.T) {
				for _, format := range []int16{
					pgx.BinaryFormatCode,
					pgx.TextFormatCode,
				} {
					tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) { //nolint:forcetypeassert
						var nullGeom, actual *geos.Geom
						assert.NoError(t, conn.QueryRow(ctx, "select $1::"+geomType, pgx.QueryResultFormats{format}, nullGeom).Scan(&actual))
						assert.Zero(t, actual)
					})
				}
			})
		}
	})
}

func TestGeometryCodecScan(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, geomType := range geomTypes {
			t.Run(geomType, func(t *testing.T) {
				for _, format := range []int16{
					pgx.BinaryFormatCode,
					pgx.TextFormatCode,
				} {
					tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) { //nolint:forcetypeassert
						original := mustNewGeomFromWKT(t, "POINT(1 2)").SetSRID(4326)
						rows, err := conn.Query(ctx, "select $1::"+geomType, pgx.QueryResultFormats{format}, original)
						assert.NoError(t, err)

						assert.True(t, rows.Next())
						values, err := rows.Values()
						assert.NoError(t, err)
						assert.Equal(t, 1, len(values))
						assert.Equal(t, original.ToEWKBWithSRID(), values[0].(*geos.Geom).ToEWKBWithSRID()) //nolint:forcetypeassert

						assert.False(t, rows.Next())
						assert.NoError(t, rows.Err())
					})
				}
			})
		}
	})
}

func TestGeometryCodecValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		for _, geomType := range geomTypes {
			t.Run(geomType, func(t *testing.T) {
				for _, format := range []int16{
					pgx.BinaryFormatCode,
					pgx.TextFormatCode,
				} {
					tb.(*testing.T).Run(strconv.Itoa(int(format)), func(t *testing.T) { //nolint:forcetypeassert
						var actual *geos.Geom
						assert.NoError(t, conn.QueryRow(ctx, "select ST_SetSRID('POINT(3 4)'::"+geomType+", 4326)", pgx.QueryResultFormats{format}).Scan(&actual))
						expected := mustNewGeomFromWKT(t, "POINT(3 4)").SetSRID(4326)
						assert.Equal(t, expected.ToEWKBWithSRID(), actual.ToEWKBWithSRID())
					})
				}
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
