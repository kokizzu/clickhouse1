package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
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

func (c ClickhouseConf) Connect() (a driver.Conn, err error) {
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
			//`wait_for_async_insert`:1,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		MaxOpenConns: 200,
		MaxIdleConns: 5,
		//Debug: true,
	}
	if c.UseSsl {
		conf.TLS = &tls.Config{
			InsecureSkipVerify: true,
		}
	}
	connectFunc := func() driver.Conn {
		conn, err := clickhouse.Open(conf)
		L.IsError(err, `clickhouse.Open`)
		return conn
	}
	conn := connectFunc()
	err = conn.Ping(context.Background())
	if isError(err) {
		return nil, err
	}
	return conn, nil

}
func main() {
	ctx := context.Background()

	const createTable = `
CREATE TABLE IF NOT EXISTS ver3(
    root LowCardinality(String) CODEC(LZ4HC),
    bucket LowCardinality(String) CODEC(LZ4HC),
    key String CODEC(LZ4HC),
    version_id String CODEC(LZ4HC),
	ver DateTime64 CODEC(LZ4HC)
) engine=ReplacingMergeTree(ver) 
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
	err = ch.Exec(ctx, createTable)
	L.PanicIf(err, `table creation`)

	// truncate table
	const truncateTable = `TRUNCATE TABLE ver3`
	err = ch.Exec(ctx, truncateTable)
	L.PanicIf(err, `table truncation`)

	eg, ctx := errgroup.WithContext(ctx)
	const maxDelete = 1_000_000 // when changing this, must also change insertTotal
	deleteQueue := make(chan [2]string, maxDelete)

	const root = `root1`
	const bucket = `bucket1`

	// do parallel insert
	const insertThread = 8         // assume 8 instance inserting altogether
	const insertTotal = 20_000_000 // when changing this, must also change maxDelete
	const randomInsertEvery = 2000
	const useDriverAsync = true
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

	for z := 0; z < insertThread; z++ {
		thread := z
		eg.Go(func() error {
			defer fmt.Println(`insert thread done`, thread)
			ctx := clickhouse.Context(context.Background(), clickhouse.WithStdAsync(false))
			for z := 0; z < insertTotal/insertThread; z++ {
				atomic.AddUint64(&insertDur, track(func() {
					var key string
					if z%randomInsertEvery == 0 {
						key = insertPatterns[rand.Int()%len(insertPatterns)]()
					} else {
						key = insertPatterns[0]()
					}
					verId := S.RandomPassword(24)
					var err error
					if useDriverAsync {
						insertQuery := fmt.Sprintf(`
INSERT INTO ver3 VALUES ('%s', '%s', '%s', '%s', '%s')
`, root, bucket, key, verId, time.Now().Format(`2006-01-02 15:04:05.000000`)) // assume no sql injection
						err = ch.AsyncInsert(ctx, insertQuery, false) // wait = false
					} else {
						// slow and high cpu usage, better use ch-timed-buffer, wait_for_async_insert=0 even worse
						const insertQuery = `
INSERT INTO ver3 SETTINGS async_insert=1, wait_for_async_insert=1 VALUES (?, ?, ?, ?, ?)
`
						err = ch.Exec(ctx, insertQuery, root, bucket, key, verId, time.Now().Format(`2006-01-02 15:04:05.000000`))
					}
					if isError(err) {
						atomic.AddUint64(&insertErr, 1)
						return
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
					const deleteQuery = `
DELETE FROM ver3 WHERE root=? AND bucket=? AND key=? AND version_id=?
`
					err := ch.Exec(ctx, deleteQuery, root, bucket, key[0], key[1])
					if isError(err) {
						atomic.AddUint64(&deleteErr, 1)
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
FROM ver3
WHERE root=? 
  AND bucket=? 
  AND key LIKE ?
GROUP BY 1
ORDER BY 1 ASC
LIMIT 1001
`
					rows, err := ch.Query(ctx, listingQuery,
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
