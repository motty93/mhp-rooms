-- room_membersテーブルのダミーデータ
-- 実行前に既存のデータをクリア（必要に応じてコメントアウト）
-- TRUNCATE TABLE room_members;

-- is_closedフラグを更新（一部のルームを閉じた状態にする）
UPDATE rooms SET is_closed = true WHERE name = '進行中の部屋';
UPDATE rooms SET is_closed = true WHERE name = 'レア素材狙い' AND current_players >= 4;

-- 各ルームにメンバーを追加
-- 注意: host_user_idは必ずis_host=trueでplayer_number=1として追加

-- 初心者歓迎部屋（MHP2G） - host: 素材コレクター
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), 'ee62c8b2-dec5-4179-a823-a99e802697e1', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '30 minutes'),
  (gen_random_uuid(), 'ee62c8b2-dec5-4179-a823-a99e802697e1', '0d791003-2a0d-4a0a-920e-86509399c611', 2, false, 'active', CURRENT_TIMESTAMP - INTERVAL '25 minutes'),
  (gen_random_uuid(), 'ee62c8b2-dec5-4179-a823-a99e802697e1', '509e4f15-eb60-4811-8486-3d23f0928419', 3, false, 'active', CURRENT_TIMESTAMP - INTERVAL '20 minutes');

-- 初心者歓迎部屋（MHP2G） - host: 猫好きハンター
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), '1107b5af-bf2e-4c1f-a234-eebab3c3fef6', '509e4f15-eb60-4811-8486-3d23f0928419', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '45 minutes');

-- 初心者歓迎部屋（MHP2G） - host: ハンター太郎
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), 'f0c7c11b-26d3-4532-9828-c06ea3c2725f', '0d791003-2a0d-4a0a-920e-86509399c611', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '1 hour'),
  (gen_random_uuid(), 'f0c7c11b-26d3-4532-9828-c06ea3c2725f', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 2, false, 'active', CURRENT_TIMESTAMP - INTERVAL '55 minutes');

-- 上位ティガレックス討伐（MHP2） - host: 素材コレクター
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at, left_at) VALUES
  (gen_random_uuid(), 'c3aa000d-b0f6-4228-a4e9-e602c285c61b', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '2 hours', NULL),
  (gen_random_uuid(), 'c3aa000d-b0f6-4228-a4e9-e602c285c61b', '509e4f15-eb60-4811-8486-3d23f0928419', 2, false, 'inactive', CURRENT_TIMESTAMP - INTERVAL '1 hour 50 minutes', CURRENT_TIMESTAMP - INTERVAL '1 hour 30 minutes');

-- 上位ティガレックス討伐（MHP2） - host: 猫好きハンター
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), '50cb4e01-6f8e-404f-a33e-690cc9d7eaca', '509e4f15-eb60-4811-8486-3d23f0928419', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '40 minutes'),
  (gen_random_uuid(), '50cb4e01-6f8e-404f-a33e-690cc9d7eaca', '0d791003-2a0d-4a0a-920e-86509399c611', 2, false, 'active', CURRENT_TIMESTAMP - INTERVAL '35 minutes'),
  (gen_random_uuid(), '50cb4e01-6f8e-404f-a33e-690cc9d7eaca', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 3, false, 'active', CURRENT_TIMESTAMP - INTERVAL '30 minutes');

-- 上位ティガレックス討伐（MHP2） - host: ハンター太郎
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), '44f03ceb-aa6f-4a6c-be9a-a03946ff9093', '0d791003-2a0d-4a0a-920e-86509399c611', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '15 minutes');

-- レア素材狙い（MHP） - host: 素材コレクター
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), '40d3da65-6064-4389-b046-46985c929b32', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '1 hour 30 minutes'),
  (gen_random_uuid(), '40d3da65-6064-4389-b046-46985c929b32', '509e4f15-eb60-4811-8486-3d23f0928419', 2, false, 'active', CURRENT_TIMESTAMP - INTERVAL '1 hour 25 minutes'),
  (gen_random_uuid(), '40d3da65-6064-4389-b046-46985c929b32', '0d791003-2a0d-4a0a-920e-86509399c611', 3, false, 'active', CURRENT_TIMESTAMP - INTERVAL '1 hour 20 minutes');

-- レア素材狙い（MHP） - host: 猫好きハンター
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), '44ea26d1-1fb1-4178-acb9-de3e34fa5703', '509e4f15-eb60-4811-8486-3d23f0928419', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '50 minutes'),
  (gen_random_uuid(), '44ea26d1-1fb1-4178-acb9-de3e34fa5703', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 2, false, 'active', CURRENT_TIMESTAMP - INTERVAL '45 minutes');

-- レア素材狙い（MHP） - host: ハンター太郎
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), 'ddee3085-2a32-4c8d-b4f8-c7a7c5a396fd', '0d791003-2a0d-4a0a-920e-86509399c611', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '2 hours 15 minutes');

-- 進行中の部屋（MHP3） - host: 素材コレクター（満室）
INSERT INTO room_members (id, room_id, user_id, player_number, is_host, status, joined_at) VALUES
  (gen_random_uuid(), '06ac86fe-3a8a-4dde-85b4-90df1f262155', '8cd1a74f-aa82-4e78-9ffb-21f77bac6294', 1, true, 'active', CURRENT_TIMESTAMP - INTERVAL '1 hour'),
  (gen_random_uuid(), '06ac86fe-3a8a-4dde-85b4-90df1f262155', '509e4f15-eb60-4811-8486-3d23f0928419', 2, false, 'active', CURRENT_TIMESTAMP - INTERVAL '55 minutes'),
  (gen_random_uuid(), '06ac86fe-3a8a-4dde-85b4-90df1f262155', '0d791003-2a0d-4a0a-920e-86509399c611', 3, false, 'active', CURRENT_TIMESTAMP - INTERVAL '50 minutes'),
  -- 4人目のダミーユーザー（満室にするため）
  (gen_random_uuid(), '06ac86fe-3a8a-4dde-85b4-90df1f262155', 'fb9595e5-c036-4e56-9b19-7d445e191191', 4, false, 'active', CURRENT_TIMESTAMP - INTERVAL '45 minutes');

-- 統計情報の確認用クエリ
-- SELECT r.name as room_name, COUNT(rm.id) as member_count, r.status 
-- FROM rooms r 
-- LEFT JOIN room_members rm ON r.id = rm.room_id AND rm.status = 'active' 
-- GROUP BY r.id, r.name, r.status 
-- ORDER BY r.created_at;