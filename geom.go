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

// A geometryCodec implements [github.com/jackc/pgx/v5/pgtype.Codec] for
// [*github.com/twpayne/go-geos.Geom] types.
type geometryCodec struct {
	geosContext *geos.Context
}

// A geometryBinaryEncodePlan implements
// [github.com/jackc/pgx/v5/pgtype.EncodePlan] for
// [*github.com/twpayne/go-geos.Geom] types in binary format.
type geometryBinaryEncodePlan struct{}

// A geometryTextEncodePlan implements
// [github.com/jackc/pgx/v5/pgtype.EncodePlan] for
// [*github.com/twpayne/go-geos.Geom] types in text format.
type geometryTextEncodePlan struct{}

// A geometryBinaryScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan]
// for [*github.com/twpayne/go-geos.Geom] types in binary format.
type geometryBinaryScanPlan struct {
	geosContext *geos.Context
}

// A geometryTextScanPlan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan]
// for [*github.com/twpayne/go-geos.Geom] types in text format.
type geometryTextScanPlan struct {
	geosContext *geos.Context
}

// FormatSupported implements
// [github.com/jackc/pgx/v5/pgtype.Codec.FormatSupported].
func (c *geometryCodec) FormatSupported(format int16) bool {
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
func (c *geometryCodec) PreferredFormat() int16 {
	return pgtype.BinaryFormatCode
}

// PlanEncode implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanEncode].
func (c *geometryCodec) PlanEncode(m *pgtype.Map, old uint32, format int16, value any) pgtype.EncodePlan {
	if _, ok := value.(*geos.Geom); !ok {
		return nil
	}
	switch format {
	case pgtype.BinaryFormatCode:
		return geometryBinaryEncodePlan{}
	case pgtype.TextFormatCode:
		return geometryTextEncodePlan{}
	default:
		return nil
	}
}

// PlanScan implements [github.com/jackc/pgx/v5/pgtype.Codec.PlanScan].
func (c *geometryCodec) PlanScan(m *pgtype.Map, old uint32, format int16, target any) pgtype.ScanPlan {
	if _, ok := target.(**geos.Geom); !ok {
		return nil
	}
	switch format {
	case pgx.BinaryFormatCode:
		return geometryBinaryScanPlan{
			geosContext: c.geosContext,
		}
	case pgx.TextFormatCode:
		return geometryTextScanPlan{
			geosContext: c.geosContext,
		}
	default:
		return nil
	}
}

// DecodeDatabaseSQLValue implements
// [github.com/jackc/pgx/v5/pgtype.Codec.DecodeDatabaseSQLValue].
func (c *geometryCodec) DecodeDatabaseSQLValue(m *pgtype.Map, oid uint32, format int16, src []byte) (driver.Value, error) {
	return nil, errors.ErrUnsupported
}

// DecodeValue implements [github.com/jackc/pgx/v5/pgtype.Codec.DecodeValue].
func (c *geometryCodec) DecodeValue(m *pgtype.Map, oid uint32, format int16, src []byte) (any, error) {
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
func (p geometryBinaryEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	geom, ok := value.(*geos.Geom)
	if !ok {
		return buf, errors.ErrUnsupported
	}
	return append(buf, geom.ToEWKBWithSRID()...), nil
}

// Encode implements [github.com/jackc/pgx/v5/pgtype.EncodePlan.Encode].
func (p geometryTextEncodePlan) Encode(value any, buf []byte) (newBuf []byte, err error) {
	geom, ok := value.(*geos.Geom)
	if !ok {
		return buf, errors.ErrUnsupported
	}
	wkb := geom.ToEWKBWithSRID()
	return append(buf, []byte(hex.EncodeToString(wkb))...), nil
}

// Scan implements [github.com/jackc/pgx/v5/pgtype.ScanPlan.Scan].
func (p geometryBinaryScanPlan) Scan(src []byte, target any) error {
	pgeom, ok := target.(**geos.Geom)
	if !ok {
		return errors.ErrUnsupported
	}
	if len(src) == 0 {
		*pgeom = nil
		return nil
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
func (p geometryTextScanPlan) Scan(src []byte, target any) error {
	pgeom, ok := target.(**geos.Geom)
	if !ok {
		return errors.ErrUnsupported
	}
	if len(src) == 0 {
		*pgeom = nil
		return nil
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

// registerGeom registers codecs for [*github.com/twpayne/go-geos.Geom] types on conn.
func registerGeom(ctx context.Context, conn *pgx.Conn, geosContext *geos.Context) error {
	var geographyOID, geometryOID uint32
	err := conn.QueryRow(ctx, "select 'geography'::text::regtype::oid, 'geometry'::text::regtype::oid").Scan(&geographyOID, &geometryOID)
	if err != nil {
		return err
	}

	if geosContext == nil {
		geosContext = geos.DefaultContext
	}

	conn.TypeMap().RegisterType(&pgtype.Type{
		Codec: &geometryCodec{
			geosContext: geosContext,
		},
		Name: "geography",
		OID:  geographyOID,
	})

	conn.TypeMap().RegisterType(&pgtype.Type{
		Codec: &geometryCodec{
			geosContext: geosContext,
		},
		Name: "geometry",
		OID:  geometryOID,
	})

	return nil
}
