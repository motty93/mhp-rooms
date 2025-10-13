-- 20部屋分のテストデータを追加するSQL
-- ページネーション表示確認用
-- 実行方法: turso db shell mhp-rooms-motty93 < scripts/seed_20_rooms.sql

-- ゲームバージョンIDを取得（実際の環境に合わせて調整してください）
-- MHP, MHP2, MHP2G, MHP3のIDを想定

-- 20部屋分のテストデータを挿入
INSERT INTO rooms (id, room_code, name, description, game_version_id, host_user_id, max_players, current_players, target_monster, rank_requirement, is_active, is_closed, created_at, updated_at) VALUES
-- 1-5部屋目
(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-001', 'ティガレックス討伐募集', 'HR6以上、装備自由です',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'ハンター太郎' LIMIT 1),
 4, 1, 'ティガレックス', 'HR6以上', true, false, datetime('now', '-2 hours'), datetime('now', '-2 hours')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-002', '初心者歓迎！リオレウス', '下位クエストでゆっくり楽しみましょう',
 (SELECT id FROM game_versions WHERE code = 'MHP2' LIMIT 1),
 (SELECT id FROM users WHERE display_name = '猫好きハンター🐱' LIMIT 1),
 4, 2, 'リオレウス', '制限なし', true, false, datetime('now', '-1 hour 50 minutes'), datetime('now', '-1 hour 50 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-003', 'レア素材周回部屋', 'リオレウス逆鱗狙いで効率周回',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = '素材コレクター' LIMIT 1),
 4, 3, 'リオレウス亜種', 'HR5以上', true, false, datetime('now', '-1 hour 40 minutes'), datetime('now', '-1 hour 40 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-004', 'ナルガクルガ討伐', '上位ナルガクルガ、弓限定',
 (SELECT id FROM game_versions WHERE code = 'MHP3' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'スピードランナー' LIMIT 1),
 4, 1, 'ナルガクルガ', 'HR7以上', true, false, datetime('now', '-1 hour 30 minutes'), datetime('now', '-1 hour 30 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-005', 'まったり探索部屋', 'のんびり素材集めしましょう',
 (SELECT id FROM game_versions WHERE code = 'MHP' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'まったりハンター' LIMIT 1),
 4, 2, NULL, '制限なし', true, false, datetime('now', '-1 hour 20 minutes'), datetime('now', '-1 hour 20 minutes')),

-- 6-10部屋目
(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-006', 'グラビモス採掘ツアー', '採掘とグラビモス討伐',
 (SELECT id FROM game_versions WHERE code = 'MHP2' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'プロハンター' LIMIT 1),
 4, 1, 'グラビモス', 'HR4以上', true, false, datetime('now', '-1 hour 10 minutes'), datetime('now', '-1 hour 10 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-007', '双剣使い限定部屋', '双剣で楽しく狩りましょう',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = '武器マスター' LIMIT 1),
 4, 2, 'ディアブロス', 'HR6以上', true, false, datetime('now', '-1 hour'), datetime('now', '-1 hour')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-008', 'モンスター研究会', '生態観察とデータ収集',
 (SELECT id FROM game_versions WHERE code = 'MHP3' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'モンスター博士' LIMIT 1),
 4, 1, NULL, '制限なし', true, false, datetime('now', '-50 minutes'), datetime('now', '-50 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-009', 'レアアイテム交換会', '不要なレアアイテムを交換',
 (SELECT id FROM game_versions WHERE code = 'MHP2' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'アイテム商人' LIMIT 1),
 4, 3, NULL, '制限なし', true, false, datetime('now', '-40 minutes'), datetime('now', '-40 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-010', 'チーム戦略研究室', '効率的な立ち回りを学ぼう',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'チームリーダー' LIMIT 1),
 4, 2, 'ラージャン', 'HR8以上', true, false, datetime('now', '-30 minutes'), datetime('now', '-30 minutes')),

-- 11-15部屋目
(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-011', 'クシャルダオラ討伐', '毒武器推奨です',
 (SELECT id FROM game_versions WHERE code = 'MHP2' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'ハンター太郎' LIMIT 1),
 4, 1, 'クシャルダオラ', 'HR7以上', true, false, datetime('now', '-25 minutes'), datetime('now', '-25 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-012', '古龍種討伐隊', '上位古龍を次々討伐',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = '猫好きハンター🐱' LIMIT 1),
 4, 3, 'テオ・テスカトル', 'HR8以上', true, false, datetime('now', '-20 minutes'), datetime('now', '-20 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-013', 'イャンクック連続狩猟', '初心者練習用',
 (SELECT id FROM game_versions WHERE code = 'MHP' LIMIT 1),
 (SELECT id FROM users WHERE display_name = '素材コレクター' LIMIT 1),
 4, 1, 'イャンクック', '制限なし', true, false, datetime('now', '-15 minutes'), datetime('now', '-15 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-014', 'タイムアタック部屋', 'リタマラ推奨、速度重視',
 (SELECT id FROM game_versions WHERE code = 'MHP3' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'スピードランナー' LIMIT 1),
 4, 2, 'ジンオウガ', 'HR7以上', true, false, datetime('now', '-10 minutes'), datetime('now', '-10 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-015', '飛竜種マラソン', '飛竜種を次々狩りましょう',
 (SELECT id FROM game_versions WHERE code = 'MHP2' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'まったりハンター' LIMIT 1),
 4, 1, 'リオレイア', 'HR4以上', true, false, datetime('now', '-8 minutes'), datetime('now', '-8 minutes')),

-- 16-20部屋目
(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-016', 'ガンナー専用部屋', 'ガンナーで連携プレイ',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'プロハンター' LIMIT 1),
 4, 2, 'モノブロス', 'HR5以上', true, false, datetime('now', '-6 minutes'), datetime('now', '-6 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-017', '亜種モンスター狩猟', '亜種の素材集め',
 (SELECT id FROM game_versions WHERE code = 'MHP3' LIMIT 1),
 (SELECT id FROM users WHERE display_name = '武器マスター' LIMIT 1),
 4, 1, 'ベリオロス亜種', 'HR6以上', true, false, datetime('now', '-5 minutes'), datetime('now', '-5 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-018', '大連続狩猟クエスト', '複数モンスター討伐',
 (SELECT id FROM game_versions WHERE code = 'MHP2' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'モンスター博士' LIMIT 1),
 4, 3, NULL, 'HR5以上', true, false, datetime('now', '-4 minutes'), datetime('now', '-4 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-019', '装備強化相談室', '装備のアドバイスします',
 (SELECT id FROM game_versions WHERE code = 'MHP2G' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'アイテム商人' LIMIT 1),
 4, 2, NULL, '制限なし', true, false, datetime('now', '-2 minutes'), datetime('now', '-2 minutes')),

(lower(hex(randomblob(4))) || '-' || lower(hex(randomblob(2))) || '-4' || substr(lower(hex(randomblob(2))),2) || '-' || substr('89ab',abs(random()) % 4 + 1, 1) || substr(lower(hex(randomblob(2))),2) || '-' || lower(hex(randomblob(6))),
 'ROOM-TEST-020', 'フルフル討伐会', '閃光玉持参推奨',
 (SELECT id FROM game_versions WHERE code = 'MHP' LIMIT 1),
 (SELECT id FROM users WHERE display_name = 'チームリーダー' LIMIT 1),
 4, 1, 'フルフル', 'HR3以上', true, false, datetime('now', '-1 minute'), datetime('now', '-1 minute'));

-- 確認用クエリ
SELECT
    COUNT(*) as total_rooms,
    SUM(CASE WHEN is_active = true THEN 1 ELSE 0 END) as active_rooms
FROM rooms;

SELECT
    name,
    room_code,
    (SELECT code FROM game_versions WHERE id = rooms.game_version_id) as game_version,
    current_players,
    max_players,
    created_at
FROM rooms
WHERE room_code LIKE 'ROOM-TEST-%'
ORDER BY created_at DESC
LIMIT 20;
