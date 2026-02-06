## TTL(Time To Live)とは

**Time To Live**: 指定した時刻を過ぎたアイテムを自動的に削除する機能。

### メリット

- ストレージコスト削減
- 手動削除不要
- 削除処理のスループット消費なし

---

## 仕組み

1. アイテムに TTL属性（Unix Epoch秒）を設定
2. DynamoDBが定期的にスキャン（バックグラウンド）
3. 有効期限切れアイテムを自動削除

### 注意点

- 削除は**即座ではない**（最大48時間の遅延）
- 削除順序は保証されない
- GSIからの削除も遅延する可能性

---

## TTL有効化（AWS CLI）

```bash
aws dynamodb update-time-to-live \
    --table-name DynamoDBShop \
    --time-to-live-specification "Enabled=true, AttributeName=TTL"
```

### 確認

```bash
aws dynamodb describe-time-to-live --table-name DynamoDBShop
```

---

## TTL属性の形式

**Unix Epoch秒**（ミリ秒ではない）

```go
// 30日後に削除
ttl := time.Now().Add(30 * 24 * time.Hour).Unix()

// 7日後に削除
ttl := time.Now().Add(7 * 24 * time.Hour).Unix()

// 特定日時に削除
expireAt := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
ttl := expireAt.Unix()
```

---

## ユーザー行動ログの実装例

### データ構造

```go
type UserActivity struct {
    PK         string    // USER#<userId>
    SK         string    // ACTIVITY#<timestamp>
    UserID     string
    ActionType string    // VIEW, CLICK, ADD_CART, PURCHASE
    ProductID  string
    Metadata   map[string]string
    TTL        int64     // Unix Epoch秒（30日後）
    CreatedAt  time.Time
}
```

### 行動ログの記録

```go
func (r *ActivityRepository) Log(ctx context.Context, activity *UserActivity) error {
    now := time.Now()
    ttl := now.Add(30 * 24 * time.Hour).Unix() // 30日後に自動削除

    item, _ := attributevalue.MarshalMap(map[string]interface{}{
        "PK":         "USER#" + activity.UserID,
        "SK":         "ACTIVITY#" + now.Format(time.RFC3339Nano),
        "EntityType": "ACTIVITY",
        "UserID":     activity.UserID,
        "ActionType": activity.ActionType,
        "ProductID":  activity.ProductID,
        "Metadata":   activity.Metadata,
        "TTL":        ttl,
        "CreatedAt":  now.Format(time.RFC3339),
    })

    _, err := r.client.PutItem(ctx, &dynamodb.PutItemInput{
        TableName: aws.String(r.tableName),
        Item:      item,
    })
    return err
}
```

### フロントエンドからの呼び出し例

```javascript
// 商品閲覧時
api.post('/activity', {
  actionType: 'VIEW',
  productId: 'prod001',
  metadata: {
    referrer: document.referrer,
    duration: '30s'
  }
})

// カート追加時
api.post('/activity', {
  actionType: 'ADD_CART',
  productId: 'prod001',
  metadata: {
    quantity: 2
  }
})
```

---

## BatchWriteItemでの一括記録

高頻度のログを効率的に書き込み。

```go
func (r *ActivityRepository) BatchLog(ctx context.Context, activities []UserActivity) error {
    ttl := time.Now().Add(30 * 24 * time.Hour).Unix()

    writeRequests := make([]types.WriteRequest, 0, len(activities))

    for _, activity := range activities {
        item, _ := attributevalue.MarshalMap(map[string]interface{}{
            "PK":         "USER#" + activity.UserID,
            "SK":         "ACTIVITY#" + activity.CreatedAt.Format(time.RFC3339Nano),
            "EntityType": "ACTIVITY",
            "ActionType": activity.ActionType,
            "ProductID":  activity.ProductID,
            "Metadata":   activity.Metadata,
            "TTL":        ttl,
        })

        writeRequests = append(writeRequests, types.WriteRequest{
            PutRequest: &types.PutRequest{Item: item},
        })
    }

    // 25件ずつバッチ処理（DynamoDB制限）
    for i := 0; i < len(writeRequests); i += 25 {
        end := i + 25
        if end > len(writeRequests) {
            end = len(writeRequests)
        }

        _, err := r.client.BatchWriteItem(ctx, &dynamodb.BatchWriteItemInput{
            RequestItems: map[string][]types.WriteRequest{
                r.tableName: writeRequests[i:end],
            },
        })
        if err != nil {
            return err
        }
    }

    return nil
}
```

---

## TTL削除のStreams連携

TTLで削除されたアイテムもDynamoDB Streamsに流れる。

```go
// Lambda で TTL削除を検知
func handler(ctx context.Context, event events.DynamoDBEvent) error {
    for _, record := range event.Records {
        // TTL削除は eventName = "REMOVE" かつ userIdentity が含まれる
        if record.EventName == "REMOVE" {
            // record.UserIdentity が nil でなければ TTL による削除
            if record.UserIdentity != nil &&
               record.UserIdentity.Type == "Service" &&
               record.UserIdentity.PrincipalID == "dynamodb.amazonaws.com" {
                log.Println("TTL による削除:", record.Change.Keys)
                // アーカイブ処理など
            }
        }
    }
    return nil
}
```

---

## TTLの使用例

| ユースケース | TTL期間 |
|-------------|---------|
| セッションデータ | 24時間 |
| 一時トークン | 1時間 |
| 行動ログ | 30日 |
| キャッシュデータ | 1時間〜1日 |
| 通知履歴 | 7日 |

---

## TTL削除を待たずに明示削除

TTLは遅延があるため、即座に削除したい場合は明示的に削除。

```go
// TTL有効期限前に手動削除
_, err := client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
    TableName: aws.String("DynamoDBShop"),
    Key: map[string]types.AttributeValue{
        "PK": &types.AttributeValueMemberS{Value: "USER#u001"},
        "SK": &types.AttributeValueMemberS{Value: "ACTIVITY#2024-01-15T10:00:00Z"},
    },
})
```

---

## クエリ時のTTL考慮

TTL切れアイテムが削除されるまでの間、クエリ結果に含まれる可能性がある。

```go
// フィルタで除外
result, err := client.Query(ctx, &dynamodb.QueryInput{
    TableName:              aws.String("DynamoDBShop"),
    KeyConditionExpression: aws.String("PK = :pk"),
    FilterExpression:       aws.String("TTL > :now"),
    ExpressionAttributeValues: map[string]types.AttributeValue{
        ":pk":  &types.AttributeValueMemberS{Value: "USER#u001"},
        ":now": &types.AttributeValueMemberN{Value: fmt.Sprintf("%d", time.Now().Unix())},
    },
})
```

---

## 学習メモ

```
ここに学習中のメモを追記
```
