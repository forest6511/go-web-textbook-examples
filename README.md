# go-web-textbook-examples

書籍『現場で使えるGo言語Webアプリ開発: Gin・認証・Cloud Runまで実装で学ぶ本番設計』の公式サンプルコードです。

章ごとに独立した `go.mod` でモジュールを分けており、`chNN-*/` ディレクトリをそのまま clone して実行・検証できます。

## 動作環境

- Go 1.26 以上
- Docker / Docker Compose（第 4 章以降の PostgreSQL、第 8 章以降の MinIO 等で使用）

## ディレクトリ構成（予定）

| 章 | ディレクトリ | 内容 |
|----|-------------|------|
| 01 | `ch01-gin-intro/` | Gin 入門と net/http との比較 |
| 02 | `ch02-routing/` | ルーティングとリクエスト処理 |
| 03 | `ch03-middleware/` | ミドルウェア設計 |
| 04 | `ch04-postgres-sqlc/` | PostgreSQL + pgx + sqlc |
| 05 | `ch05-validation/` | バリデーションとエラーハンドリング |
| 06 | `ch06-auth-concepts/` | 認証: Session vs JWT |
| 07 | `ch07-jwt-impl/` | JWT 実装（signup / login / refresh） |
| 08 | `ch08-file-upload/` | ファイルアップロード / S3 互換ストレージ |
| 09 | `ch09-slog-otel/` | 構造化ログと OpenTelemetry |
| 10 | `ch10-testing/` | テスト戦略 |
| 11 | `ch11-deploy/` | Docker / Cloud Run / Fly.io デプロイ |
| 12 | `ch12-production/` | 本番運用 |
| 付録A | `appendix-a-echo/` | Echo への移植ガイド |
| 付録B | `appendix-b-gorm-vs-sqlc/` | GORM vs sqlc 比較 |

## 使い方

```bash
git clone https://github.com/forest6511/go-web-textbook-examples.git
cd go-web-textbook-examples/ch01-gin-intro
go run ./cmd/api
```

## ライセンス

MIT License
