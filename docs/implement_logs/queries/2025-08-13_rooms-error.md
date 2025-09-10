## 部屋一覧にアクセスした際のログ
```bash
2025/08/13 00:25:26 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/user_repository.go:56 sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB
[4.331ms] [rows:1] SELECT * FROM "users" WHERE supabase_user_id = '7c92fde0-2785-454d-86d7-dd19f4290034' ORDER BY "users"."id" LIMIT 1

2025/08/13 00:25:26 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[0.838ms] [rows:2] SELECT * FROM "platforms" WHERE "platforms"."id" IN ('d784a1f7-c7e0-4633-9ceb-2f237b9a3323','d28c2776-439f-45ca-94de-d222bd50bab9')

2025/08/13 00:25:26 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[2.161ms] [rows:5] SELECT * FROM "game_versions" WHERE is_active = true ORDER BY display_order ASC
2025/08/13 00:25:26 [DEBUG] Rooms: 未認証ユーザー

2025/08/13 00:25:26 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:106
[1.263ms] [rows:5] SELECT * FROM "game_versions" WHERE "game_versions"."id" IN ('05ec7899-6afd-48a2-a8e3-2495348c9990','4568e2b0-c875-4d3a-977e-2ba995fc67c8','6ad13aca-8185-4c5f-b0ce-042519aadfe9','008e41f2-7c63-4c1b-b24d-f4e7ad865acb','98a9b5c5-7dda-4887-a4d9-cc0556272271')

2025/08/13 00:25:26 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:106 sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB; sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB; sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB; sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB
[0.633ms] [rows:4] SELECT * FROM "users" WHERE "users"."id" IN ('52612959-3044-4c7c-a623-2fdf52088388','d56f582b-4490-4399-b698-29e57d3b208c','cdb262ff-2889-4b3b-9446-1017e206eed9','9c40956c-88b2-4de6-a441-c8028b8566d7')

2025/08/13 00:25:26 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:106 sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB; sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB; sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB; sql: Scan error on column index 16, name "favorite_games": json: cannot unmarshal array into Go value of type models.JSONB
[6.712ms] [rows:18] SELECT rooms.*, COUNT(room_members.id) as current_players FROM "rooms" LEFT JOIN room_members ON rooms.id = room_members.room_id AND room_members.status = 'active' WHERE rooms.is_active = true GROUP BY "rooms"."id" ORDER BY rooms.created_at DESC LIMIT 100
2025/08/13 00:25:26 "GET http://localhost:8080/rooms HTTP/1.1" from [::1]:54364 - 500 46B in 13.606848ms
2025/08/13 00:25:26 "GET http://localhost:8080/.well-known/appspecific/com.chrome.devtools.json HTTP/1.1" from [::1]:54364 - 404 19B in 30.921µs
```
