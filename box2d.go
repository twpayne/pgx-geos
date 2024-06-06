package pgxgeos

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/twpayne/go-geos"
)

var box2DRegexp = regexp.MustCompile(`\ABOX\((\d+(?:\.\d*)?) (\d+(?:\.\d*)?),(\d+(?:\.\d*)?) (\d+(?:\.\d*)?)\)\z`)

// A box2DCodec implements [github.com/jackc/pgx/v5/pgtype.Codec] for
// [github.com/twpayne/go-geos.Box2D] types.
type box2DCodec struct{}

// A box2DTextEncodePlan implements
// [github.com/jackc/pgx/v5/pgtype.EncodePlan] for
// [github.com/twpayne/go-geos.Box2D] types in text format.
type box2DTextEncodePlan struct{}

// A box2DTextScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan]
// for [github.com/twpayne/go-geos.Box2D] types in text format.
type box2DTextScanPlan struct{}

// FormatSupported implements
// [github.com/jackc/pgx/v5/pgtype.Codec.FormatSupported].
func (c *box2DCodec) FormatSupported(format int16) bool {
	switch format {
	case pgtype.TextFormatCode:
		return true
	default:
		return false
	}
}

// PreferredFormat implements
// [github.com/jackc/pgx/v5/pgtype.Codec.PreferredFormat].
func (c *box2DCodec) PreferredFormat() int16 {
	return pgtype.TextFormatCode
}

// PlanEncode implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanEncode].
func (c *box2DCodec) PlanEncode(m *pgtype.Map, old uint32, format int16, value any) pgtype.EncodePlan {
	switch value.(type) {
	case geos.Box2D, *geos.Box2D:
		switch format {
		case pgtype.TextFormatCode:
			return box2DTextEncodePlan{}
		default:
			return nil
		}
	default:
		return nil
	}
}

// PlanScan implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanScan].
func (c *box2DCodec) PlanScan(m *pgtype.Map, old uint32, format int16, target any) pgtype.ScanPlan {
	if _, ok := target.(*geos.Box2D); !ok {
		return nil
	}
	switch format {
	case pgx.TextFormatCode:
		return box2DTextScanPlan{}
	default:
		return nil
	}
}

// DecodeDatabaseSQLValue implements
// [github.com/jackc/pgx/v5/pgtype.Codec.DecodeDatabaseSQLValue].
func (c *box2DCodec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	return nil, errors.ErrUnsupported
}

// DecodeValue implements [github.com/jackc/pgx/v5/pgtype.Codec.DecodeValue].
func (c *box2DCodec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	switch format {
	case pgtype.TextFormatCode:
		var box2D geos.Box2D
		if err := decodeBox2D(&box2D, src); err != nil {
			return nil, err
		}
		return &box2D, nil
	default:
		return nil, errors.ErrUnsupported
	}
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p box2DTextEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	switch box2D := value.(type) {
	case geos.Box2D:
		return encodeBox2D(&box2D)
	case *geos.Box2D:
		return encodeBox2D(box2D)
	default:
		return nil, errors.ErrUnsupported
	}
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p box2DTextScanPlan) Scan(src []byte, target any) error {
	box2D, ok := target.(*geos.Box2D)
	if !ok {
		return errors.ErrUnsupported
	}
	return decodeBox2D(box2D, src)
}

func decodeBox2D(box2D *geos.Box2D, src []byte) error {
	m := box2DRegexp.FindSubmatch(src)
	if m == nil {
		return fmt.Errorf("%q: invalid BOX2D", string(src))
	}
	box2D.MinX, _ = strconv.ParseFloat(string(m[1]), 64)
	box2D.MinY, _ = strconv.ParseFloat(string(m[2]), 64)
	box2D.MaxX, _ = strconv.ParseFloat(string(m[3]), 64)
	box2D.MaxY, _ = strconv.ParseFloat(string(m[4]), 64)
	return nil
}

func encodeBox2D(box2D *geos.Box2D) ([]byte, error) {
	var builder strings.Builder
	builder.Grow(64)
	builder.WriteString("BOX(")
	builder.WriteString(strconv.FormatFloat(box2D.MinX, 'f', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(box2D.MinY, 'f', -1, 64))
	builder.WriteByte(',')
	builder.WriteString(strconv.FormatFloat(box2D.MaxX, 'f', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(box2D.MaxY, 'f', -1, 64))
	builder.WriteByte(')')
	return []byte(builder.String()), nil
}

// registerBox2D registers codecs for [github.com/twpayne/go-geos.Box2D] types on conn.
func registerBox2D(ctx context.Context, conn *pgx.Conn) error {
	var box2dOID uint32
	if err := conn.QueryRow(ctx, "select 'box2d'::text::regtype::oid").Scan(&box2dOID); err != nil {
		return err
	}

	conn.TypeMap().RegisterType(&pgtype.Type{
		Codec: &box2DCodec{},
		Name:  "box2d",
		OID:   box2dOID,
	})

	return nil
}
