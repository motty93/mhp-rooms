-- マルチプラットフォーム対応のためのデータベーススキーマ拡張
-- 実行前に必ずバックアップを取得してください

-- 1. プラットフォーム設定テーブルの作成
CREATE TABLE IF NOT EXISTS platform_configs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    platform VARCHAR(10) NOT NULL UNIQUE,
    display_name VARCHAR(50) NOT NULL,
    display_name_en VARCHAR(50),
    network_type VARCHAR(30) NOT NULL,
    network_description TEXT,
    setup_guide_url VARCHAR(255),
    icon_path VARCHAR(255),
    color_theme VARCHAR(7), -- hex color
    display_order INT NOT NULL DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 2. game_versionsテーブルの拡張
ALTER TABLE game_versions 
    ADD COLUMN IF NOT EXISTS platform VARCHAR(10) NOT NULL DEFAULT 'PSP',
    ADD COLUMN IF NOT EXISTS network_type VARCHAR(30) NOT NULL DEFAULT 'adhoc_party',
    ADD COLUMN IF NOT EXISTS platform_display_name VARCHAR(50);

-- 3. platform列にインデックスを追加（検索性能向上のため）
CREATE INDEX IF NOT EXISTS idx_game_versions_platform ON game_versions(platform);

-- 4. roomsテーブルの拡張（オプション：プラットフォーム直接参照用）
ALTER TABLE rooms 
    ADD COLUMN IF NOT EXISTS platform VARCHAR(10);

-- platform列を自動設定するトリガー
CREATE OR REPLACE FUNCTION set_room_platform()
RETURNS TRIGGER AS $$
BEGIN
    SELECT platform INTO NEW.platform
    FROM game_versions
    WHERE id = NEW.game_version_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_room_platform_trigger
BEFORE INSERT OR UPDATE ON rooms
FOR EACH ROW
EXECUTE FUNCTION set_room_platform();

-- 5. 初期データの投入

-- プラットフォーム設定
INSERT INTO platform_configs (platform, display_name, display_name_en, network_type, network_description, icon_path, color_theme, display_order) VALUES
('PSP', 'PSP', 'PSP', 'adhoc_party', 'PS3のアドホックパーティを使用してオンラインプレイ', '/static/images/platforms/psp.png', '#0055CC', 1),
('3DS', '3DS', '3DS', 'pretendo', 'Pretendoネットワークを使用してオンラインプレイ', '/static/images/platforms/3ds.png', '#D52B1E', 2),
('WIIU', 'Wii U', 'Wii U', 'pretendo', 'Pretendoネットワークを使用してオンラインプレイ', '/static/images/platforms/wiiu.png', '#00A652', 3)
ON CONFLICT (platform) DO NOTHING;

-- 既存のgame_versionsデータを更新
UPDATE game_versions 
SET 
    platform = 'PSP',
    network_type = 'adhoc_party',
    platform_display_name = 'PSP'
WHERE platform IS NULL OR platform = '';

-- 6. 新しいゲームバージョンの追加（3DS/WiiU用）
-- ※既存のデータと重複しないように、存在チェックを含む
INSERT INTO game_versions (code, name, platform, network_type, platform_display_name, display_order, is_active) VALUES
-- 3DS向け
('MH3', 'モンスターハンター3（トライ）', '3DS', 'pretendo', '3DS', 21, true),
('MH3G', 'モンスターハンター3G', '3DS', 'pretendo', '3DS', 22, true),
('MH4', 'モンスターハンター4', '3DS', 'pretendo', '3DS', 23, true),
('MH4G', 'モンスターハンター4G', '3DS', 'pretendo', '3DS', 24, true),
-- Wii U向け
('MH3U', 'モンスターハンター3G HD', 'WIIU', 'pretendo', 'Wii U', 31, true)
ON CONFLICT (code) DO NOTHING;

-- 7. ビューの作成（プラットフォーム別のルーム表示用）
CREATE OR REPLACE VIEW rooms_by_platform AS
SELECT 
    r.*,
    gv.platform,
    gv.network_type,
    pc.display_name as platform_display_name,
    pc.color_theme as platform_color
FROM rooms r
JOIN game_versions gv ON r.game_version_id = gv.id
LEFT JOIN platform_configs pc ON gv.platform = pc.platform
WHERE r.is_active = true AND r.is_closed = false;

-- 8. 統計用のマテリアライズドビュー（オプション）
CREATE MATERIALIZED VIEW IF NOT EXISTS platform_statistics AS
SELECT 
    gv.platform,
    COUNT(DISTINCT r.id) as total_rooms,
    COUNT(DISTINCT r.host_user_id) as unique_hosts,
    SUM(r.current_players) as total_players,
    COUNT(DISTINCT CASE WHEN r.created_at > NOW() - INTERVAL '24 hours' THEN r.id END) as rooms_last_24h
FROM rooms r
JOIN game_versions gv ON r.game_version_id = gv.id
WHERE r.is_active = true
GROUP BY gv.platform;

-- マテリアライズドビューのリフレッシュ用インデックス
CREATE UNIQUE INDEX IF NOT EXISTS idx_platform_statistics_platform ON platform_statistics(platform);

-- 9. 関数の作成：プラットフォーム別のアクティブルーム数取得
CREATE OR REPLACE FUNCTION get_active_rooms_by_platform(p_platform VARCHAR)
RETURNS INTEGER AS $$
BEGIN
    RETURN (
        SELECT COUNT(*)
        FROM rooms r
        JOIN game_versions gv ON r.game_version_id = gv.id
        WHERE gv.platform = p_platform
        AND r.is_active = true
        AND r.is_closed = false
    );
END;
$$ LANGUAGE plpgsql;

-- 10. 既存データの整合性チェック
DO $$
BEGIN
    -- game_versionsのplatform列が正しく設定されているか確認
    IF EXISTS (SELECT 1 FROM game_versions WHERE platform IS NULL OR platform = '') THEN
        RAISE NOTICE 'Warning: Some game_versions records have NULL or empty platform values';
    END IF;
    
    -- roomsテーブルのplatform列を更新
    UPDATE rooms r
    SET platform = gv.platform
    FROM game_versions gv
    WHERE r.game_version_id = gv.id
    AND r.platform IS NULL;
    
    RAISE NOTICE 'Migration completed successfully';
END $$;