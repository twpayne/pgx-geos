package pgxgeos_test

import (
	"context"
	"testing"

	"github.com/alecthomas/assert/v2"
	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

func TestBox2DCodecPointer(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		box2D := geos.NewBox2D(1, 2, 3, 4)
		var actual geos.Box2D
		assert.NoError(tb, conn.QueryRow(ctx, "select $1::box2d", box2D).Scan(&actual))
		assert.Equal(tb, *box2D, actual)
	})
}

func TestBox2DCodecPointerToPointer(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		box2D := geos.NewBox2D(1, 2, 3, 4)
		var actual *geos.Box2D
		assert.NoError(tb, conn.QueryRow(ctx, "select $1::box2d", box2D).Scan(&actual))
		assert.Equal(tb, *box2D, *actual)
	})
}

func TestBox2DCodecNull(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		actual := geos.NewBox2D(1, 2, 3, 4)
		assert.NoError(tb, conn.QueryRow(ctx, "select NULL::box2d").Scan(&actual))
		assert.Zero(tb, actual)
	})
}

func TestBox2DCodecScan(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		original := geos.NewBox2D(1, 2, 3, 4)
		rows, err := conn.Query(ctx, "select $1::box2d", original)
		assert.NoError(t, err)

		assert.True(t, rows.Next())
		values, err := rows.Values()
		assert.NoError(t, err)
		assert.Equal(t, 1, len(values))
		assert.Equal(t, *original, *values[0].(*geos.Box2D)) //nolint:forcetypeassert

		assert.False(t, rows.Next())
		assert.NoError(t, rows.Err())
	})
}

func TestBox2DCodecValue(t *testing.T) {
	defaultConnTestRunner.RunTest(context.Background(), t, func(ctx context.Context, tb testing.TB, conn *pgx.Conn) {
		tb.Helper()
		box2D := geos.Box2D{MinX: 1, MinY: 2, MaxX: 3, MaxY: 4}
		var actual geos.Box2D
		assert.NoError(tb, conn.QueryRow(ctx, "select $1::box2d", box2D).Scan(&actual))
		assert.Equal(tb, box2D, actual)
	})
}
