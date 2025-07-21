-- メッセージリアクション機能のマイグレーション

-- 1. リアクションタイプのマスターテーブル
CREATE TABLE IF NOT EXISTS reaction_types (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    emoji VARCHAR(10) NOT NULL,
    display_order INTEGER NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- インデックス
CREATE INDEX idx_reaction_types_code ON reaction_types(code);
CREATE INDEX idx_reaction_types_is_active ON reaction_types(is_active);

-- 2. メッセージリアクションテーブル
CREATE TABLE IF NOT EXISTS message_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    message_id UUID NOT NULL REFERENCES room_messages(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reaction_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    
    -- 同一ユーザーが同じメッセージに同じリアクションを複数回つけられないように
    CONSTRAINT unique_user_message_reaction UNIQUE (message_id, user_id, reaction_type),
    
    -- reaction_typeは必ずreaction_types.codeに存在する値である必要がある
    CONSTRAINT fk_reaction_type FOREIGN KEY (reaction_type) REFERENCES reaction_types(code)
);

-- インデックス
CREATE INDEX idx_message_reactions_message_id ON message_reactions(message_id);
CREATE INDEX idx_message_reactions_user_id ON message_reactions(user_id);
CREATE INDEX idx_message_reactions_reaction_type ON message_reactions(reaction_type);
CREATE INDEX idx_message_reactions_created_at ON message_reactions(created_at DESC);

-- 3. デフォルトのリアクションタイプを挿入
INSERT INTO reaction_types (code, name, emoji, display_order) VALUES
    ('like', 'いいね', '👍', 1),
    ('heart', 'ハート', '❤️', 2),
    ('laugh', '笑い', '😄', 3),
    ('surprised', '驚き', '😮', 4),
    ('sad', '悲しい', '😢', 5),
    ('angry', '怒り', '😠', 6),
    ('fire', '炎', '🔥', 7),
    ('party', 'パーティー', '🎉', 8)
ON CONFLICT (code) DO NOTHING;

-- 4. リアクション集計用のビュー（パフォーマンス向上のため）
CREATE OR REPLACE VIEW message_reaction_counts AS
SELECT 
    mr.message_id,
    mr.reaction_type,
    rt.emoji,
    rt.name AS reaction_name,
    COUNT(mr.user_id) AS reaction_count,
    ARRAY_AGG(mr.user_id ORDER BY mr.created_at) AS user_ids
FROM message_reactions mr
JOIN reaction_types rt ON mr.reaction_type = rt.code
WHERE rt.is_active = true
GROUP BY mr.message_id, mr.reaction_type, rt.emoji, rt.name;

-- 5. 特定ユーザーのリアクション状態を取得する関数
CREATE OR REPLACE FUNCTION get_user_reactions_for_messages(
    p_user_id UUID,
    p_message_ids UUID[]
) RETURNS TABLE (
    message_id UUID,
    reaction_types TEXT[]
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        mr.message_id,
        ARRAY_AGG(mr.reaction_type ORDER BY mr.created_at) AS reaction_types
    FROM message_reactions mr
    WHERE mr.user_id = p_user_id
      AND mr.message_id = ANY(p_message_ids)
    GROUP BY mr.message_id;
END;
$$ LANGUAGE plpgsql;

-- 6. updated_atの自動更新トリガー
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_reaction_types_updated_at BEFORE UPDATE ON reaction_types
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();