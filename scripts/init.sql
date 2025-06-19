-- 開発環境用の初期化スクリプト

-- データベースが存在しない場合は作成
-- CREATE DATABASE mhp_rooms_dev;

-- 接続
\c mhp_rooms_dev;

-- 拡張機能の有効化
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 開発用データの挿入例（必要に応じて）
-- ゲームバージョンのマスターデータ
-- GORMのマイグレーション後に実行される
-- INSERT INTO game_versions (id, code, name, display_order, is_active, created_at) VALUES
--   (uuid_generate_v4(), 'MHP', 'モンスターハンターポータブル', 1, true, CURRENT_TIMESTAMP),
--   (uuid_generate_v4(), 'MHP2', 'モンスターハンターポータブル 2nd', 2, true, CURRENT_TIMESTAMP),
--   (uuid_generate_v4(), 'MHP2G', 'モンスターハンターポータブル 2nd G', 3, true, CURRENT_TIMESTAMP),
--   (uuid_generate_v4(), 'MHP3', 'モンスターハンターポータブル 3rd', 4, true, CURRENT_TIMESTAMP)
-- ON CONFLICT (code) DO NOTHING;

-- 開発完了時にはこのコメントを外してテーブルを作成
-- 現在はGORMのマイグレーション機能を使用予定