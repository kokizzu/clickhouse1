package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	chBuffer "github.com/kokizzu/ch-timed-buffer"
	"github.com/kokizzu/gotro/D/Ch"
	"github.com/kokizzu/gotro/L"
	"github.com/kokizzu/gotro/S"
	"golang.org/x/sync/errgroup"
)

const debug = true

type ClickhouseConf struct {
	User   string
	Pass   string
	DB     string
	Host   string
	Port   int
	UseSsl bool
}

func (c ClickhouseConf) Connect() (a *Ch.Adapter, err error) {
	hostPort := fmt.Sprintf("%s:%d", c.Host, c.Port)
	conf := &clickhouse.Options{
		Addr: []string{hostPort},
		Auth: clickhouse.Auth{
			Database: c.DB,
			Username: c.User,
			Password: c.Pass,
		},
		Settings: clickhouse.Settings{
			`max_execution_time`:                    60,
			`allow_experimental_lightweight_delete`: 1,
			`async_insert`:                          1,
			//`async_insert_busy_timeout_ms`:          1000,   // 1 sec
			`async_insert_stale_timeout_ms`: 1000,
			`async_insert_max_query_number`: 40_000, // 40k block
			//`max_threads`:                   128,

			//`wait_for_async_insert`:1,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		//Debug: true,
	}
	if c.UseSsl {
		conf.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	connectFunc := func() *sql.DB {
		conn := clickhouse.OpenDB(conf)
		conn.SetMaxIdleConns(5)
		conn.SetMaxOpenConns(200)
		conn.SetConnMaxLifetime(time.Hour)
		return conn
	}
	conn := connectFunc()
	err = conn.Ping()
	if isError(err) {
		return nil, err
	}
	a = &Ch.Adapter{
		DB:        conn,
		Reconnect: connectFunc,
	}
	return a, nil

}
func main() {

	const createTable = `
CREATE TABLE IF NOT EXISTS ver4(
    root LowCardinality(String) CODEC(LZ4HC),
    bucket LowCardinality(String) CODEC(LZ4HC),
    key String CODEC(LZ4HC),
    version_id String CODEC(LZ4HC),
	ver DateTime64 CODEC(LZ4HC),
	is_deleted UInt8 CODEC(LZ4HC)
) engine=ReplacingMergeTree(ver, is_deleted) 
ORDER BY (root, bucket, key, version_id)
`

	cc := ClickhouseConf{
		DB:   "default",
		Host: "localhost",
		Port: 9000,
	}
	ch, err := cc.Connect()
	L.PanicIf(err, `cc.Connect`)

	// create table
	_, err = ch.Exec(createTable)
	L.PanicIf(err, `table creation`)

	// truncate table
	const truncateTable = `TRUNCATE TABLE ver4`
	_, err = ch.Exec(truncateTable)
	L.PanicIf(err, `table truncation`)

	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)
	const maxDelete = 1_000_000 // when changing this, must also change insertTotal
	deleteQueue := make(chan [2]string, maxDelete)

	const root = `root1`
	const bucket = `bucket1`

	// do parallel insert
	const insertThread = 8         // assume 8 instance inserting altogether
	const insertTotal = 20_000_000 // when changing this, must also change maxDelete
	const randomInsertEvery = 2000
	const insertUseAsync = false
	deleteChance := float32(maxDelete) / insertTotal
	insertErr := uint64(0)
	insertThreadDone := uint32(0)
	insertDone := uint64(0)
	insertDur := uint64(0)
	// testing creation of random pattern
	insertPatterns := []func() string{ // frequency
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/blocks/` + HEX32(0) + `/` + INT3() + `.` + HEX32(1) + `.` + HEX32(2) + `.blk`
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/checkpoints/checkpoint.` + INT2()
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/objs/` + STR() + `.` + INT1()
		},
		func() string {
			return `Veeam/Archive/Backups/KeySetParts/key-` + HEX32(0) + `-` + HEX32(1) + `.rep`
		},
		func() string {
			return `Veeam/Archive/Backups/KeySetParts/key-` + HEX32(0) + `.dat`
		},
	}

	var timedBuffer *chBuffer.TimedBuffer
	if !insertUseAsync {
		const insertEvery = 200_000
		timedBuffer = chBuffer.NewTimedBuffer(ch.DB, insertEvery, 1*time.Second, func(tx *sql.Tx) *sql.Stmt {
			const insertQuery = `
INSERT INTO ver4 VALUES(?, ?, ?, ?, ?, ?)
`
			stmt, err := tx.Prepare(insertQuery)
			L.IsError(err, `failed to tx.Prepare: `+insertQuery)
			return stmt
		})
	}
	for z := 0; z < insertThread; z++ {
		thread := z
		eg.Go(func() error {
			defer fmt.Println(`insert thread done`, thread)
			for z := 0; z < insertTotal/insertThread; z++ {
				atomic.AddUint64(&insertDur, track(func() {
					var key string
					if z%randomInsertEvery == 0 {
						key = insertPatterns[rand.Int()%len(insertPatterns)]()
					} else {
						key = insertPatterns[0]()
					}
					verId := S.RandomPassword(24)
					if insertUseAsync {
						// slow and high cpu usage, better use ch-timed-buffer, wait_for_async_insert=0 even worse
						const insertQuery = `
INSERT INTO ver4 SETTINGS async_insert=1, wait_for_async_insert=1 VALUES (?, ?, ?, ?, ?, 0)
`
						_, err := ch.Exec(insertQuery, root, bucket, key, verId, time.Now().Format(`2006-01-02 15:04:05.000000`))
						if isError(err) {
							atomic.AddUint64(&insertErr, 1)
							return
						}
					} else {
						// fast insert but slow to do other queries, also high cpu usage
						if !timedBuffer.Insert([]any{root, bucket, key, verId, time.Now().Format(`2006-01-02 15:04:05.000000`), uint8(0)}) {
							atomic.AddUint64(&insertErr, 1)
							return
						}
					}
					atomic.AddUint64(&insertDone, 1)
					if rand.Float32() < deleteChance {
						deleteQueue <- [2]string{key, verId}
					}
				}))
			}
			if atomic.AddUint32(&insertThreadDone, 1) == insertThread {
				close(deleteQueue)
			}
			return nil
		})
	}

	// do parallel deletion
	const deleteThread = 4
	const stopAnywaySec = 10
	exitAfterSec := stopAnywaySec
	deleteErr := uint64(0)
	deleteDone := uint64(0)
	deleteDur := uint64(0)
	for z := 0; z < deleteThread; z++ {
		thread := z
		eg.Go(func() error {

			defer fmt.Println(`delete thread done`, thread)
			for key := range deleteQueue {
				atomic.AddUint64(&deleteDur, track(func() {
					if insertUseAsync {
						const deleteQuery = `
DELETE FROM ver4 WHERE root=? AND bucket=? AND key=? AND version_id=?
`
						_, err := ch.Exec(deleteQuery, root, bucket, key[0], key[1])
						if isError(err) {
							atomic.AddUint64(&deleteErr, 1)
						}
					} else {
						if !timedBuffer.Insert([]any{root, bucket, key[0], key[1], time.Now().Format(`2006-01-02 15:04:05.000000`), uint8(1)}) {
							atomic.AddUint64(&insertErr, 1)
							return
						}
					}
					atomic.AddUint64(&deleteDone, 1)
				}))
				if exitAfterSec <= 0 {
					return nil
				}
			}
			return nil
		})
	}

	// do LIST query
	_ = []func() string{ // frequency
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/blocks/` + HEX32e(0) + `/` + INT3() + `/` + HEX32e(1)
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/blocks/` + HEX32e(0) + `/` + INT3()
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/blocks/` + HEX32e(0)
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/blocks`
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/checkpoints`
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1)
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0)
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1) + `/objs`
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0) + `/` + HEX32(1)
		},
		func() string {
			return `Veeam/Archive/Backups/` + HEX32(0)
		},
		func() string {
			return `Veeam/Archive/Backups/KeySetParts`
		},
		func() string {
			return `Veeam/Archive/Backups`
		},
		func() string {
			return `Veeam/Archive`
		},
		func() string {
			return `Veeam`
		},

		func() string {
			return ``
		},
	}

	const listingTotal = 100_000
	const listingThread = 4
	const maxRandomPattern = 15
	listingErr := uint64(0)
	listingDone := uint64(0)
	listingRows := uint64(0)
	listingDur := uint64(0)
	listingHighestRows := uint64(0)
	for z := 0; z < listingThread; z++ {
		thread := z
		eg.Go(func() error {
			defer fmt.Println(`LIST thread done`, thread)
			time.Sleep(time.Second)
			patterns := [maxRandomPattern]string{`Veeam`}
			for z := 0; z < listingTotal/listingThread; z++ {
				atomic.AddUint64(&listingDur, track(func() {
					pattern := patterns[rand.Int()%maxRandomPattern]
					parts := len(strings.Split(pattern, `/`)) + 1
					const listingQuery = `
SELECT arrayStringConcat(splitByChar('/', key, ?),'/'), MAX(version_id)
FROM ver4 FINAL
WHERE root=? 
  AND bucket=? 
  AND key LIKE ?
GROUP BY 1
ORDER BY 1 ASC
LIMIT 1001
`
					rows, err := ch.Query(listingQuery,
						parts, root, bucket, pattern+`%`)
					if isError(err) {
						atomic.AddUint64(&listingErr, 1)
						return
					}
					atomic.AddUint64(&listingDone, 1)
					var str1, str2 string
					var total = uint64(0)
					for rows.Next() {
						err := rows.Scan(&str1, &str2)
						if isError(err) {
							atomic.AddUint64(&listingErr, 1)
							return
						}
						total++
						if total < maxRandomPattern && !S.Contains(str1, `.`) { // most likely a directory
							patterns[(z+int(total))%maxRandomPattern] = str1
						}
					}
					atomic.AddUint64(&listingRows, total)
					if listingHighestRows < total {
						listingHighestRows = total
					}
					if exitAfterSec <= 0 {
						return
					}
				}))
			}
			return nil
		})
	}

	// show statistics
	const microSec2sec = 1000_000
	eg.Go(func() error {
		defer fmt.Println(`stats thread done`)
		ticker := time.NewTicker(time.Second)
		sec := 1
		for {
			select {
			case <-ticker.C:
				fmt.Printf("INS %d (%.1f/s), DEL %d (%.1f/s), LIST %d (%.1f/s), ROWS %d (high=%d), ERR: %d/%d/%d, %d sec\n",
					insertDone,
					float64(insertDone)/(float64(insertDur)/microSec2sec),
					deleteDone,
					float64(deleteDone)/(float64(deleteDur)/microSec2sec),
					listingDone,
					float64(listingDone)/(float64(listingDur)/microSec2sec),
					listingRows,
					listingHighestRows,
					insertErr,
					deleteErr,
					listingErr,
					sec,
				)
				if insertThreadDone >= insertThread && len(deleteQueue) == 0 {
					exitAfterSec--
					if exitAfterSec < 0 {
						if !insertUseAsync {
							timedBuffer.Close()
						}
						return nil
					}
				}
				sec++
			}
		}
	})

	L.IsError(eg.Wait(), `eg.Wait`)
}

func isError(err error) bool {
	if debug {
		if err != nil {
			panic(err)
		}
	}
	return err != nil
}

func STR() string {
	return S.RandomPassword(40)
}

func INT1() string {
	return fmt.Sprint(rand.Int() % 256)
}
func INT2() string {
	return fmt.Sprint(rand.Int() % (256 * 256))
}

func INT3() string {
	r := rand.Float32()
	if r < 0.33 {
		return `100000000`
	}
	if r < 0.66 {
		return `200000000`
	}
	return `300000000`
}

var existingKeys [3][]string

func HEX32(pos int) string {
	slice := existingKeys[pos]
	// 33% of returning existing key
	if pos < 2 && len(slice) > 100 && rand.Float32() < 0.66 {
		key := slice[rand.Int()%len(slice)]
		return key
	}
	key := S.RandomPassword(32)
	if pos < 2 {
		existingKeys[pos] = append(existingKeys[pos], key) // ignore race condition
	}
	return key
}

func HEX32e(pos int) string {
	slice := existingKeys[pos]
	if pos < 2 && len(slice) > 0 {
		key := slice[rand.Int()%len(slice)]
		return key
	}
	return S.RandomPassword(32)
}

func track(f func()) uint64 {
	start := time.Now()
	f()
	return uint64(time.Since(start).Microseconds())
}
