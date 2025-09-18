-- 既存の部屋のcurrent_playersを修正するスクリプト
-- ホストを含めた人数にするため、メンバー数+1にする

BEGIN;

-- 全ての部屋のcurrent_playersを一時的に0にする
UPDATE rooms SET current_players = 0;

-- アクティブなメンバー数を数えて更新
UPDATE rooms r
SET current_players = (
    SELECT COUNT(*) + 1  -- +1はホストの分
    FROM room_members rm
    WHERE rm.room_id = r.id
    AND rm.status = 'active'
)
WHERE r.is_active = true;

-- 確認のためのクエリ
SELECT 
    r.id,
    r.name,
    r.current_players as updated_count,
    (SELECT COUNT(*) FROM room_members rm WHERE rm.room_id = r.id AND rm.status = 'active') as member_count,
    u.display_name as host_name
FROM rooms r
JOIN users u ON r.host_user_id = u.id
WHERE r.is_active = true
ORDER BY r.created_at DESC;

COMMIT;