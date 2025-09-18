-- ユーザープロフィール機能拡張のためのマイグレーション
-- 実行前に必ずバックアップを取得してください

-- 1. usersテーブルの拡張
ALTER TABLE users 
    ADD COLUMN IF NOT EXISTS favorite_games JSONB DEFAULT '[]',
    ADD COLUMN IF NOT EXISTS play_times JSONB DEFAULT '{}';

-- favorite_gamesカラムにインデックスを追加（JSONB検索性能向上）
CREATE INDEX IF NOT EXISTS idx_users_favorite_games ON users USING GIN (favorite_games);

-- 2. フレンド（フォロー）機能のためのテーブル作成
CREATE TABLE IF NOT EXISTS user_follows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    follower_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    following_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    accepted_at TIMESTAMP,
    CONSTRAINT unique_follow_pair UNIQUE(follower_user_id, following_user_id),
    CONSTRAINT no_self_follow CHECK (follower_user_id != following_user_id)
);

-- インデックスの作成
CREATE INDEX IF NOT EXISTS idx_user_follows_follower ON user_follows(follower_user_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_following ON user_follows(following_user_id);
CREATE INDEX IF NOT EXISTS idx_user_follows_status ON user_follows(status);
CREATE INDEX IF NOT EXISTS idx_user_follows_accepted ON user_follows(follower_user_id, following_user_id) WHERE status = 'accepted';

-- 3. フレンド数を取得する関数
CREATE OR REPLACE FUNCTION get_friend_count(p_user_id UUID)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM user_follows uf1
        WHERE uf1.follower_user_id = p_user_id
        AND uf1.status = 'accepted'
        AND EXISTS (
            SELECT 1
            FROM user_follows uf2
            WHERE uf2.follower_user_id = uf1.following_user_id
            AND uf2.following_user_id = uf1.follower_user_id
            AND uf2.status = 'accepted'
        )
    );
END;
$$ LANGUAGE plpgsql;

-- 4. 相互フォロー（フレンド）を取得するビュー
CREATE OR REPLACE VIEW user_friends AS
SELECT 
    uf1.follower_user_id as user_id,
    uf1.following_user_id as friend_user_id,
    u.username as friend_username,
    u.display_name as friend_display_name,
    u.avatar_url as friend_avatar_url,
    GREATEST(uf1.accepted_at, uf2.accepted_at) as friend_since
FROM user_follows uf1
INNER JOIN user_follows uf2 
    ON uf1.follower_user_id = uf2.following_user_id
    AND uf1.following_user_id = uf2.follower_user_id
INNER JOIN users u ON u.id = uf1.following_user_id
WHERE uf1.status = 'accepted' 
AND uf2.status = 'accepted';

-- 5. サンプルデータの挿入（開発環境用）
-- 注意: 本番環境では実行しないでください
DO $$
BEGIN
    -- 開発環境チェック（ENVが'development'の場合のみ実行）
    IF current_setting('app.env', true) = 'development' THEN
        -- サンプルユーザーのfavorite_gamesとplay_timesを更新
        UPDATE users 
        SET 
            favorite_games = '["MHP3", "MHP2G"]'::jsonb,
            play_times = '{"weekday": "19:00-23:00", "weekend": "13:00-17:00"}'::jsonb
        WHERE email LIKE '%@example.com'
        LIMIT 5;
    END IF;
END $$;

-- 6. マイグレーション完了メッセージ
DO $$
BEGIN
    RAISE NOTICE 'User profile migration completed successfully';
    RAISE NOTICE 'Added columns: users.favorite_games, users.play_times';
    RAISE NOTICE 'Created table: user_follows';
    RAISE NOTICE 'Created function: get_friend_count';
    RAISE NOTICE 'Created view: user_friends';
END $$;