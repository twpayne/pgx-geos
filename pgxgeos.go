package pgxgeos

import (
	"context"
	"database/sql/driver"
	"encoding/hex"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/twpayne/go-geos"
)

// A codec implements [github.com/jackc/pgx/v5/pgtype.Codec] for
// [*github.com/twpayne/go-geos.Geom] types.
type codec struct {
	geosContext *geos.Context
}

// A binaryEncodePlan implements [github.com/jackc/pgx/v5/pgtype.EncodePlan] for
// [*github.com/twpayne/go-geos.Geom] types in binary format.
type binaryEncodePlan struct{}

// A textEncodePlan implements [github.com/jackc/pgx/v5/pgtype.EncodePlan] for
// [*github.com/twpayne/go-geos.Geom] types in text format.
type textEncodePlan struct{}

// A binaryScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan] for
// [*github.com/twpayne/go-geos.Geom] types in binary format.
type binaryScanPlan struct {
	geosContext *geos.Context
}

// A textScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan] for
// [*github.com/twpayne/go-geos.Geom] types in text format.
type textScanPlan struct {
	geosContext *geos.Context
}

// FormatSupported implements
// [github.com/jackc/pgx/v5/pgtype.Codec.FormatSupported].
func (c *codec) FormatSupported(format int16) bool {
	switch format {
	case pgtype.BinaryFormatCode:
		return true
	case pgtype.TextFormatCode:
		return true
	default:
		return false
	}
}

// PreferredFormat implements
// [github.com/jackc/pgx/v5/pgtype.Codec.PreferredFormat].
func (c *codec) PreferredFormat() int16 {
	return pgtype.BinaryFormatCode
}

// PlanEncode implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanEncode].
func (c *codec) PlanEncode(m *pgtype.Map, old uint32, format int16, value any) pgtype.EncodePlan {
	if _, ok := value.(*geos.Geom); !ok {
		return nil
	}
	switch format {
	case pgtype.BinaryFormatCode:
		return binaryEncodePlan{}
	case pgtype.TextFormatCode:
		return textEncodePlan{}
	default:
		return nil
	}
}

// PlanScan implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanScan].
func (c *codec) PlanScan(m *pgtype.Map, old uint32, format int16, target any) pgtype.ScanPlan {
	if _, ok := target.(**geos.Geom); !ok {
		return nil
	}
	switch format {
	case pgx.BinaryFormatCode:
		return &binaryScanPlan{
			geosContext: c.geosContext,
		}
	case pgx.TextFormatCode:
		return &textScanPlan{
			geosContext: c.geosContext,
		}
	default:
		return nil
	}
}

// DecodeDatabaseSQLValue implements
// [github.com/jackc/pgx/v5/pgtype.Codec.DecodeDatabaseSQLValue].
func (c *codec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	return nil, errors.ErrUnsupported
}

// DecodeValue implements [github.com/jackc/pgx/v5/pgtype.Codec.DecodeValue].
func (c *codec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
	switch format {
	case pgtype.TextFormatCode:
		var err error
		src, err = hex.DecodeString(string(src))
		if err != nil {
			return nil, err
		}
		fallthrough
	case pgtype.BinaryFormatCode:
		geom, err := c.geosContext.NewGeomFromWKB(src)
		return geom, err
	default:
		return nil, errors.ErrUnsupported
	}
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p binaryEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	geom, ok := value.(*geos.Geom)
	if !ok {
		return buf, errors.ErrUnsupported
	}
	return append(buf, geom.ToEWKBWithSRID()...), nil
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p textEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	geom, ok := value.(*geos.Geom)
	if !ok {
		return buf, errors.ErrUnsupported
	}
	wkb := geom.ToEWKBWithSRID()
	return append(buf, []byte(hex.EncodeToString(wkb))...), nil
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p *binaryScanPlan) Scan(src []byte, target any) error {
	pgeom, ok := target.(**geos.Geom)
	if !ok {
		return errors.ErrUnsupported
	}
	geom, err := p.geosContext.NewGeomFromWKB(src)
	if err != nil {
		return err
	}
	(*pgeom).Destroy()
	*pgeom = geom
	return nil
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p *textScanPlan) Scan(src []byte, target any) error {
	pgeom, ok := target.(**geos.Geom)
	if !ok {
		return errors.ErrUnsupported
	}
	var err error
	src, err = hex.DecodeString(string(src))
	if err != nil {
		return err
	}
	geom, err := p.geosContext.NewGeomFromWKB(src)
	if err != nil {
		return err
	}
	(*pgeom).Destroy()
	*pgeom = geom
	return nil
}

// Register registers a codec for [*github.com/twpayne/go-geos.Geom] types on
// conn.
func Register(ctx context.Context, conn *pgx.Conn, geosContext *geos.Context) error {
	var oid uint32
	err := conn.QueryRow(ctx, "select 'geometry'::text::regtype::oid").Scan(&oid)
	if err != nil {
		return err
	}

	if geosContext == nil {
		geosContext = geos.DefaultContext
	}

	conn.TypeMap().RegisterType(&pgtype.Type{
		Codec: &codec{
			geosContext: geosContext,
		},
		Name: "geometry",
		OID:  oid,
	})

	return nil
}
