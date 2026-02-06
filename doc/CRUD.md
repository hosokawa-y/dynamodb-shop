## 基本操作一覧

| 操作 | API | 用途 |
|------|-----|------|
| 作成 | PutItem | 新規アイテム作成 |
| 読み取り | GetItem | 単一アイテム取得（PK+SK指定） |
| 読み取り | Query | 条件に合うアイテム取得（PKは必須） |
| 読み取り | Scan | テーブル全体をスキャン（非推奨） |
| 更新 | UpdateItem | 属性の部分更新 |
| 削除 | DeleteItem | アイテム削除 |
| 一括 | BatchGetItem | 複数アイテム一括取得 |
| 一括 | BatchWriteItem | 複数アイテム一括書き込み/削除 |

---

## PutItem - 作成

```go
// Go (AWS SDK v2)
_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
    TableName: aws.String("DynamoDBShop"),
    Item: map[string]types.AttributeValue{
        "PK":   &types.AttributeValueMemberS{Value: "USER#u001"},
        "SK":   &types.AttributeValueMemberS{Value: "PROFILE"},
        "Name": &types.AttributeValueMemberS{Value: "田中太郎"},
        "Email": &types.AttributeValueMemberS{Value: "tanaka@example.com"},
    },
})
```

### 上書き防止（存在しない場合のみ作成）

```go
_, err := client.PutItem(ctx, &dynamodb.PutItemInput{
    TableName: aws.String("DynamoDBShop"),
    Item: item,
    ConditionExpression: aws.String("attribute_not_exists(PK)"),
})
```

---

## GetItem - 単一取得

```go
result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
    TableName: aws.String("DynamoDBShop"),
    Key: map[string]types.AttributeValue{
        "PK": &types.AttributeValueMemberS{Value: "USER#u001"},
        "SK": &types.AttributeValueMemberS{Value: "PROFILE"},
    },
})

// result.Item に結果が入る
```

### 結合一貫性（強い整合性）

```go
result, err := client.GetItem(ctx, &dynamodb.GetItemInput{
    TableName:      aws.String("DynamoDBShop"),
    Key:            key,
    ConsistentRead: aws.Bool(true), // 強い整合性
})
```

---

## Query - 条件取得

PKは必須、SKは任意で条件指定可能。

```go
// ユーザーのカート全取得
result, err := client.Query(ctx, &dynamodb.QueryInput{
    TableName:              aws.String("DynamoDBShop"),
    KeyConditionExpression: aws.String("PK = :pk AND begins_with(SK, :sk)"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk": &types.AttributeValueMemberS{Value: "USER#u001"},
        ":sk": &types.AttributeValueMemberS{Value: "CART#"},
    },
})
```

### SK条件演算子

| 演算子 | 例 |
|--------|-----|
| `=` | `SK = :sk` |
| `begins_with` | `begins_with(SK, :prefix)` |
| `between` | `SK BETWEEN :start AND :end` |
| `<`, `<=`, `>`, `>=` | `SK >= :value` |

### FilterExpression（結果のフィルタ）

```go
result, err := client.Query(ctx, &dynamodb.QueryInput{
    TableName:              aws.String("DynamoDBShop"),
    KeyConditionExpression: aws.String("PK = :pk"),
    FilterExpression:       aws.String("Price > :minPrice"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk":       &types.AttributeValueMemberS{Value: "CATEGORY#電子機器"},
        ":minPrice": &types.AttributeValueMemberN{Value: "5000"},
    },
})
```

> **注意**: FilterExpressionは**取得後**にフィルタするため、RCU（読み込み容量）は節約されない

---

## Scan - 全体スキャン

```go
// 全商品取得（非推奨：大量データでは遅い）
result, err := client.Scan(ctx, &dynamodb.ScanInput{
    TableName:        aws.String("DynamoDBShop"),
    FilterExpression: aws.String("EntityType = :type"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":type": &types.AttributeValueMemberS{Value: "PRODUCT"},
    },
})
```

> **警告**: テーブル全体を読むため、本番では避ける。GSIを使ったQueryに置き換える。

---

## UpdateItem - 部分更新

```go
_, err := client.UpdateItem(ctx, &dynamodb.UpdateItemInput{
    TableName: aws.String("DynamoDBShop"),
    Key: map[string]types.AttributeValue{
        "PK": &types.AttributeValueMemberS{Value: "PRODUCT#p001"},
        "SK": &types.AttributeValueMemberS{Value: "METADATA"},
    },
    UpdateExpression: aws.String("SET Price = :price, Stock = Stock - :qty"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":price": &types.AttributeValueMemberN{Value: "1200"},
        ":qty":   &types.AttributeValueMemberN{Value: "1"},
    },
})
```

### UpdateExpression構文

| 操作 | 構文 | 例 |
|------|-----|-----|
| 設定 | `SET` | `SET Name = :name` |
| 削除 | `REMOVE` | `REMOVE OldAttr` |
| 加算 | `ADD` | `ADD ViewCount :one` |
| リスト追加 | `SET` | `SET Tags = list_append(Tags, :newTag)` |

---

## DeleteItem - 削除

```go
_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
    TableName: aws.String("DynamoDBShop"),
    Key: map[string]types.AttributeValue{
        "PK": &types.AttributeValueMemberS{Value: "USER#u001"},
        "SK": &types.AttributeValueMemberS{Value: "CART#p001"},
    },
})
```

---

## BatchWriteItem - 一括書き込み

最大25アイテムまで1回のリクエストで処理。

```go
_, err := client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
    RequestItems: map[string][]types.WriteRequest{
        "DynamoDBShop": {
            {PutRequest: &types.PutRequest{Item: item1}},
            {PutRequest: &types.PutRequest{Item: item2}},
            {DeleteRequest: &types.DeleteRequest{Key: key1}},
        },
    },
})
```

> **注意**: BatchWriteItemはトランザクションではない。部分的に失敗する可能性あり。

---

## 学習メモ

```
ここに学習中のメモを追記
```
