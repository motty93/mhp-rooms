-- Add og_version column to rooms table for OGP image versioning
-- SQLiteでは ADD COLUMN IF NOT EXISTS はサポートされていないため、
-- カラムが存在しない場合のみ追加する
ALTER TABLE rooms ADD COLUMN og_version INTEGER NOT NULL DEFAULT 0;
