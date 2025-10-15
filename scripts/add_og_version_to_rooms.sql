-- Add og_version column to rooms table for OGP image versioning
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS og_version INTEGER NOT NULL DEFAULT 0;

-- Add comment
COMMENT ON COLUMN rooms.og_version IS 'OGP画像のバージョン番号。部屋の作成・更新時にインクリメントされる';
