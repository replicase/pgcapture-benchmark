package main

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	replicasecursor "github.com/replicase/pgcapture/pkg/cursor"
	replicasepb "github.com/replicase/pgcapture/pkg/pb"
	replicasink "github.com/replicase/pgcapture/pkg/sink"
	replicasesource "github.com/replicase/pgcapture/pkg/source"
	"github.com/rueian/pgcapture/pkg/cursor"
	"github.com/rueian/pgcapture/pkg/pb"
	"github.com/rueian/pgcapture/pkg/sink"
	"github.com/rueian/pgcapture/pkg/source"
	"github.com/sirupsen/logrus"
	"io"
	"testing"
	"time"
)

func BenchmarkPostgresSink(b *testing.B) {
	connStr := "postgres://postgres@localhost:5432/postgres?sslmode=disable"
	benches := []struct {
		name  string
		bench func(b *testing.B, connStr string, totalTXSize []int, pipelineBatchTXSize int)
	}{
		{
			name:  "single_insert",
			bench: benchmarkPostgresSinkSingleInsert,
		},
		{
			name:  "single_insert_update_delete",
			bench: benchmarkPostgresSinkSingleInsertUpdateDelete,
		},
	}
	for _, be := range benches {
		b.Run(be.name, func(b *testing.B) {
			be.bench(b, connStr, []int{100, 500, 1000, 2000, 3000, 4000, 5000, 6000, 7000, 8000, 9000, 10000}, 100)
		})
	}
}

func benchmarkPostgresSinkSingleInsert(b *testing.B, connStr string, totalTXSize []int, pipelineBatchTXSize int) {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close(context.Background())

	logrus.SetOutput(io.Discard)

	b.ResetTimer()
	for _, size := range totalTXSize {
		b.Run(fmt.Sprintf("default_txSize_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				if err := resetDatabase(conn); err != nil {
					b.Fatal(err)
				}
				changes, err := prepareDefaultChangesWithSingleInsert(size)
				if err != nil {
					b.Fatal(err)
				}
				sink := sink.PGXSink{
					ConnStr:  connStr,
					SourceID: "test",
				}
				if _, err = sink.Setup(); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				committed := sink.Apply(changes)
				for i := 0; i < size; i++ {
					<-committed
				}
				sink.Stop()
			}
		})
		b.Run(fmt.Sprintf("pipeline_txSize_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				if err = resetDatabase(conn); err != nil {
					b.Fatal(err)
				}
				changes, err := preparePipelineChangesWithSingleInsert(size)
				if err != nil {
					b.Fatal(err)
				}

				sink := replicasink.PGXSink{
					ConnStr:     connStr,
					SourceID:    "test",
					BatchTXSize: pipelineBatchTXSize,
				}
				if _, err = sink.Setup(); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				committed := sink.Apply(changes)
				for i := 0; i < size; i++ {
					<-committed
				}
				sink.Stop()
			}
		})
	}
}

func benchmarkPostgresSinkSingleInsertUpdateDelete(b *testing.B, connStr string, totalTXSize []int, pipelineBatchTXSize int) {
	conn, err := pgx.Connect(context.Background(), connStr)
	if err != nil {
		b.Fatal(err)
	}
	defer conn.Close(context.Background())

	logrus.SetOutput(io.Discard)

	b.ResetTimer()
	for _, size := range totalTXSize {
		b.Run(fmt.Sprintf("default_txSize_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				if err := resetDatabase(conn); err != nil {
					b.Fatal(err)
				}
				changes, err := prepareDefaultChangesWithSingleInsertUpdateDelete(size)
				if err != nil {
					b.Fatal(err)
				}
				sink := sink.PGXSink{
					ConnStr:  connStr,
					SourceID: "test",
				}
				if _, err = sink.Setup(); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				committed := sink.Apply(changes)
				for i := 0; i < size; i++ {
					<-committed
				}
				sink.Stop()
			}
		})
		b.Run(fmt.Sprintf("pipeline_txSize_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				if err = resetDatabase(conn); err != nil {
					b.Fatal(err)
				}
				changes, err := preparePipelineChangesWithSingleInsertUpdateDelete(size)
				if err != nil {
					b.Fatal(err)
				}
				sink := replicasink.PGXSink{
					ConnStr:     connStr,
					SourceID:    "test",
					BatchTXSize: pipelineBatchTXSize,
				}
				if _, err = sink.Setup(); err != nil {
					b.Fatal(err)
				}
				b.StartTimer()
				committed := sink.Apply(changes)
				for i := 0; i < size; i++ {
					<-committed
				}
				sink.Stop()
			}
		})
	}
}

var pgTypeMap = pgtype.NewMap()

func resetDatabase(conn *pgx.Conn) error {
	if _, err := conn.Exec(context.Background(), "DROP SCHEMA public CASCADE; CREATE SCHEMA public"); err != nil {
		return err
	}
	if _, err := conn.Exec(context.Background(), "DROP EXTENSION IF EXISTS pgcapture"); err != nil {
		return err
	}
	if _, err := conn.Exec(context.Background(), "CREATE TABLE t1 (f1 int primary key, f2 text)"); err != nil {
		return err
	}
	return nil
}

func prepareDefaultChangesWithSingleInsert(txSize int) (chan source.Change, error) {
	var (
		now     = time.Now()
		lsn     uint64
		changes = make(chan source.Change, txSize*3)
	)
	for i := 0; i < txSize; i++ {
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message:    &pb.Message{Type: &pb.Message_Begin{Begin: &pb.Begin{}}},
		}

		f1, err := pgTypeMap.Encode(pgtype.Int4OID, pgtype.BinaryFormatCode, i, nil)
		if err != nil {
			return nil, err
		}
		f2, err := pgTypeMap.Encode(pgtype.TextOID, pgtype.BinaryFormatCode, "test", nil)
		if err != nil {
			return nil, err
		}
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message: &pb.Message{Type: &pb.Message_Change{Change: &pb.Change{
				Op:     pb.Change_INSERT,
				Schema: "public",
				Table:  "t1",
				New: []*pb.Field{
					{Name: "f1", Oid: 23, Value: &pb.Field_Binary{Binary: f1}},
					{Name: "f2", Oid: 25, Value: &pb.Field_Binary{Binary: f2}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message:    &pb.Message{Type: &pb.Message_Commit{Commit: &pb.Commit{CommitTime: uint64(now.Unix())}}},
		}
	}
	return changes, nil
}

func prepareDefaultChangesWithSingleInsertUpdateDelete(txSize int) (chan source.Change, error) {
	var (
		now     = time.Now()
		lsn     uint64
		changes = make(chan source.Change, txSize*5)
	)
	for i := 0; i < txSize; i++ {
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message:    &pb.Message{Type: &pb.Message_Begin{Begin: &pb.Begin{}}},
		}

		f1, err := pgTypeMap.Encode(pgtype.Int4OID, pgtype.BinaryFormatCode, i, nil)
		if err != nil {
			return nil, err
		}
		f2, err := pgTypeMap.Encode(pgtype.TextOID, pgtype.BinaryFormatCode, "test", nil)
		if err != nil {
			return nil, err
		}
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message: &pb.Message{Type: &pb.Message_Change{Change: &pb.Change{
				Op:     pb.Change_INSERT,
				Schema: "public",
				Table:  "t1",
				New: []*pb.Field{
					{Name: "f1", Oid: 23, Value: &pb.Field_Binary{Binary: f1}},
					{Name: "f2", Oid: 25, Value: &pb.Field_Binary{Binary: f2}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message: &pb.Message{Type: &pb.Message_Change{Change: &pb.Change{
				Op:     pb.Change_UPDATE,
				Schema: "public",
				Table:  "t1",
				New: []*pb.Field{
					{Name: "f1", Oid: 23, Value: &pb.Field_Binary{Binary: f1}},
					{Name: "f2", Oid: 25, Value: &pb.Field_Binary{Binary: f2}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message: &pb.Message{Type: &pb.Message_Change{Change: &pb.Change{
				Op:     pb.Change_DELETE,
				Schema: "public",
				Table:  "t1",
				Old: []*pb.Field{
					{Name: "f1", Oid: 23, Value: &pb.Field_Binary{Binary: f1}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- source.Change{
			Checkpoint: cursor.Checkpoint{LSN: lsn},
			Message:    &pb.Message{Type: &pb.Message_Commit{Commit: &pb.Commit{CommitTime: uint64(now.Unix())}}},
		}
	}
	return changes, nil
}

func preparePipelineChangesWithSingleInsert(txSize int) (chan replicasesource.Change, error) {
	var (
		now     = time.Now()
		lsn     uint64
		changes = make(chan replicasesource.Change, txSize*3)
	)
	for i := 0; i < txSize; i++ {
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message:    &replicasepb.Message{Type: &replicasepb.Message_Begin{Begin: &replicasepb.Begin{}}},
		}

		f1, err := pgTypeMap.Encode(pgtype.Int4OID, pgtype.BinaryFormatCode, i, nil)
		if err != nil {
			return nil, err
		}
		f2, err := pgTypeMap.Encode(pgtype.TextOID, pgtype.BinaryFormatCode, "test", nil)
		if err != nil {
			return nil, err
		}
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message: &replicasepb.Message{Type: &replicasepb.Message_Change{Change: &replicasepb.Change{
				Op:     replicasepb.Change_INSERT,
				Schema: "public",
				Table:  "t1",
				New: []*replicasepb.Field{
					{Name: "f1", Oid: 23, Value: &replicasepb.Field_Binary{Binary: f1}},
					{Name: "f2", Oid: 25, Value: &replicasepb.Field_Binary{Binary: f2}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message:    &replicasepb.Message{Type: &replicasepb.Message_Commit{Commit: &replicasepb.Commit{CommitTime: uint64(now.Unix())}}},
		}
	}
	return changes, nil
}

func preparePipelineChangesWithSingleInsertUpdateDelete(txSize int) (chan replicasesource.Change, error) {
	var (
		now     = time.Now()
		lsn     uint64
		changes = make(chan replicasesource.Change, txSize*5)
	)
	for i := 0; i < txSize; i++ {
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message:    &replicasepb.Message{Type: &replicasepb.Message_Begin{Begin: &replicasepb.Begin{}}},
		}

		f1, err := pgTypeMap.Encode(pgtype.Int4OID, pgtype.BinaryFormatCode, i, nil)
		if err != nil {
			return nil, err
		}
		f2, err := pgTypeMap.Encode(pgtype.TextOID, pgtype.BinaryFormatCode, "test", nil)
		if err != nil {
			return nil, err
		}
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message: &replicasepb.Message{Type: &replicasepb.Message_Change{Change: &replicasepb.Change{
				Op:     replicasepb.Change_INSERT,
				Schema: "public",
				Table:  "t1",
				New: []*replicasepb.Field{
					{Name: "f1", Oid: 23, Value: &replicasepb.Field_Binary{Binary: f1}},
					{Name: "f2", Oid: 25, Value: &replicasepb.Field_Binary{Binary: f2}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message: &replicasepb.Message{Type: &replicasepb.Message_Change{Change: &replicasepb.Change{
				Op:     replicasepb.Change_UPDATE,
				Schema: "public",
				Table:  "t1",
				New: []*replicasepb.Field{
					{Name: "f1", Oid: 23, Value: &replicasepb.Field_Binary{Binary: f1}},
					{Name: "f2", Oid: 25, Value: &replicasepb.Field_Binary{Binary: f2}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message: &replicasepb.Message{Type: &replicasepb.Message_Change{Change: &replicasepb.Change{
				Op:     replicasepb.Change_DELETE,
				Schema: "public",
				Table:  "t1",
				Old: []*replicasepb.Field{
					{Name: "f1", Oid: 23, Value: &replicasepb.Field_Binary{Binary: f1}},
				},
			}}},
		}
		lsn++
		now.Add(time.Second)
		changes <- replicasesource.Change{
			Checkpoint: replicasecursor.Checkpoint{LSN: lsn},
			Message:    &replicasepb.Message{Type: &replicasepb.Message_Commit{Commit: &replicasepb.Commit{CommitTime: uint64(now.Unix())}}},
		}
	}
	return changes, nil
}
