# DynamoDB Shop

DynamoDBの主要機能を学習するためのEコマースアプリケーション。

## 技術スタック

- **Backend**: Go
- **Frontend**: Vue.js
- **Database**: AWS DynamoDB
- **Infrastructure**: AWS Lambda (Streams処理)

## 学習する機能

- Single Table Design
- GSI (グローバルセカンダリインデックス)
- 条件付き書き込み・楽観的ロック
- トランザクション
- 時系列データモデリング
- TTL (Time To Live)
- DynamoDB Streams + Lambda

## プロジェクト構成

```
dynamodb-shop/
├── backend/          # Go API サーバー
├── frontend/         # Vue.js フロントエンド
├── lambda/           # DynamoDB Streams 処理
└── infrastructure/   # AWS リソース設定スクリプト
```

## セットアップ

### 1. DynamoDB テーブル作成

```bash
./infrastructure/scripts/create-table.sh
```

### 2. Backend 起動

```bash
cd backend
cp .env.example .env
# .env を編集してAWS認証情報を設定
go run cmd/api/main.go
```

### 3. Frontend 起動

```bash
cd frontend
npm install
npm run dev
```

## API エンドポイント

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/auth/register | 会員登録 |
| POST | /api/v1/auth/login | ログイン |
| GET | /api/v1/products | 商品一覧 |
| GET | /api/v1/products/:id | 商品詳細 |
| GET | /api/v1/cart | カート取得 |
| POST | /api/v1/orders | 注文確定 |

## ライセンス

MIT
