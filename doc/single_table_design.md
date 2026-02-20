# DynamoDB - Single Table Design

---

## Single Table Designとは

1つのテーブルで複数のエンティティ（ユーザー、商品、注文など）を管理する設計パターン。

### なぜ必要か

DynamoDBは**JOINができない**ため、関連データを効率的に取得するには：
1. 複数回のクエリを実行する（遅い）
2. 1つのテーブルに関連データをまとめる（推奨）

---

## PK/SK設計の基本

```
PK (Partition Key): データの分散単位
SK (Sort Key): PK内でのソート・範囲検索
```

### 設計パターン

| エンティティ | PK | SK |
|------------|----|----|
| ユーザー | `USER#<userId>` | `PROFILE` |
| 商品 | `PRODUCT#<productId>` | `METADATA` |
| カート | `USER#<userId>` | `CART#<productId>` |
| 注文 | `USER#<userId>` | `ORDER#<orderId>` |
| 注文明細 | `ORDER#<orderId>` | `ITEM#<productId>` |

### プレフィックスを付ける理由

```
# 悪い例
PK: "user123"
PK: "product456"

# 良い例
PK: "USER#user123"
PK: "PRODUCT#product456"
```

- 視覚的に区別しやすい
- begins_withで特定種類のデータを取得可能
- 将来の拡張性

---

## アクセスパターン駆動設計

**RDB**: データ構造を先に設計 → 後からクエリを考える
**DynamoDB**: アクセスパターンを先に洗い出す → それに合わせてキー設計

### アクセスパターン例

| No | パターン | 使用するキー |
|----|---------|-------------|
| 1 | ユーザー情報取得 | `PK=USER#id, SK=PROFILE` |
| 2 | ユーザーのカート全取得 | `PK=USER#id, SK begins_with CART#` |
| 3 | ユーザーの注文履歴 | `PK=USER#id, SK begins_with ORDER#` |
| 4 | 注文の明細一覧 | `PK=ORDER#id, SK begins_with ITEM#` |

---

## 実装例（今回のプロジェクト）

```
テーブル名: DynamoDBShop

# ユーザー
PK: USER#u001          SK: PROFILE
{name: "田中太郎", email: "tanaka@example.com"}

# 同じユーザーのカート
PK: USER#u001          SK: CART#prod001
{productId: "prod001", quantity: 2}

PK: USER#u001          SK: CART#prod002
{productId: "prod002", quantity: 1}

# 同じユーザーの注文
PK: USER#u001          SK: ORDER#ord001
{status: "CONFIRMED", totalAmount: 5000}
```

### 1回のQueryで取得できるデータ

```go
// ユーザーのカート全取得
Query(PK = "USER#u001" AND SK begins_with "CART#")

// ユーザーの全データ取得（プロフィール+カート+注文）
Query(PK = "USER#u001")
```

---

## 注意点

### 1. ホットパーティション
特定のPKにアクセスが集中するとスロットリング発生
→ PKの分散を意識する

### 2. アイテムサイズ制限
1アイテム最大400KB
→ 大きなデータは分割または S3 に保存

### 3. GSIの活用
PKだけでは対応できないアクセスパターンはGSIで解決

---

## 学習メモ

### DynamoDBにはデータベースという概念がない
  DynamoDBの構造

  AWS DynamoDB (サービス)
  └── テーブル (例: DynamoDBShop)  ← 今回作成したもの
      └── アイテム (データ1件1件)

  RDBMSとの比較:

  | RDBMS        | DynamoDB         |
  |--------------|------------------|
  | データベース | ― (概念なし)     |
  | テーブル     | テーブル         |
  | 行 (Row)     | アイテム (Item)  |
  | 列 (Column)  | 属性 (Attribute) |

  ---
  Single Table Design

  このプロジェクトでは Single Table Design を採用しています。

  - 従来のRDBMS: エンティティごとにテーブル (users, products, orders...)
  - DynamoDB: 1つのテーブルに全エンティティを格納

  PK / SK のプレフィックスでエンティティを区別:
  USER#123           | PROFILE            → ユーザー情報
  PRODUCT#456  | METADATA        → 商品情報
  USER#123           | ORDER#789      → 注文情報

### PKとSKの複合キーによる一意性

DynamoDBでは**PKとSKの組み合わせ**がプライマリキー（複合キー）となる。

```
┌─────────────────────────────┬─────────────────────────┐
│ PK                          │ SK                      │
├─────────────────────────────┼─────────────────────────┤
│ USER#22aa6fd7-...           │ PROFILE                 │ ← 一意
│ USER#22aa6fd7-...           │ ORDER#b3e80d89-...      │ ← 一意
│ USER#22aa6fd7-...           │ ORDER#xxxxxxxx-...      │ ← 別の注文も追加可能
│ USER#22aa6fd7-...           │ CART#yyyyyyyy-...       │ ← カートも追加可能
└─────────────────────────────┴─────────────────────────┘
```

| ルール | 説明 |
|--------|------|
| PK単独 | 重複OK |
| SK単独 | 重複OK |
| PK + SK | **一意である必要がある** |

同じPKを持つデータは**1回のQuery**で取得できる：
```go
// USER#22aa6fd7-... に紐づく全データを取得
// → PROFILE, ORDER, CART などが一度に取れる
KeyConditionExpression: "PK = :pk"
```

これにより、ユーザー情報・注文履歴・カートを個別にクエリする必要がなく、
効率的にデータを取得できる。

