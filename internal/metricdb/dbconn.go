package metricdb

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
)

type Connection struct {
	conn *pgx.Conn
}

func CreateConnnection(ctx context.Context, connStr string) *Connection {
	c := Connection{}
	var err error

	c.conn, err = pgx.Connect(ctx, connStr)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
	}
	return &c
}

func (c *Connection) Begin(ctx context.Context) (pgx.Tx, error) {
	return c.conn.Begin(ctx)
}

func (c *Connection) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	var row pgx.Row
	if len(args) > 0 {
		row = c.conn.QueryRow(ctx, sql, args)
	} else {
		row = c.conn.QueryRow(ctx, sql)
	}
	return row
}

func (c *Connection) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

func (c *Connection) CloseConnection(ctx context.Context) error {
	return c.conn.Close(ctx)
}

func (c *Connection) IsConnClose() bool {
	return (c.conn == nil) || c.conn.IsClosed()
}
