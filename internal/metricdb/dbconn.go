package metricdb

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4"
)

type Metricdb struct {
	Conn *pgx.Conn
	ctx  context.Context
}

func CreateDBConnnect(ctx context.Context, connStr string) Metricdb {
	mdb := Metricdb{
		ctx: ctx,
	}
	var err error

	mdb.Conn, err = pgx.Connect(mdb.ctx, connStr)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", err)
	}
	return mdb
}

func (mdb *Metricdb) GetDBContext() context.Context {
	return mdb.ctx
}

func (mdb *Metricdb) Ping() error {
	if err := mdb.Conn.Ping(mdb.ctx); err != nil {
		return err
	}
	return nil
}

func (mdb *Metricdb) DBClose() {
	mdb.Conn.Close(mdb.ctx)
}

func (mdb *Metricdb) IsConnActive() bool {
	return mdb.Conn != nil
}
