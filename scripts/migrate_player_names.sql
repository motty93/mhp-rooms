-- player_namesテーブルの構造を更新するマイグレーションスクリプト

-- 1. 新しいカラムを追加
ALTER TABLE player_names ADD COLUMN IF NOT EXISTS game_version_id UUID;

-- 2. 既存のデータを変換（game_versionの文字列からgame_version_idへ）
UPDATE player_names pn
SET game_version_id = gv.id
FROM game_versions gv
WHERE pn.game_version = gv.code
AND pn.game_version_id IS NULL;

-- 3. 外部キー制約を追加
ALTER TABLE player_names 
ADD CONSTRAINT fk_player_names_game_version 
FOREIGN KEY (game_version_id) 
REFERENCES game_versions(id) 
ON DELETE CASCADE;

-- 4. NOT NULL制約を追加
ALTER TABLE player_names ALTER COLUMN game_version_id SET NOT NULL;

-- 5. ユニーク制約を追加（ユーザーとゲームバージョンの組み合わせ）
ALTER TABLE player_names 
ADD CONSTRAINT uk_player_names_user_game 
UNIQUE (user_id, game_version_id);

-- 6. インデックスを作成
CREATE INDEX IF NOT EXISTS idx_player_names_user_game 
ON player_names(user_id, game_version_id);

-- 7. 旧カラムを削除（オプション：データ移行確認後に実行）
-- ALTER TABLE player_names DROP COLUMN game_version;

-- 8. カラム名を変更（player_name -> name）
ALTER TABLE player_names RENAME COLUMN player_name TO name;