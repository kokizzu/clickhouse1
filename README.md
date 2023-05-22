# Testing Clickhouse Async Insert

Testing clickhouse [async insert](https://clickhouse.com/docs/en/cloud/bestpractices/asynchronous-inserts) wait vs no wait vs multithread [ch-timed-buffer](//github.com/kokizzu/ch-timed-buffer) vs single thread ch-timed-buffer, also check how fast [lighweight delete](//clickhouse.com/docs/en/guides/developer/lightweght-delete) are (it's slow, can only get 20 rps).

## Result:

- async insert with wait (`insertThread = 8 wait_for_async_insert=1`):
```
INS 16 (2.7/s), DEL 0 (NaN/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR 0/0/0, 1 sec
INS 40 (2.6/s), DEL 0 (NaN/s), LIST 1596 (399.9/s), ROWS 1721 (high=12), ERR 0/0/0, 2 sec
INS 56 (2.5/s), DEL 0 (NaN/s), LIST 3446 (431.2/s), ROWS 3574 (high=12), ERR 0/0/0, 3 sec
INS 80 (2.5/s), DEL 3 (61.0/s), LIST 5034 (420.0/s), ROWS 5159 (high=12), ERR 0/0/0, 4 sec
INS 96 (2.5/s), DEL 4 (63.2/s), LIST 6824 (426.7/s), ROWS 6952 (high=12), ERR 0/0/0, 5 sec
INS 120 (2.5/s), DEL 4 (63.2/s), LIST 8452 (422.8/s), ROWS 8579 (high=12), ERR 0/0/0, 6 sec
INS 136 (2.5/s), DEL 5 (67.8/s), LIST 10311 (429.9/s), ROWS 10438 (high=12), ERR 0/0/0, 7 sec
INS 160 (2.5/s), DEL 8 (67.3/s), LIST 11962 (427.4/s), ROWS 12089 (high=12), ERR 0/0/0, 8 sec
INS 176 (2.5/s), DEL 10 (69.8/s), LIST 13776 (430.7/s), ROWS 13904 (high=12), ERR 0/0/0, 9 sec
INS 200 (2.5/s), DEL 10 (69.8/s), LIST 15385 (427.5/s), ROWS 15511 (high=12), ERR 0/0/0, 10 sec
```
*LIST qps is high because of small data size

- async insert without wait (`insertThread = 8 wait_for_async_insert=0`): 
```
INS 17717 (2218.3/s), DEL 38 (20.2/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR 0/0/0, 1 sec
INS 32110 (2009.9/s), DEL 52 (13.3/s), LIST 368 (93.1/s), ROWS 4384 (high=1001), ERR 0/0/0, 2 sec
INS 45276 (1889.1/s), DEL 66 (11.1/s), LIST 644 (80.6/s), ROWS 4664 (high=1001), ERR 0/0/0, 3 sec
INS 57669 (1804.5/s), DEL 82 (10.3/s), LIST 902 (75.5/s), ROWS 4918 (high=1001), ERR 0/0/0, 4 sec
INS 72383 (1811.9/s), DEL 119 (12.0/s), LIST 1164 (72.9/s), ROWS 5180 (high=1001), ERR 0/0/0, 5 sec
INS 85208 (1777.3/s), DEL 151 (12.7/s), LIST 1378 (68.9/s), ROWS 5398 (high=1001), ERR 0/0/0, 6 sec
INS 96482 (1725.3/s), DEL 173 (13.1/s), LIST 1576 (65.8/s), ROWS 5592 (high=1001), ERR 0/0/0, 7 sec
INS 110311 (1725.6/s), DEL 201 (12.8/s), LIST 1837 (65.7/s), ROWS 5854 (high=1001), ERR 0/0/0, 8 sec
INS 122754 (1706.9/s), DEL 239 (13.3/s), LIST 2023 (63.4/s), ROWS 6039 (high=1001), ERR 0/0/0, 9 sec
INS 133171 (1673.9/s), DEL 257 (13.0/s), LIST 2203 (61.6/s), ROWS 6219 (high=1001), ERR 0/0/0, 10 sec
```
*LIST qps is high because of small data size

- manual buffered insert 8 thread (`insertThread = 8 insertUseAsync = false`):
```
INS 436775 (56429.2/s), DEL 248 (127.7/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR 0/0/0, 1 sec
INS 622030 (39842.5/s), DEL 250 (85.5/s), LIST 23 (6.5/s), ROWS 19 (high=1), ERR 0/0/0, 2 sec
INS 802581 (34153.2/s), DEL 250 (85.5/s), LIST 39 (5.1/s), ROWS 36 (high=1), ERR 0/0/0, 3 sec
INS 964593 (30829.9/s), DEL 252 (35.5/s), LIST 53 (4.7/s), ROWS 2052 (high=1001), ERR 0/0/0, 4 sec
INS 1136264 (28952.6/s), DEL 252 (35.5/s), LIST 87 (5.6/s), ROWS 3095 (high=1001), ERR 0/0/0, 5 sec
INS 1336872 (28334.6/s), DEL 252 (35.5/s), LIST 121 (6.1/s), ROWS 4132 (high=1001), ERR 0/0/0, 6 sec
INS 1514232 (27493.8/s), DEL 254 (19.3/s), LIST 161 (6.7/s), ROWS 4176 (high=1001), ERR 0/0/0, 7 sec
INS 1683076 (26873.8/s), DEL 254 (19.3/s), LIST 200 (7.2/s), ROWS 4214 (high=1001), ERR 0/0/0, 8 sec
INS 1838064 (25979.4/s), DEL 256 (14.6/s), LIST 233 (7.3/s), ROWS 4247 (high=1001), ERR 0/0/0, 9 sec
INS 2021747 (25667.3/s), DEL 256 (14.6/s), LIST 267 (7.4/s), ROWS 4281 (high=1001), ERR 0/0/0, 10 sec
```
- manual buffered insert (`insertThread = 1 insertUseAsync = false`):
```
INS 200000 (233092.4/s), DEL 78 (44.7/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 262577 (140613.1/s), DEL 82 (28.9/s), LIST 64 (16.7/s), ROWS 2083 (high=1001), ERR: 0/0/0, 2 sec
INS 309355 (128538.7/s), DEL 82 (28.9/s), LIST 268 (33.6/s), ROWS 4312 (high=1001), ERR: 0/0/0, 3 sec
INS 389355 (109336.6/s), DEL 104 (13.3/s), LIST 503 (42.1/s), ROWS 4545 (high=1001), ERR: 0/0/0, 4 sec
INS 429355 (100869.0/s), DEL 121 (12.2/s), LIST 675 (42.3/s), ROWS 4717 (high=1001), ERR: 0/0/0, 5 sec
INS 485990 (84686.8/s), DEL 125 (12.1/s), LIST 829 (41.6/s), ROWS 4871 (high=1001), ERR: 0/0/0, 6 sec
INS 526021 (83792.4/s), DEL 125 (12.1/s), LIST 963 (40.2/s), ROWS 5006 (high=1001), ERR: 0/0/0, 7 sec
INS 566021 (77860.3/s), DEL 125 (12.1/s), LIST 1091 (39.0/s), ROWS 5135 (high=1001), ERR: 0/0/0, 8 sec
INS 606024 (74776.2/s), DEL 139 (7.7/s), LIST 1266 (39.7/s), ROWS 5308 (high=1001), ERR: 0/0/0, 9 sec
INS 658128 (68341.9/s), DEL 141 (7.7/s), LIST 1416 (39.4/s), ROWS 5459 (high=1001), ERR: 0/0/0, 10 sec
```

- with driver.Conn `wait = true`
```
INS 16 (3.2/s), DEL 1 (79.7/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 40 (2.7/s), DEL 1 (79.7/s), LIST 1768 (442.7/s), ROWS 1900 (high=20), ERR: 0/0/0, 2 sec
INS 56 (2.7/s), DEL 1 (79.7/s), LIST 3515 (439.8/s), ROWS 3646 (high=20), ERR: 0/0/0, 3 sec
INS 80 (2.6/s), DEL 4 (113.1/s), LIST 5234 (436.6/s), ROWS 5365 (high=20), ERR: 0/0/0, 4 sec
INS 96 (2.6/s), DEL 5 (119.4/s), LIST 6648 (415.8/s), ROWS 6780 (high=20), ERR: 0/0/0, 5 sec
INS 120 (2.6/s), DEL 6 (123.2/s), LIST 8350 (417.7/s), ROWS 8484 (high=20), ERR: 0/0/0, 6 sec
INS 136 (2.6/s), DEL 6 (123.2/s), LIST 10052 (419.0/s), ROWS 10183 (high=20), ERR: 0/0/0, 7 sec
INS 160 (2.6/s), DEL 8 (120.3/s), LIST 11782 (421.0/s), ROWS 11914 (high=20), ERR: 0/0/0, 8 sec
INS 176 (2.5/s), DEL 10 (21.5/s), LIST 12837 (401.3/s), ROWS 12970 (high=20), ERR: 0/0/0, 9 sec
INS 200 (2.5/s), DEL 12 (25.1/s), LIST 14560 (404.6/s), ROWS 14692 (high=20), ERR: 0/0/0, 10 sec
```

- with driver.Conn `wait = false`
```
INS 27420 (3434.5/s), DEL 143 (37.5/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 43631 (2731.5/s), DEL 176 (22.4/s), LIST 336 (84.9/s), ROWS 8358 (high=1001), ERR: 0/0/0, 2 sec
INS 56265 (2348.1/s), DEL 192 (19.5/s), LIST 608 (76.2/s), ROWS 8630 (high=1001), ERR: 0/0/0, 3 sec
INS 73072 (2287.0/s), DEL 276 (17.3/s), LIST 942 (78.8/s), ROWS 8964 (high=1001), ERR: 0/0/0, 4 sec
INS 88179 (2207.0/s), DEL 338 (17.0/s), LIST 1168 (73.2/s), ROWS 9190 (high=1001), ERR: 0/0/0, 5 sec
INS 100871 (2104.5/s), DEL 395 (17.1/s), LIST 1391 (69.6/s), ROWS 9416 (high=1001), ERR: 0/0/0, 6 sec
INS 115596 (2067.2/s), DEL 449 (16.1/s), LIST 1667 (69.6/s), ROWS 9689 (high=1001), ERR: 0/0/0, 7 sec
INS 129851 (2031.8/s), DEL 513 (16.2/s), LIST 1899 (68.0/s), ROWS 9921 (high=1001), ERR: 0/0/0, 8 sec
INS 141553 (1968.7/s), DEL 569 (15.9/s), LIST 2094 (65.5/s), ROWS 10116 (high=1001), ERR: 0/0/0, 9 sec
INS 153710 (1923.9/s), DEL 601 (15.1/s), LIST 2278 (63.3/s), ROWS 10300 (high=1001), ERR: 0/0/0, 10 sec
```

## Configs

- `insertThread` - number of goroutine to insert
- `deleteThread` - number of goroutine to delete
- `insertUseAsync` - use async insert syntax or ch-timed-buffer
- `maxDelete` - maximum estimated number of record to delete after insert
- `insertTotal` - total number of record to insert
- `randomInsertEvery` - insert random pattern every n insert
- `listingTotal` - number of listing query to perform, will stop 10 sec
- `stopAnywaySec` - stop anyway after n second all insert done
- `debug` - throw panic on error
- `clickhouse.WithStdAsync` if `useDriverAsync` or `wait_for_async_insert`