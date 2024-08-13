package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	pgxgeos "github.com/twpayne/pgx-geos"
	"io"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/twpayne/go-geos"
)

var (
	connStr = flag.String("conn", "database=pgxgeostest", "connection string")

	create   = flag.Bool("create", false, "create database schema")
	populate = flag.Bool("populate", false, "populate waypoints")
	read     = flag.Bool("read", false, "import waypoint from stdin in GeoJSON format")
	write    = flag.Bool("write", false, "write waypoints to stdout in GeoJSON format")
)

// A Waypoint is a location with an identifier and a name.
type Waypoint struct {
	ID       int
	Name     string
	Geometry *geos.Geom
}

func (w *Waypoint) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		ID         int             `json:"id"`
		Geometry   json.RawMessage `json:"geometry"`
		Properties map[string]any  `json:"properties"`
	}{
		ID:       w.ID,
		Geometry: []byte(w.Geometry.ToGeoJSON(0)),
		Properties: map[string]any{
			"name": w.Name,
		},
	})
}

func (w *Waypoint) UnmarshalJSON(data []byte) error {
	var geoJSONWaypoint struct {
		ID         int             `json:"id"`
		Geometry   json.RawMessage `json:"geometry"`
		Properties map[string]any  `json:"properties"`
	}
	if err := json.Unmarshal(data, &geoJSONWaypoint); err != nil {
		return err
	}
	geom, err := geos.NewGeomFromGeoJSON(string(geoJSONWaypoint.Geometry))
	if err != nil {
		return err
	}
	w.ID = geoJSONWaypoint.ID
	w.Name = geoJSONWaypoint.Properties["name"].(string)
	w.Geometry = geom
	return nil
}

// registerGeos registers required codecs
func registerGeos(ctx context.Context, conn *pgx.Conn) error {
	return pgxgeos.Register(ctx, conn, geos.NewContext())
}

// createDB demonstrates create a PostgreSQL/PostGIS database with a table with
// a geometry column.
func createDB(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		create extension if not exists postgis;
		create table if not exists waypoints (
			id serial primary key,
			name text not null,
			geom geometry(POINT, 4326) not null
		);
	`)
	return err
}

// populateDB demonstrates populating a PostgreSQL/PostGIS database using insert.
func populateDB(ctx context.Context, conn *pgx.Conn) error {
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	for _, waypoint := range []Waypoint{
		{
			Name:     "London",
			Geometry: geos.NewPoint([]float64{0.1275, 51.50722}).SetSRID(4326),
		},
		{
			Name:     "Berlin",
			Geometry: geos.NewPoint([]float64{13.405, 52.52}).SetSRID(4326),
		},
	} {
		if _, err := conn.Exec(ctx, `
			insert into waypoints (name, geom) values ($1, $2);
		`, waypoint.Name, waypoint.Geometry); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

// readGeoJSON demonstrates reading GeoJSON data and inserting it into a
// database with INSERT.
func readGeoJSON(ctx context.Context, conn *pgx.Conn, r io.Reader) error {
	var waypoint Waypoint
	if err := json.NewDecoder(r).Decode(&waypoint); err != nil {
		return err
	}
	_, err := conn.Exec(ctx, `
		insert into waypoints(name, geom) values ($1, $2);
	`, waypoint.Name, waypoint.Geometry)
	return err
}

// writeGeoJSON demonstrates reading data from a database with SELECT and
// writing it as GeoJSON.
func writeGeoJSON(ctx context.Context, conn *pgx.Conn, w io.Writer) error {
	rows, err := conn.Query(ctx, `
		select id, name, geom from waypoints order by id asc;
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var waypoint Waypoint
		if err := rows.Scan(&waypoint.ID, &waypoint.Name, &waypoint.Geometry); err != nil {
			return err
		}
		if err := json.NewEncoder(w).Encode(&waypoint); err != nil {
			return err
		}
	}
	return rows.Err()
}

func run() error {
	ctx := context.Background()

	flag.Parse()

	conn, err := pgx.Connect(ctx, *connStr)
	if err != nil {
		return err
	}
	defer conn.Close(ctx)
	if err := conn.Ping(ctx); err != nil {
		return err
	}
	if *create {
		if err := createDB(ctx, conn); err != nil {
			return err
		}
	}
	if *populate {
		if err := populateDB(ctx, conn); err != nil {
			return err
		}
	}
	if *read {
		if err := readGeoJSON(ctx, conn, os.Stdin); err != nil {
			return err
		}
	}
	if *write {
		if err := registerGeos(ctx, conn); err != nil {
			return err
		}
		if err := writeGeoJSON(ctx, conn, os.Stdout); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
