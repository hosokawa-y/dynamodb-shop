## GSIとは

**Global Secondary Index**: メインテーブルとは異なるPK/SKでデータにアクセスするためのインデックス。

### なぜ必要か

メインテーブルのPK/SKでは対応できないアクセスパターンを実現するため。

```
例: 商品テーブル
- メインキー: PK=PRODUCT#id, SK=METADATA
- 問題: 「カテゴリで商品を検索したい」→ PKがカテゴリではないので効率的に取得できない
- 解決: GSI を作成し、カテゴリをPKにする
```

---

## GSI vs LSI

| 項目 | GSI | LSI |
|------|-----|-----|
| PK | 任意の属性 | メインテーブルと同じ |
| SK | 任意の属性 | 任意の属性 |
| 作成タイミング | いつでも | テーブル作成時のみ |
| 整合性 | 結果整合性のみ | 強い整合性も可能 |
| 容量制限 | なし | 10GB/パーティション |
| 数の上限 | 20個/テーブル | 5個/テーブル |

> **結論**: ほとんどの場合GSIを使う。LSIは特殊なケースのみ。

---

## GSI設計（今回のプロジェクト）

### GSI1: カテゴリ検索・メール検索・月別注文

| 用途 | GSI1PK | GSI1SK |
|------|--------|--------|
| カテゴリ別商品 | `CATEGORY#電子機器` | `PRODUCT#p001` |
| メールでユーザー検索 | `EMAIL#user@example.com` | `USER` |
| 月別注文一覧 | `ORDERS#2024-01` | `2024-01-15T10:00:00Z#ord001` |

### GSI2: ステータス検索

| 用途 | GSI2PK | GSI2SK |
|------|--------|--------|
| ステータス別注文 | `STATUS#CONFIRMED` | `ORDER#ord001` |
| 在庫少商品 | `STOCK#LOW` | `PRODUCT#p001` |

---

## GSI作成（AWS CLI）

```bash
aws dynamodb update-table \
    --table-name DynamoDBShop \
    --attribute-definitions \
        AttributeName=GSI1PK,AttributeType=S \
        AttributeName=GSI1SK,AttributeType=S \
    --global-secondary-index-updates '[
        {
            "Create": {
                "IndexName": "GSI1",
                "KeySchema": [
                    {"AttributeName": "GSI1PK", "KeyType": "HASH"},
                    {"AttributeName": "GSI1SK", "KeyType": "RANGE"}
                ],
                "Projection": {"ProjectionType": "ALL"}
            }
        }
    ]'
```

---

## GSI Projection（射影）

GSIに含める属性を指定。

| タイプ | 説明 | ストレージ | 用途 |
|--------|------|-----------|------|
| `ALL` | 全属性 | 大 | 全データが必要な場合 |
| `KEYS_ONLY` | キーのみ | 小 | IDだけ取得してから詳細取得 |
| `INCLUDE` | 指定属性のみ | 中 | 特定属性だけ必要な場合 |

```bash
# INCLUDE の例
"Projection": {
    "ProjectionType": "INCLUDE",
    "NonKeyAttributes": ["Name", "Price"]
}
```

---

## GSI使用例（Go）

### カテゴリ別商品検索

```go
result, err := client.Query(ctx, &dynamodb.QueryInput{
    TableName:              aws.String("DynamoDBShop"),
    IndexName:              aws.String("GSI1"), // GSI指定
    KeyConditionExpression: aws.String("GSI1PK = :pk"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "CATEGORY#電子機器"},
    },
})
```

### メールでユーザー検索

```go
result, err := client.Query(ctx, &dynamodb.QueryInput{
    TableName:              aws.String("DynamoDBShop"),
    IndexName:              aws.String("GSI1"),
    KeyConditionExpression: aws.String("GSI1PK = :email"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":email": &types.AttributeValueMemberS{Value: "EMAIL#user@example.com"},
    },
})
```

---

## GSI設計のポイント

### 1. オーバーロード（多目的利用）

1つのGSIで複数のアクセスパターンに対応。

```
GSI1PK の値:
- CATEGORY#xxx  → カテゴリ検索
- EMAIL#xxx     → メール検索
- ORDERS#yyyy-mm → 月別注文

同じGSI1を異なる目的で使い回す
```

### 2. スパースインデックス

GSIキー属性を持つアイテムのみがインデックスに含まれる。

```
# GSI1PKを持たないアイテムはGSI1に含まれない
# → インデックスサイズとコストを削減
```

### 3. GSIのコスト

- ストレージ: メインテーブル+GSIの両方で課金
- 書き込み: メインテーブル書き込み時にGSIも更新される
- → 必要最小限のGSIに抑える

---

## よくある間違い

### 1. Scanで代用しようとする

```go
// ダメな例: Scanでカテゴリフィルタ
Scan(FilterExpression: "Category = :cat")

// 良い例: GSIでQuery
Query(IndexName: "GSI1", GSI1PK = "CATEGORY#xxx")
```

### 2. GSIに強い整合性を期待する

GSIは**結果整合性のみ**。書き込み直後のQueryで最新データが返らない可能性あり。

```
対策:
- 書き込み直後の読み取りはメインテーブルから
- または少し待ってからGSI Query
```

---

## 学習メモ

```
ここに学習中のメモを追記
```
