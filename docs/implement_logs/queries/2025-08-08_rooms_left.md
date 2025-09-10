## 部屋一覧にて
参加済みの部屋から退出したあとに表示されるクエリ。

record not foundが表示されている。リロードしても全く同じクエリが表示されている状態。

```bash
2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[1.299ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59034 - 200 22B in 1.454354ms

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[0.636ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59050 - 200 22B in 886.623µs

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[0.827ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59050 - 200 22B in 999.093µs

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[0.670ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59034 - 200 22B in 856.532µs

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[1.266ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59050 - 200 22B in 1.509694ms

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[0.842ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59066 - 200 22B in 1.113353ms
2025/08/08 21:14:03 "POST http://localhost:8080/api/auth/sync HTTP/1.1" from [::1]:59028 - 200 176B in 42.00538ms

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[0.509ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59028 - 200 22B in 702.914µs

2025/08/08 21:14:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:385 record not found
[0.545ms] [rows:0] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' ORDER BY "room_members"."id" LIMIT 1
2025/08/08 21:14:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:59066 - 200 22B in 704.028µs
```
