package metricdb

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
)

type Metricdb struct {
	conn *pgx.Conn
	ctx  context.Context
}

func CreateDBConnnect(ctx context.Context, connStr string) Metricdb {
	mdb := Metricdb{
		ctx: ctx,
	}
	var err error

	mdb.conn, err = pgx.Connect(mdb.ctx, connStr)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
	}
	return mdb
}

func (mdb *Metricdb) Ping() error {
	if err := mdb.conn.Ping(mdb.ctx); err != nil {
		return err
	}
	return nil
}

func (mdb *Metricdb) DBClose() {
	mdb.conn.Close(mdb.ctx)
}

func (mdb *Metricdb) IsConnActive() bool {
	return mdb.conn != nil
}
