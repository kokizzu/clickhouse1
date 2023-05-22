# Testing Clickhouse async insert

Testing clickhouse async insert:

## Result:

- async insert with wait (`wait_for_async_insert=1`):
```
inserted 16 (2.7/s), deleted 0 (NaN/s), listing 0 (NaN/s), listing rows 0 (high=0), ERR 0/0/0, 1 sec
inserted 40 (2.6/s), deleted 0 (NaN/s), listing 1596 (399.9/s), listing rows 1721 (high=12), ERR 0/0/0, 2 sec
inserted 56 (2.5/s), deleted 0 (NaN/s), listing 3446 (431.2/s), listing rows 3574 (high=12), ERR 0/0/0, 3 sec
inserted 80 (2.5/s), deleted 3 (61.0/s), listing 5034 (420.0/s), listing rows 5159 (high=12), ERR 0/0/0, 4 sec
inserted 96 (2.5/s), deleted 4 (63.2/s), listing 6824 (426.7/s), listing rows 6952 (high=12), ERR 0/0/0, 5 sec
inserted 120 (2.5/s), deleted 4 (63.2/s), listing 8452 (422.8/s), listing rows 8579 (high=12), ERR 0/0/0, 6 sec
inserted 136 (2.5/s), deleted 5 (67.8/s), listing 10311 (429.9/s), listing rows 10438 (high=12), ERR 0/0/0, 7 sec
inserted 160 (2.5/s), deleted 8 (67.3/s), listing 11962 (427.4/s), listing rows 12089 (high=12), ERR 0/0/0, 8 sec
inserted 176 (2.5/s), deleted 10 (69.8/s), listing 13776 (430.7/s), listing rows 13904 (high=12), ERR 0/0/0, 9 sec
inserted 200 (2.5/s), deleted 10 (69.8/s), listing 15385 (427.5/s), listing rows 15511 (high=12), ERR 0/0/0, 10 sec
```
*listing qps is high because of small data size

- async insert without wait (`wait_for_async_insert=0`): 
```
inserted 17717 (2218.3/s), deleted 38 (20.2/s), listing 0 (NaN/s), listing rows 0 (high=0), ERR 0/0/0, 1 sec
inserted 32110 (2009.9/s), deleted 52 (13.3/s), listing 368 (93.1/s), listing rows 4384 (high=1001), ERR 0/0/0, 2 sec
inserted 45276 (1889.1/s), deleted 66 (11.1/s), listing 644 (80.6/s), listing rows 4664 (high=1001), ERR 0/0/0, 3 sec
inserted 57669 (1804.5/s), deleted 82 (10.3/s), listing 902 (75.5/s), listing rows 4918 (high=1001), ERR 0/0/0, 4 sec
inserted 72383 (1811.9/s), deleted 119 (12.0/s), listing 1164 (72.9/s), listing rows 5180 (high=1001), ERR 0/0/0, 5 sec
inserted 85208 (1777.3/s), deleted 151 (12.7/s), listing 1378 (68.9/s), listing rows 5398 (high=1001), ERR 0/0/0, 6 sec
inserted 96482 (1725.3/s), deleted 173 (13.1/s), listing 1576 (65.8/s), listing rows 5592 (high=1001), ERR 0/0/0, 7 sec
inserted 110311 (1725.6/s), deleted 201 (12.8/s), listing 1837 (65.7/s), listing rows 5854 (high=1001), ERR 0/0/0, 8 sec
inserted 122754 (1706.9/s), deleted 239 (13.3/s), listing 2023 (63.4/s), listing rows 6039 (high=1001), ERR 0/0/0, 9 sec
inserted 133171 (1673.9/s), deleted 257 (13.0/s), listing 2203 (61.6/s), listing rows 6219 (high=1001), ERR 0/0/0, 10 sec
```
*listing qps is high because of small data size

- manual buffered insert 8 thread (`insertUseAsync = false`):
```
inserted 436775 (56429.2/s), deleted 248 (127.7/s), listing 0 (NaN/s), listing rows 0 (high=0), ERR 0/0/0, 1 sec
inserted 622030 (39842.5/s), deleted 250 (85.5/s), listing 23 (6.5/s), listing rows 19 (high=1), ERR 0/0/0, 2 sec
inserted 802581 (34153.2/s), deleted 250 (85.5/s), listing 39 (5.1/s), listing rows 36 (high=1), ERR 0/0/0, 3 sec
inserted 964593 (30829.9/s), deleted 252 (35.5/s), listing 53 (4.7/s), listing rows 2052 (high=1001), ERR 0/0/0, 4 sec
inserted 1136264 (28952.6/s), deleted 252 (35.5/s), listing 87 (5.6/s), listing rows 3095 (high=1001), ERR 0/0/0, 5 sec
inserted 1336872 (28334.6/s), deleted 252 (35.5/s), listing 121 (6.1/s), listing rows 4132 (high=1001), ERR 0/0/0, 6 sec
inserted 1514232 (27493.8/s), deleted 254 (19.3/s), listing 161 (6.7/s), listing rows 4176 (high=1001), ERR 0/0/0, 7 sec
inserted 1683076 (26873.8/s), deleted 254 (19.3/s), listing 200 (7.2/s), listing rows 4214 (high=1001), ERR 0/0/0, 8 sec
inserted 1838064 (25979.4/s), deleted 256 (14.6/s), listing 233 (7.3/s), listing rows 4247 (high=1001), ERR 0/0/0, 9 sec
inserted 2021747 (25667.3/s), deleted 256 (14.6/s), listing 267 (7.4/s), listing rows 4281 (high=1001), ERR 0/0/0, 10 sec
```
- manual buffered insert single thread:
```
inserted 200000 (233092.4/s), deleted 78 (44.7/s), listing 0 (NaN/s), listing rows 0 (high=0), ERR: 0/0/0, 1 sec
inserted 262577 (140613.1/s), deleted 82 (28.9/s), listing 64 (16.7/s), listing rows 2083 (high=1001), ERR: 0/0/0, 2 sec
inserted 309355 (128538.7/s), deleted 82 (28.9/s), listing 268 (33.6/s), listing rows 4312 (high=1001), ERR: 0/0/0, 3 sec
inserted 389355 (109336.6/s), deleted 104 (13.3/s), listing 503 (42.1/s), listing rows 4545 (high=1001), ERR: 0/0/0, 4 sec
inserted 429355 (100869.0/s), deleted 121 (12.2/s), listing 675 (42.3/s), listing rows 4717 (high=1001), ERR: 0/0/0, 5 sec
inserted 485990 (84686.8/s), deleted 125 (12.1/s), listing 829 (41.6/s), listing rows 4871 (high=1001), ERR: 0/0/0, 6 sec
inserted 526021 (83792.4/s), deleted 125 (12.1/s), listing 963 (40.2/s), listing rows 5006 (high=1001), ERR: 0/0/0, 7 sec
inserted 566021 (77860.3/s), deleted 125 (12.1/s), listing 1091 (39.0/s), listing rows 5135 (high=1001), ERR: 0/0/0, 8 sec
inserted 606024 (74776.2/s), deleted 139 (7.7/s), listing 1266 (39.7/s), listing rows 5308 (high=1001), ERR: 0/0/0, 9 sec
inserted 658128 (68341.9/s), deleted 141 (7.7/s), listing 1416 (39.4/s), listing rows 5459 (high=1001), ERR: 0/0/0, 10 sec
```