package postgresqlred

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/XSAM/otelsql"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/lib/pq"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("Journals-service")

type App struct {
	db  *sql.DB
	rdb *redis.Client
}

type Journal struct {
	id    int64
	time  string
	text1 string
	text2 string
}

func NewApp(ctx context.Context) (*App, error) {
	// Postgres
	addressPostgreSQL := os.Getenv("ADDRESS_POSTGRESQL")
	if addressPostgreSQL == "" {
		addressPostgreSQL = "localhost:5432"
	}
	//fmt.Println("ADDRESS_POSTGRESQL=", addressPostgreSQL)

	dsn := fmt.Sprintf("postgresql://postgres:postgres@%s/app?sslmode=disable", addressPostgreSQL)

	driverName, err := otelsql.Register("postgres", otelsql.WithAttributes(attribute.String("db.system", "postgresql")))
	if err != nil {
		log.Println("error otelsql.Register")
		return nil, err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		log.Println("This will not be a connection error, but a DSN parse error or another initialization error")
		return nil, err
	}

	if err := initSchema(ctx, db); err != nil {
		log.Println("Error init DB")
		return nil, err
	}

	addressRedis := os.Getenv("ADDRESS_REDIS")
	if addressRedis == "" {
		addressRedis = "localhost:6379"
	}
	//fmt.Println("ADDRESS_REDIS=", addressRedis)

	// Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: addressRedis,
	})

	if err := redisotel.InstrumentTracing(rdb); err != nil {
		log.Printf("redis tracing instrumentation failed: %v", err)
	}

	return &App{
		db:  db,
		rdb: rdb,
	}, nil
}

func CreateNotesHandler(a *App, dBody []byte, ctx context.Context) {
	span := trace.SpanFromContext(ctx)
	text1 := string(dBody)
	span.SetAttributes(attribute.String("journal.text1", text1))

	ctx, dbSpan := tracer.Start(ctx, "postgres.insert_journal",
		trace.WithAttributes(
			attribute.String("db.operation", "INSERT"),
			attribute.String("journal.text1", text1),
		))

	time.Sleep(300 * time.Millisecond)

	var id int64

	str1 := strings.Split(text1, ";")

	if err := a.db.QueryRowContext(ctx,
		"INSERT INTO journals(time,text1,text2) VALUES($1,$2,$3) RETURNING id", str1[0], str1[1], str1[2]).Scan(&id); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, "insert failed")
		dbSpan.End()

		span.SetStatus(codes.Error, "insert failed")
		log.Println(err)
		return
	}

	dbSpan.SetAttributes(attribute.Int64("journal.id", id))
	dbSpan.End()

	span.SetAttributes(attribute.Int64("journal.id", id))
}

func (a *App) Close() {
	_ = a.db.Close()
	_ = a.rdb.Close()
}

func initSchema(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS journals (
    id BIGSERIAL,
    time TIMESTAMP without time zone PRIMARY KEY,
	text1 TEXT NOT NULL,
	text2 TEXT NOT NULL
);
`)
	return err
}

// Request
func GetJournalHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	span := trace.SpanFromContext(ctx)
	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		span.RecordError(errors.New("missing id parameter"))
		span.SetStatus(codes.Error, "missing id parameter")
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "invalid id parameter")
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	span.SetAttributes(attribute.Int64("journal.id", id))

	// Calling Redis
	cacheKey := fmt.Sprintf("journal:%d", id)
	ctxCache, cacheSpan := tracer.Start(ctx, "redis.cache_lookup",
		trace.WithAttributes(
			attribute.String("cache.key", cacheKey),
		))

	time.Sleep(10 * time.Millisecond)

	// Connect to DB
	dbRes, err := NewApp(context.Background())
	if err != nil {
		log.Println("link fail")
	}

	if b, err := dbRes.rdb.Get(ctxCache, cacheKey).Bytes(); err == nil {
		cacheSpan.SetAttributes(attribute.Bool("cache.hit", true))
		cacheSpan.End()

		span.SetAttributes(attribute.String("cache.status", "hit"))
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Cache", "hit")
		_, _ = w.Write(b)
		fmt.Println("REDIS CONNECTED SUCCESSFULLY")
		return
	}

	cacheSpan.SetAttributes(attribute.Bool("cache.hit", false))
	cacheSpan.End()

	// Calling DB
	ctxDB, dbSpan := tracer.Start(ctx, "postgres.query_journal",
		trace.WithAttributes(
			attribute.Int64("db.journal_id", id),
		))

	var u Journal
	row := dbRes.db.QueryRowContext(ctxDB, "SELECT id, time, text1, text2 FROM journals WHERE id=$1", id)
	fmt.Println("POSTGRESQL CONNECTED SUCCESSFULLY")

	if err := row.Scan(&u.id, &u.time, &u.text1, &u.text2); err != nil {
		dbSpan.RecordError(err)
		dbSpan.SetStatus(codes.Error, "journal not found")
		dbSpan.End()

		span.SetStatus(codes.Error, "journal not found")
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	dbSpan.SetAttributes(attribute.String("journal.name", u.text1))
	dbSpan.End()

	body1 := []string{u.time, u.text1, u.text2}
	body, _ := json.Marshal(body1)
	log.Printf("Calling transaction service for journal %d", u.id)

	// Calling Redis to write cache
	ctx, writeSpan := tracer.Start(ctx, "redis.cache_write",
		trace.WithAttributes(
			attribute.String("cache.key", cacheKey),
			attribute.Int("cache.ttl_seconds", 60),
		))

	//time.Sleep(5 * time.Millisecond)
	if err := dbRes.rdb.Set(ctx, cacheKey, body, 60*time.Second).Err(); err != nil {
		writeSpan.RecordError(err)
		writeSpan.SetStatus(codes.Error, "cache write failed")
	}
	writeSpan.End()

	span.SetAttributes(attribute.String("cache.status", "miss"))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Cache", "miss")
	_, _ = w.Write(body)
}
