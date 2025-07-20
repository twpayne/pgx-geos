# pgx-geos

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/pgx-geos)](https://pkg.go.dev/github.com/twpayne/pgx-geos)

Package pgx-geos provides [PostGIS](https://postgis.net/) and
[GEOS](https://libgeos.org/) support for
[`github.com/jackc/pgx/v5`](https://pkg.go.dev/github.com/jackc/pgx/v5) via
[`github.com/twpayne/go-geos`](https://pkg.go.dev/github.com/twpayne/go-geos).

## Usage

### Single connection

```go
import (
    // ...

    "github.com/jackc/pgx/v5"
    "github.com/twpayne/go-geos"
    pgxgeos "github.com/twpayne/pgx-geos"
)

// ...

    connectionStr := os.Getenv("DATABASE_URL")
    conn, err := pgx.Connect(context.Background(), connectionStr)
    if err != nil {
        return err
    }
    if err := pgxgeos.Register(ctx, conn, geos.NewContext()); err != nil {
        return err
    }
```

### Connection pool

```go
import (
    // ...

    "github.com/jackc/pgx/v5/pgxpool"
)

// ...

    config, err := pgxpool.ParseConfig(connectionStr)
    if err != nil {
        return err
    }
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        if err := pgxgeos.Register(ctx, conn, geos.NewContext()); err != nil {
            return err
        }
        return nil
    }

    pool, err := pgxpool.NewWithConfig(context.Background(), config)
    if err != nil {
        return err
    }
```

## sqlc

See [the sqlc documentation](https://docs.sqlc.dev/en/latest/reference/datatypes.html#using-github-com-twpayne-go-geos-pgx-v5-only).

## License

MIT