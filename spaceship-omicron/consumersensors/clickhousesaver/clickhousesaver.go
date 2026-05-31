package clickhousesaver

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/ClickHouse/clickhouse-go/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

func ClickhouseConn(address string) clickhouse.Conn {
	connect, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{address},
		Auth: clickhouse.Auth{
			Database: "spaceshipomicron_db",
			Username: "username",
			Password: "password",
		},
		// TLS: &tls.Config{},
	})
	if err != nil {
		panic(err)
	}
	return connect
}

// func ClickhouseConn2() {
// 	ClickhouseConn(clickhouseAddress)
// }

// clickhouse version show
func ClickhouseServerVersion(address string) {
	v, err := ClickhouseConn(address).ServerVersion()
	fmt.Println(v)
	if err != nil {
		panic(err)
	}
}

// clickhouse create/check db
func ClickhouseCreateTableCoords(address string) {
	err2 := ClickhouseConn(address).Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS coords1
(
	Time String,
    x1 Float64,
    y1 Float64,
	z1 Float64,
    Vx1 Float64,
    Vy1 Float64,
	Vz1 Float64,
	Offset Int64,
) ENGINE = MergeTree()
 PRIMARY KEY Time
`)
	if err2 != nil {
		panic(err2)
	}
}

func ClickhouseCreateTableTemps(address string) {
	err2 := ClickhouseConn(address).Exec(context.Background(), `
    CREATE TABLE IF NOT EXISTS temps1
(
	Time String,
    t1 Float64,
    h1 Float64,
	t2 Float64,
    h2 Float64,
    t3 Float64,
	h3 Float64,
    t4 Float64,
    h4 Float64,
	t5 Float64,
    h5 Float64,
    t6 Float64,
	h6 Float64,

	Offset Int64,
) ENGINE = MergeTree()
 PRIMARY KEY Time
`)
	if err2 != nil {
		panic(err2)
	}
}

func ClickhouseSaveCoords(ctx context.Context, messageValue []byte, messageOffset int64) {

	clickhouseAddress := os.Getenv("CLICKHOUSE_ADDRESS")
	if clickhouseAddress == "" {
		clickhouseAddress = "localhost:9000"
	}

	tracer := otel.Tracer("SaveToClickhouseCoords")
	ctx, span := tracer.Start(ctx, "sendToClickhouse")
	span.SetAttributes(attribute.String("clikhouseSend", string(messageValue)))
	defer span.End()

	messageFromKafka := string(messageValue)
	str1 := strings.Split(messageFromKafka, ";")
	err := ClickhouseConn(clickhouseAddress).Exec(context.Background(), fmt.Sprintf("INSERT INTO spaceshipomicron_db.coords1 (Time,x1,y1,z1,Vx1,Vy1,Vz1,Offset) VALUES ('%s','%s','%s','%s','%s','%s','%s','%d')", str1[0], str1[1], str1[2], str1[3], str1[4], str1[5], str1[6], messageOffset))
	if err != nil {
		panic(err)
	}
}

func ClickhouseSaveTemps(ctx context.Context, messageValue []byte, messageOffset int64) {

	clickhouseAddress := os.Getenv("CLICKHOUSE_ADDRESS")
	if clickhouseAddress == "" {
		clickhouseAddress = "localhost:9000"
	}

	tracer := otel.Tracer("SaveToClickhouseTemps")
	ctx, span := tracer.Start(ctx, "sendToClickhouse")
	span.SetAttributes(attribute.String("clikhouseSend", string(messageValue)))
	defer span.End()

	messageFromKafka := string(messageValue)
	str1 := strings.Split(messageFromKafka, ";")
	err := ClickhouseConn(clickhouseAddress).Exec(context.Background(), fmt.Sprintf("INSERT INTO spaceshipomicron_db.temps1 (Time,t1,h1,t2,h2,t3,h3,t4,h4,t5,h5,t6,h6,Offset) VALUES ('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%d')", str1[0], str1[1], str1[2], str1[3], str1[4], str1[5], str1[6], str1[7], str1[8], str1[9], str1[10], str1[11], str1[12], messageOffset))
	if err != nil {
		panic(err)
	}
}
