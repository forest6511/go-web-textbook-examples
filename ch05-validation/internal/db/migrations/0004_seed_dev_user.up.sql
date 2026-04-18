-- Ch 04-06 用の開発ユーザー。認証を導入する Ch 07 までは、handler が
-- devUserID = 1 をハードコードで usecase に渡すため、tasks テーブルの
-- 外部キー制約 (user_id REFERENCES users(id)) が通るように
-- 固定 ID=1 のダミーユーザーをシードしておく。
-- Ch 07 以降はシードではなく signup エンドポイントで実ユーザーを作る。
INSERT INTO users (id, email, password)
    VALUES (1, 'dev@localhost', 'dev')
    ON CONFLICT (id) DO NOTHING;

-- BIGSERIAL のシーケンスを既存レコードの最大 id に合わせて進める。
-- これをしないと「id=1 を手動で挿入したあと、次の INSERT が id=1 で採番される」
-- というシーケンス競合が起きる。
SELECT setval(
    pg_get_serial_sequence('users', 'id'),
    GREATEST((SELECT COALESCE(MAX(id), 1) FROM users), 1)
);
