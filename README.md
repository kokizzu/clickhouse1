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

- manual buffered insert and using insert and FINAL
```
INS 190350 (26272.7/s), DEL 9650 (17331.6/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 248993 (15707.8/s), DEL 12612 (19995.1/s), LIST 25 (6.9/s), ROWS 1023 (high=1001), ERR: 0/0/0, 2 sec
INS 319932 (13438.5/s), DEL 16139 (10769.3/s), LIST 64 (8.3/s), ROWS 3077 (high=1001), ERR: 0/0/0, 3 sec
INS 381250 (12001.2/s), DEL 19243 (12229.0/s), LIST 135 (11.4/s), ROWS 4157 (high=1001), ERR: 0/0/0, 4 sec
INS 457516 (12111.4/s), DEL 22977 (14331.6/s), LIST 239 (15.0/s), ROWS 5277 (high=1001), ERR: 0/0/0, 5 sec
INS 571961 (12076.5/s), DEL 28533 (17193.4/s), LIST 337 (16.9/s), ROWS 5375 (high=1001), ERR: 0/0/0, 6 sec
INS 648152 (12037.2/s), DEL 32342 (18833.9/s), LIST 448 (18.7/s), ROWS 5486 (high=1001), ERR: 0/0/0, 7 sec
INS 749165 (11789.8/s), DEL 37413 (18412.8/s), LIST 546 (19.6/s), ROWS 5584 (high=1001), ERR: 0/0/0, 8 sec
INS 825432 (11905.7/s), DEL 41148 (18178.4/s), LIST 672 (21.0/s), ROWS 5711 (high=1001), ERR: 0/0/0, 9 sec
INS 938205 (11811.2/s), DEL 46630 (17887.1/s), LIST 769 (21.4/s), ROWS 5808 (high=1001), ERR: 0/0/0, 10 sec
```

- same as above but 200k buffer
```
INS 381003 (60494.5/s), DEL 18997 (3389295.3/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 403794 (25620.5/s), DEL 20114 (15372.8/s), LIST 1593 (473.0/s), ROWS 4 (high=1), ERR: 0/0/0, 2 sec
INS 431197 (18239.0/s), DEL 21468 (6024.7/s), LIST 1613 (210.5/s), ROWS 24 (high=1), ERR: 0/0/0, 3 sec
INS 481545 (15184.0/s), DEL 23898 (3779.8/s), LIST 1636 (138.1/s), ROWS 4049 (high=1001), ERR: 0/0/0, 4 sec
INS 630076 (15924.0/s), DEL 31368 (4923.5/s), LIST 1778 (111.4/s), ROWS 9198 (high=1001), ERR: 0/0/0, 5 sec
INS 630076 (15924.0/s), DEL 31368 (4923.5/s), LIST 1990 (99.7/s), ROWS 9410 (high=1001), ERR: 0/0/0, 6 sec
INS 820431 (15323.8/s), DEL 41016 (6404.4/s), LIST 2143 (89.5/s), ROWS 9563 (high=1001), ERR: 0/0/0, 7 sec
INS 820431 (15323.8/s), DEL 41016 (6404.4/s), LIST 2305 (82.5/s), ROWS 9725 (high=1001), ERR: 0/0/0, 8 sec
INS 1010979 (14414.1/s), DEL 50470 (6328.6/s), LIST 2431 (76.1/s), ROWS 9851 (high=1001), ERR: 0/0/0, 9 sec
INS 1010979 (14414.1/s), DEL 50470 (6328.6/s), LIST 2556 (71.1/s), ROWS 9977 (high=1001), ERR: 0/0/0, 10 sec
```

- same as above but using chproxy (http not native protocol) for select queries
```
INS 190554 (25950.2/s), DEL 9446 (47594.8/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 228660 (19223.0/s), DEL 11340 (11868.1/s), LIST 19 (5.2/s), ROWS 1019 (high=1001), ERR: 0/0/0, 2 sec
INS 304906 (13287.4/s), DEL 15096 (14448.0/s), LIST 49 (6.5/s), ROWS 3057 (high=1001), ERR: 0/0/0, 3 sec
INS 353932 (11140.0/s), DEL 17509 (16312.6/s), LIST 103 (8.9/s), ROWS 4115 (high=1001), ERR: 0/0/0, 4 sec
INS 430147 (11007.2/s), DEL 21331 (18163.5/s), LIST 200 (12.7/s), ROWS 7225 (high=1001), ERR: 0/0/0, 5 sec
INS 506316 (10867.4/s), DEL 25164 (16182.5/s), LIST 316 (15.9/s), ROWS 10350 (high=1001), ERR: 0/0/0, 6 sec
INS 605750 (10893.6/s), DEL 30182 (19237.8/s), LIST 577 (24.1/s), ROWS 10611 (high=1001), ERR: 0/0/0, 7 sec
INS 682669 (11078.9/s), DEL 34034 (17910.6/s), LIST 727 (26.0/s), ROWS 10761 (high=1001), ERR: 0/0/0, 8 sec
INS 788532 (11029.9/s), DEL 39306 (17561.7/s), LIST 838 (26.3/s), ROWS 10872 (high=1001), ERR: 0/0/0, 9 sec
INS 867115 (11149.0/s), DEL 43354 (17227.0/s), LIST 920 (25.7/s), ROWS 10954 (high=1001), ERR: 0/0/0, 10 sec
```

- same as above but 2s cache duration
```
INS 190323 (26336.1/s), DEL 9677 (29100.3/s), LIST 0 (NaN/s), ROWS 0 (high=0), ERR: 0/0/0, 1 sec
INS 236398 (14905.4/s), DEL 11964 (31028.5/s), LIST 22 (6.0/s), ROWS 1022 (high=1001), ERR: 0/0/0, 2 sec
INS 312597 (13882.3/s), DEL 15765 (35568.8/s), LIST 89 (11.5/s), ROWS 4105 (high=1001), ERR: 0/0/0, 3 sec
INS 388713 (13141.7/s), DEL 19650 (24275.2/s), LIST 269 (22.5/s), ROWS 7305 (high=1001), ERR: 0/0/0, 4 sec
INS 503034 (12819.9/s), DEL 25330 (18432.0/s), LIST 466 (29.2/s), ROWS 7504 (high=1001), ERR: 0/0/0, 5 sec
INS 579206 (12728.1/s), DEL 29158 (20930.4/s), LIST 614 (30.8/s), ROWS 7652 (high=1001), ERR: 0/0/0, 6 sec
INS 693397 (12561.7/s), DEL 34968 (24325.7/s), LIST 724 (30.3/s), ROWS 7762 (high=1001), ERR: 0/0/0, 7 sec
INS 769597 (12476.3/s), DEL 38768 (18859.4/s), LIST 833 (29.9/s), ROWS 7871 (high=1001), ERR: 0/0/0, 8 sec
INS 883866 (12436.4/s), DEL 44502 (21358.6/s), LIST 916 (28.7/s), ROWS 7954 (high=1001), ERR: 0/0/0, 9 sec
INS 960038 (12403.7/s), DEL 48331 (20108.0/s), LIST 1009 (28.1/s), ROWS 8047 (high=1001), ERR: 0/0/0, 10 sec
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