-- ユーザーアクティビティテーブルのマイグレーション
-- GitHub Issue #21: プロフィール画面 アクティビティ

-- user_activitiesテーブル作成
CREATE TABLE IF NOT EXISTS user_activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    activity_type VARCHAR(50) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    related_entity_type VARCHAR(50), -- 'room', 'user', 'message'など
    related_entity_id UUID,
    metadata JSONB DEFAULT '{}',
    icon VARCHAR(100), -- Font Awesomeアイコンクラス
    icon_color VARCHAR(50), -- Tailwind CSSカラークラス
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- インデックス作成
-- ユーザー別時系列取得用（最も重要）
CREATE INDEX IF NOT EXISTS idx_user_activities_user_created 
ON user_activities(user_id, created_at DESC);

-- アクティビティタイプ別フィルタリング用
CREATE INDEX IF NOT EXISTS idx_user_activities_type 
ON user_activities(activity_type);

-- 全体の時系列ソート用（管理機能等で使用）
CREATE INDEX IF NOT EXISTS idx_user_activities_created_at 
ON user_activities(created_at DESC);

-- 関連エンティティ検索用（特定の部屋・ユーザーに関するアクティビティ検索）
CREATE INDEX IF NOT EXISTS idx_user_activities_related_entity 
ON user_activities(related_entity_type, related_entity_id);

-- アクティビティタイプ定数のコメント
COMMENT ON TABLE user_activities IS 'ユーザーの行動履歴を記録するテーブル';
COMMENT ON COLUMN user_activities.activity_type IS '
アクティビティタイプ:
- room_create: 部屋作成
- room_join: 部屋参加
- room_leave: 部屋退出
- room_close: 部屋終了
- room_update: 部屋設定変更
- follow_add: フォロー開始
- follow_accept: フォロー承認
- follow_remove: フォロー解除
- message_send: メッセージ送信
- profile_update: プロフィール更新
- user_join: ユーザー登録
';
COMMENT ON COLUMN user_activities.metadata IS 'アクティビティの詳細情報をJSON形式で保存';
COMMENT ON COLUMN user_activities.related_entity_type IS '関連エンティティのタイプ（room, user, messageなど）';
COMMENT ON COLUMN user_activities.related_entity_id IS '関連エンティティのID（部屋ID、ユーザーIDなど）';