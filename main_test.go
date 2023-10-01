package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	replicasecursor "github.com/replicase/pgcapture/pkg/cursor"
	replicasepb "github.com/replicase/pgcapture/pkg/pb"
	replicasesink "github.com/replicase/pgcapture/pkg/sink"
	replicasesource "github.com/replicase/pgcapture/pkg/source"
	pgcapturecursor "github.com/rueian/pgcapture/pkg/cursor"
	pgcapturesink "github.com/rueian/pgcapture/pkg/sink"
	pgcapturesource "github.com/rueian/pgcapture/pkg/source"
	"github.com/sirupsen/logrus"
)

var (
	showModePrefix = flag.Bool("showModePrefix", false, "show mode prefix")
	modes          = flag.String("modes", "default,pipeline,native", "postgres sink mode to benchmark")
)

func BenchmarkPostgresSink(b *testing.B) {
	sourceConnStr := "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	sinkConnStr := "postgres://postgres@localhost:5433/postgres?sslmode=disable"
	benches := []struct {
		name  string
		bench func(b *testing.B, sourceConnStr string, sinkConnStr string, totalTXSize []int, pipelineBatchTXSize int)
	}{
		{
			name:  "single_insert",
			bench: benchmarkPostgresSinkSingleInsert,
		},
	}
	for _, be := range benches {
		b.Run(be.name, func(b *testing.B) {
			be.bench(b, sourceConnStr, sinkConnStr, []int{100, 1000, 10000}, 100)
		})
	}
}

const testSlot = "test_slot"

func resetDatabase(sourceConn, sinkConn *pgx.Conn) error {
	createTable := func(conn *pgx.Conn) error {
		if _, err := conn.Exec(context.Background(), "CREATE TABLE t1 (f1 int primary key, f2 text)"); err != nil {
			return err
		}
		return nil
	}

	if _, err := sinkConn.Exec(context.Background(), "DROP subscription IF EXISTS test_sub"); err != nil {
		return err
	}
	if _, err := sourceConn.Exec(context.Background(), "DROP publication IF EXISTS test_pub"); err != nil {
		return err
	}

	if _, err := sourceConn.Exec(context.Background(), fmt.Sprintf("SELECT pg_drop_replication_slot('%s')", testSlot)); err != nil {
		var pge *pgconn.PgError
		if !errors.As(err, &pge) || pge.Code != "42704" {
			return err
		}
	}
	if _, err := sinkConn.Exec(context.Background(), fmt.Sprintf("SELECT pg_drop_replication_slot('%s')", testSlot)); err != nil {
		var pge *pgconn.PgError
		if !errors.As(err, &pge) || pge.Code != "42704" {
			return err
		}
	}

	if _, err := sourceConn.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); err != nil {
		return err
	}
	if _, err := sinkConn.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); err != nil {
		return err
	}

	if err := createTable(sourceConn); err != nil {
		return err
	}
	if err := createTable(sinkConn); err != nil {
		return err
	}
	return nil
}

func benchmarkPostgresSinkSingleInsert(b *testing.B, sourceConnStr string, sinkConnStr string, totalTXSize []int, pipelineBatchTXSize int) {
	const testSlot = "test_slot"
	sourceConn, err := pgx.Connect(context.Background(), sourceConnStr)
	if err != nil {
		b.Fatal(err)
	}
	defer sourceConn.Close(context.Background())
	sinkConn, err := pgx.Connect(context.Background(), sinkConnStr)
	if err != nil {
		b.Fatal(err)
	}
	defer sinkConn.Close(context.Background())

	benches := []struct {
		name  string
		setup func() error
		bench func(b *testing.B, txSize int) error
	}{
		{
			name: "default",
			setup: func() error {
				return resetDatabase(sourceConn, sinkConn)
			},
			bench: func(b *testing.B, txSize int) error {
				b.StopTimer()
				source := pgcapturesource.PGXSource{
					SetupConnStr:      sourceConnStr,
					ReplConnStr:       sourceConnStr + "&replication=database",
					ReplSlot:          testSlot,
					CreateSlot:        true,
					CreatePublication: true,
					DecodePlugin:      "pgoutput",
				}
				changes, err := source.Capture(pgcapturecursor.Checkpoint{})
				if err != nil {
					return err
				}
				defer source.Stop()
				_, err = commitSingleInsert(sourceConn, txSize)
				if err != nil {
					return err
				}
				sink := pgcapturesink.PGXSink{
					ConnStr:  sinkConnStr,
					SourceID: "test",
				}
				if _, err = sink.Setup(); err != nil {
					return err
				}
				defer sink.Stop()
				b.StartTimer()

				committed := sink.Apply(changes)
				for i := 0; i < txSize; i++ {
					<-committed
				}
				return nil
			},
		},
		{
			name: "pipeline",
			setup: func() error {
				return resetDatabase(sourceConn, sinkConn)
			},
			bench: func(b *testing.B, txSize int) error {
				b.StopTimer()
				source := replicasesource.PGXSource{
					SetupConnStr:      sourceConnStr,
					ReplConnStr:       sourceConnStr + "&replication=database",
					ReplSlot:          testSlot,
					CreateSlot:        true,
					CreatePublication: true,
					DecodePlugin:      "pgoutput",
				}
				changes, err := source.Capture(replicasecursor.Checkpoint{})
				if err != nil {
					return err
				}
				defer source.Stop()
				_, err = commitSingleInsert(sourceConn, txSize)
				if err != nil {
					return err
				}
				sink := replicasesink.PGXSink{
					ConnStr:     sinkConnStr,
					SourceID:    "test",
					BatchTXSize: pipelineBatchTXSize,
				}
				if _, err = sink.Setup(); err != nil {
					return err
				}
				defer sink.Stop()
				b.StartTimer()

				committed := sink.Apply(changes)
				for i := 0; i < txSize; i++ {
					<-committed
				}
				return nil
			},
		},
		{
			name: "native",
			setup: func() error {
				if err := resetDatabase(sourceConn, sinkConn); err != nil {
					return err
				}
				if _, err = sourceConn.Exec(context.Background(), "CREATE PUBLICATION test_pub FOR ALL TABLES"); err != nil {
					return err
				}
				if _, err = sinkConn.Exec(context.Background(), "CREATE subscription test_sub connection 'dbname=postgres host=pg_source_14 user=postgres port=5432' publication test_pub with (binary = true)"); err != nil {
					return err
				}
				if _, err = sinkConn.Exec(context.Background(), "ALTER SUBSCRIPTION test_sub DISABLE"); err != nil {
					return err
				}
				return nil
			},
			bench: func(b *testing.B, txSize int) error {
				b.StopTimer()
				source := replicasesource.PGXSource{
					SetupConnStr:      sinkConnStr,
					ReplConnStr:       sinkConnStr + "&replication=database",
					ReplSlot:          testSlot,
					CreateSlot:        true,
					CreatePublication: true,
					DecodePlugin:      "pgoutput",
				}
				changes, err := source.Capture(replicasecursor.Checkpoint{})
				if err != nil {
					return err
				}
				defer source.Stop()
				lastT, err := commitSingleInsert(sourceConn, txSize)
				if err != nil {
					return err
				}
				b.StartTimer()

				if _, err = sinkConn.Exec(context.Background(), "ALTER SUBSCRIPTION test_sub ENABLE"); err != nil {
					return err
				}
				for msg := range changes {
					if change, ok := msg.Message.Type.(*replicasepb.Message_Change); ok {
						var (
							f1 pgtype.Int4
							f2 pgtype.Text
						)
						if err = pgTypeMap.Scan(pgtype.Int4OID, pgtype.BinaryFormatCode, change.Change.New[0].GetBinary(), &f1); err != nil {
							return err
						}
						if err = pgTypeMap.Scan(pgtype.TextOID, pgtype.BinaryFormatCode, change.Change.New[1].GetBinary(), &f2); err != nil {
							return err
						}
						if f1.Int32 == lastT.F1 && f2.String == lastT.F2 {
							break
						}
					}
				}
				return nil
			},
		},
	}

	logrus.SetOutput(io.Discard)
	b.ResetTimer()
	modes := strings.Split(*modes, ",")
	for _, size := range totalTXSize {
		for _, be := range benches {
			for _, mode := range modes {
				if mode == be.name {
					desc := fmt.Sprintf("txSize_%d", size)
					if *showModePrefix {
						desc = fmt.Sprintf("%s_txSize_%d", mode, size)
					}
					b.Run(desc, func(b *testing.B) {
						for i := 0; i < b.N; i++ {
							b.StopTimer()
							if err := be.setup(); err != nil {
								b.Fatal(err)
							}
							b.StartTimer()
							if err := be.bench(b, size); err != nil {
								b.Fatal(err)
							}
						}
					})
				}
			}
		}
	}
}

var pgTypeMap = pgtype.NewMap()

type T struct {
	F1 int32
	F2 string
}

func commitSingleInsert(conn *pgx.Conn, txSize int) (*T, error) {
	var lastT T
	for i := 0; i < txSize; i++ {
		tx, err := conn.Begin(context.Background())
		if err != nil {
			return nil, err
		}
		lastT = T{
			F1: int32(i),
			F2: fmt.Sprintf("test_%d", i),
		}
		if _, err = tx.Exec(context.Background(), "insert into t1 (f1, f2) values ($1, $2)", lastT.F1, lastT.F2); err != nil {
			return nil, err
		}
		if err := tx.Commit(context.Background()); err != nil {
			return nil, err
		}
	}
	return &lastT, nil
}
