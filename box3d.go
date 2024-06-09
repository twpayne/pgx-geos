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

var box3DRegexp = regexp.MustCompile(`\ABOX3D\((\d+(?:\.\d*)?) (\d+(?:\.\d*)?) (\d+(?:\.\d*)?),(\d+(?:\.\d*)?) (\d+(?:\.\d*)?) (\d+(?:\.\d*)?)\)\z`)

// A box3DCodec implements [github.com/jackc/pgx/v5/pgtype.Codec] for
// [github.com/twpayne/go-geos.Box3D] types.
type box3DCodec struct{}

// A box3DTextEncodePlan implements
// [github.com/jackc/pgx/v5/pgtype.EncodePlan] for
// [github.com/twpayne/go-geos.Box3D] types in text format.
type box3DTextEncodePlan struct{}

// A box3DTextScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan]
// for [github.com/twpayne/go-geos.Box3D] types in text format.
type box3DTextScanPlan struct{}

// FormatSupported implements
// [github.com/jackc/pgx/v5/pgtype.Codec.FormatSupported].
func (c *box3DCodec) FormatSupported(format int16) bool {
	switch format {
	case pgtype.TextFormatCode:
		return true
	default:
		return false
	}
}

// PreferredFormat implements
// [github.com/jackc/pgx/v5/pgtype.Codec.PreferredFormat].
func (c *box3DCodec) PreferredFormat() int16 {
	return pgtype.TextFormatCode
}

// PlanEncode implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanEncode].
func (c *box3DCodec) PlanEncode(m *pgtype.Map, old uint32, format int16, value any) pgtype.EncodePlan {
	switch value.(type) {
	case geos.Box3D, *geos.Box3D:
		switch format {
		case pgtype.TextFormatCode:
			return box3DTextEncodePlan{}
		default:
			return nil
		}
	default:
		return nil
	}
}

// PlanScan implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanScan].
func (c *box3DCodec) PlanScan(m *pgtype.Map, old uint32, format int16, target any) pgtype.ScanPlan {
	if _, ok := target.(*geos.Box3D); !ok {
		return nil
	}
	switch format {
	case pgx.TextFormatCode:
		return box3DTextScanPlan{}
	default:
		return nil
	}
}

// DecodeDatabaseSQLValue implements
// [github.com/jackc/pgx/v5/pgtype.Codec.DecodeDatabaseSQLValue].
func (c *box3DCodec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	return nil, errors.ErrUnsupported
}

// DecodeValue implements [github.com/jackc/pgx/v5/pgtype.Codec.DecodeValue].
func (c *box3DCodec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	switch format {
	case pgtype.TextFormatCode:
		var box3D geos.Box3D
		if err := decodeBox3D(&box3D, src); err != nil {
			return nil, err
		}
		return &box3D, nil
	default:
		return nil, errors.ErrUnsupported
	}
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p box3DTextEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	switch box3D := value.(type) {
	case geos.Box3D:
		return encodeBox3D(&box3D)
	case *geos.Box3D:
		return encodeBox3D(box3D)
	default:
		return nil, errors.ErrUnsupported
	}
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p box3DTextScanPlan) Scan(src []byte, target any) error {
	box3D, ok := target.(*geos.Box3D)
	if !ok {
		return errors.ErrUnsupported
	}
	return decodeBox3D(box3D, src)
}

func decodeBox3D(box3D *geos.Box3D, src []byte) error {
	m := box3DRegexp.FindSubmatch(src)
	if m == nil {
		return fmt.Errorf("%q: invalid BOX3D", string(src))
	}
	box3D.MinX, _ = strconv.ParseFloat(string(m[1]), 64)
	box3D.MinY, _ = strconv.ParseFloat(string(m[2]), 64)
	box3D.MinZ, _ = strconv.ParseFloat(string(m[3]), 64)
	box3D.MaxX, _ = strconv.ParseFloat(string(m[4]), 64)
	box3D.MaxY, _ = strconv.ParseFloat(string(m[5]), 64)
	box3D.MaxZ, _ = strconv.ParseFloat(string(m[6]), 64)
	return nil
}

func encodeBox3D(box3D *geos.Box3D) ([]byte, error) {
	var builder strings.Builder
	builder.Grow(128)
	builder.WriteString("BOX3D(")
	builder.WriteString(strconv.FormatFloat(box3D.MinX, 'f', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(box3D.MinY, 'f', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(box3D.MinZ, 'f', -1, 64))
	builder.WriteByte(',')
	builder.WriteString(strconv.FormatFloat(box3D.MaxX, 'f', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(box3D.MaxY, 'f', -1, 64))
	builder.WriteByte(' ')
	builder.WriteString(strconv.FormatFloat(box3D.MaxZ, 'f', -1, 64))
	builder.WriteByte(')')
	return []byte(builder.String()), nil
}

// registerBox3D registers codecs for [github.com/twpayne/go-geos.Box3D] types on conn.
func registerBox3D(ctx context.Context, conn *pgx.Conn) error {
	var box3dOID uint32
	if err := conn.QueryRow(ctx, "select 'box3d'::text::regtype::oid").Scan(&box3dOID); err != nil {
		return err
	}

	conn.TypeMap().RegisterType(&pgtype.Type{
		Codec: &box3DCodec{},
		Name:  "box3d",
		OID:   box3dOID,
	})

	return nil
}
