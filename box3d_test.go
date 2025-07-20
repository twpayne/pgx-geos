package pgxgeos_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

func TestBox3DCodecPointer(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		box3D := geos.NewBox3D(1, 2, 3, 4, 5, 6)
		var actual geos.Box3D
		assert.NoError(tb, conn.QueryRow(ctx, "select $1::box3d", box3D).Scan(&actual))
		assert.Equal(tb, *box3D, actual)
	})
}

func TestBox3DCodecPointerToPointer(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		box3D := geos.NewBox3D(1, 2, 3, 4, 5, 6)
		var actual *geos.Box3D
		assert.NoError(tb, conn.QueryRow(ctx, "select $1::box3d", box3D).Scan(&actual))
		assert.Equal(tb, *box3D, *actual)
	})
}

func TestBox3DCodecNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		actual := geos.NewBox3D(1, 2, 3, 4, 5, 6)
		assert.NoError(tb, conn.QueryRow(ctx, "select NULL::box3d").Scan(&actual))
		assert.Zero(tb, actual)
	})
}

func TestBox3DCodecScan(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		original := geos.NewBox3D(1, 2, 3, 4, 5, 6)
		rows, err := conn.Query(ctx, "select $1::box3d", original)
		assert.NoError(t, err)

		assert.True(t, rows.Next())
		values, err := rows.Values()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(values))
		assert.Equal(t, *original, *values[0].(*geos.Box3D)) //nolint:forcetypeassert

		assert.False(t, rows.Next())
		assert.NoError(t, rows.Err())
	})
}

func TestBox3DCodecValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		box3D := geos.Box3D{MinX: 1, MinY: 2, MinZ: 3, MaxX: 4, MaxY: 5, MaxZ: 6}
		var actual geos.Box3D
		assert.NoError(tb, conn.QueryRow(ctx, "select $1::box3d", box3D).Scan(&actual))
		assert.Equal(tb, box3D, actual)
	})
}
