-- quest_typeカラムをroomsテーブルから削除するマイグレーション
-- 実行日: 2025-07-14

-- quest_typeカラムが存在するかチェックしてから削除
DO $$
BEGIN
    IF EXISTS (
        SELECT column_name 
        FROM information_schema.columns 
        WHERE table_name = 'rooms' 
        AND column_name = 'quest_type'
    ) THEN
        EXECUTE 'ALTER TABLE rooms DROP COLUMN quest_type';
        RAISE NOTICE 'quest_type カラムを削除しました';
    ELSE
        RAISE NOTICE 'quest_type カラムは既に存在しません';
    END IF;
END
$$;