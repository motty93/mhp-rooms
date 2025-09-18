-- 既存の部屋のcurrent_playersを修正するスクリプト
-- ホストを含めた人数にする

BEGIN;

-- 進行中の部屋の人数を修正（ホスト1人 + メンバー1人 = 2人）
UPDATE rooms SET current_players = 2 WHERE id = 'eee173f4-30b1-40b7-b233-f5a5b5da577e';

-- 確認
SELECT 
    r.id,
    r.name,
    r.current_players,
    (SELECT COUNT(*) FROM room_members rm WHERE rm.room_id = r.id AND rm.status = 'active') as active_members,
    u.display_name as host_name
FROM rooms r
JOIN users u ON r.host_user_id = u.id
WHERE r.is_active = true
ORDER BY r.created_at DESC;

COMMIT;