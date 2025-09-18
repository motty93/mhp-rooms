-- supabase_user_idカラムにインデックスを追加
-- SLOW SQL対策（286ms -> 数ms）

-- 既存のインデックスがあれば削除
DROP INDEX IF EXISTS idx_users_supabase_user_id;

-- 新しいインデックスを作成
CREATE INDEX idx_users_supabase_user_id ON users(supabase_user_id);

-- インデックスの確認
SELECT name, sql FROM sqlite_master 
WHERE type = 'index' 
AND tbl_name = 'users' 
AND name LIKE '%supabase%';