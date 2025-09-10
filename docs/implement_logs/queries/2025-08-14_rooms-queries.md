## 部屋一覧のクエリとログ
```bash
2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/user_repository.go:60
[5.336ms] [rows:1] SELECT "id","supabase_user_id","email","username","display_name","avatar_url","bio","psn_online_id","nintendo_network_id","nintendo_switch_id","pretendo_network_id","twitter_id","is_active","role","created_at","updated_at" FROM "users" WHERE supabase_user_id = '7c92fde0-2785-454d-86d7-dd19f4290034' ORDER BY "users"."id" LIMIT 1

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[1.797ms] [rows:2] SELECT * FROM "platforms" WHERE "platforms"."id" IN ('d784a1f7-c7e0-4633-9ceb-2f237b9a3323','d28c2776-439f-45ca-94de-d222bd50bab9')

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[4.078ms] [rows:5] SELECT * FROM "game_versions" WHERE is_active = true ORDER BY display_order ASC
2025/08/14 13:56:03 [DEBUG] Rooms: 認証済みユーザー: ID=52612959-3044-4c7c-a623-2fdf52088388, Email=rdwbocungelt5@gmail.com

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:195
[6.969ms] [rows:17]
                SELECT
                        rooms.*,
                        gv.name as game_version_name,
                        gv.code as game_version_code,
                        u.display_name as host_display_name,
                        u.psn_online_id as host_psn_online_id,
                        COUNT(DISTINCT rm_all.id) as current_players,
                        CASE WHEN rm_user.id IS NOT NULL THEN true ELSE false END as is_joined
                FROM rooms
                LEFT JOIN game_versions gv ON rooms.game_version_id = gv.id
                LEFT JOIN users u ON rooms.host_user_id = u.id
                LEFT JOIN room_members rm_all ON rooms.id = rm_all.room_id AND rm_all.status = 'active'
                LEFT JOIN room_members rm_user ON rooms.id = rm_user.room_id AND rm_user.user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND rm_user.status = 'active'
                WHERE rooms.is_active = true

                GROUP BY rooms.id, gv.id, u.id, rm_user.id
                ORDER BY
                        CASE WHEN rm_user.id IS NOT NULL THEN 0 ELSE 1 END,
                        rooms.created_at DESC
                LIMIT 100 OFFSET 0

2025/08/14 13:56:03 "GET http://localhost:8080/rooms HTTP/1.1" from [::1]:33060 - 200 116826B in 31.316863ms

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[0.304ms] [rows:2] SELECT * FROM "platforms" WHERE "platforms"."id" IN ('d784a1f7-c7e0-4633-9ceb-2f237b9a3323','d28c2776-439f-45ca-94de-d222bd50bab9')

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[1.012ms] [rows:5] SELECT * FROM "game_versions" WHERE is_active = true ORDER BY display_order ASC
2025/08/14 13:56:03 "GET http://localhost:8080/api/game-versions/active HTTP/1.1" from [::1]:33060 - 200 2443B in 1.105531ms

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[0.500ms] [rows:2] SELECT * FROM "platforms" WHERE "platforms"."id" IN ('d784a1f7-c7e0-4633-9ceb-2f237b9a3323','d28c2776-439f-45ca-94de-d222bd50bab9')

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/game_version_repository.go:57
[1.006ms] [rows:5] SELECT * FROM "game_versions" WHERE is_active = true ORDER BY display_order ASC
2025/08/14 13:56:03 "GET http://localhost:8080/api/game-versions/active HTTP/1.1" from [::1]:33060 - 200 2443B in 1.06745ms
2025/08/14 13:56:03 "GET http://localhost:8080/api/config/supabase HTTP/1.1" from [::1]:33060 - 200 297B in 40.077µs
2025/08/14 13:56:03 "POST http://localhost:8080/api/auth/sync HTTP/1.1" from [::1]:33060 - 200 176B in 131.341µs
2025/08/14 13:56:03 "POST http://localhost:8080/api/auth/sync HTTP/1.1" from [::1]:33060 - 200 176B in 236.32µs
2025/08/14 13:56:03 "POST http://localhost:8080/api/auth/sync HTTP/1.1" from [::1]:33068 - 200 176B in 108.404µs
2025/08/14 13:56:03 "POST http://localhost:8080/api/auth/sync HTTP/1.1" from [::1]:33090 - 200 176B in 110.821µs

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:408
[6.227ms] [rows:1] SELECT * FROM "room_members" WHERE user_id = '52612959-3044-4c7c-a623-2fdf52088388' AND status = 'active' LIMIT 1

2025/08/14 13:56:03 /home/motty93/github.com/motty93/mhp-rooms/internal/repository/room_repository.go:423
[2.485ms] [rows:1] SELECT * FROM "rooms" WHERE id = 'be08d649-a001-4ba3-b2da-f05b1fafd2ac' ORDER BY "rooms"."id" LIMIT 1
2025/08/14 13:56:03 "GET http://localhost:8080/api/user/current-room HTTP/1.1" from [::1]:33076 - 200 1395B in 9.9887ms
```
