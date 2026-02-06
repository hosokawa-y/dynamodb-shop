## DynamoDB Streamsとは

テーブルへの変更（INSERT, MODIFY, REMOVE）をリアルタイムにキャプチャするストリーム。

### ユースケース

- データ変更のトリガー処理
- データ同期（別テーブル、Elasticsearch等）
- 監査ログ
- リアルタイム通知
- 集計・分析

---

## Streams有効化

### テーブル作成時

```bash
aws dynamodb create-table \
    --table-name DynamoDBShop \
    ... \
    --stream-specification StreamEnabled=true,StreamViewType=NEW_AND_OLD_IMAGES
```

### 既存テーブルに追加

```bash
aws dynamodb update-table \
    --table-name DynamoDBShop \
    --stream-specification StreamEnabled=true,StreamViewType=NEW_AND_OLD_IMAGES
```

---

## StreamViewType

| タイプ | 内容 |
|--------|------|
| KEYS_ONLY | キー属性のみ |
| NEW_IMAGE | 変更後のアイテム全体 |
| OLD_IMAGE | 変更前のアイテム全体 |
| NEW_AND_OLD_IMAGES | 変更前後の両方 |

> **推奨**: `NEW_AND_OLD_IMAGES`（差分検知が可能）

---

## Lambda関数の作成（Go）

### プロジェクト構造

```
lambda/
└── inventory-stream-handler/
    ├── main.go
    ├── go.mod
    └── Makefile
```

### main.go

```go
package main

import (
    "context"
    "log"
    "strconv"

    "github.com/aws/aws-lambda-go/events"
    "github.com/aws/aws-lambda-go/lambda"
)

const lowStockThreshold = 10

func handler(ctx context.Context, event events.DynamoDBEvent) error {
    for _, record := range event.Records {
        log.Printf("EventName: %s", record.EventName)

        // 商品データの変更のみ処理
        pk := record.Change.Keys["PK"].String()
        sk := record.Change.Keys["SK"].String()

        if sk != "METADATA" {
            continue
        }

        switch record.EventName {
        case "MODIFY":
            handleModify(record)
        case "INSERT":
            handleInsert(record)
        case "REMOVE":
            handleRemove(record)
        }
    }
    return nil
}

func handleModify(record events.DynamoDBEventRecord) {
    newImage := record.Change.NewImage
    oldImage := record.Change.OldImage

    // 商品以外はスキップ
    entityType, ok := newImage["EntityType"]
    if !ok || entityType.String() != "PRODUCT" {
        return
    }

    // 在庫変動をチェック
    newStockStr := newImage["Stock"].Number()
    oldStockStr := oldImage["Stock"].Number()

    newStock, _ := strconv.Atoi(newStockStr)
    oldStock, _ := strconv.Atoi(oldStockStr)

    productName := newImage["Name"].String()
    productID := newImage["ProductId"].String()

    // 在庫が減少した場合
    if newStock < oldStock {
        log.Printf("在庫減少: %s (%s) %d -> %d", productName, productID, oldStock, newStock)
    }

    // 在庫が閾値以下になった場合のアラート
    if oldStock > lowStockThreshold && newStock <= lowStockThreshold {
        log.Printf("LOW STOCK ALERT: %s (%s) - 残り %d 個", productName, productID, newStock)
        // ここで SNS 通知や Slack 通知を送信
        // sendLowStockAlert(ctx, productID, productName, newStock)
    }

    // 在庫がゼロになった場合
    if newStock == 0 && oldStock > 0 {
        log.Printf("OUT OF STOCK: %s (%s)", productName, productID)
        // sendOutOfStockAlert(ctx, productID, productName)
    }
}

func handleInsert(record events.DynamoDBEventRecord) {
    newImage := record.Change.NewImage
    log.Printf("新規アイテム追加: PK=%s", newImage["PK"].String())
}

func handleRemove(record events.DynamoDBEventRecord) {
    oldImage := record.Change.OldImage
    log.Printf("アイテム削除: PK=%s", oldImage["PK"].String())
}

func main() {
    lambda.Start(handler)
}
```

---

## Lambdaのデプロイ

### go.mod

```go
module inventory-stream-handler

go 1.21

require (
    github.com/aws/aws-lambda-go v1.41.0
)
```

### Makefile

```makefile
.PHONY: build deploy

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go
	zip function.zip bootstrap

deploy: build
	aws lambda update-function-code \
		--function-name dynamodb-shop-inventory-handler \
		--zip-file fileb://function.zip
```

### 初回セットアップスクリプト

```bash
#!/bin/bash
# infrastructure/scripts/setup-lambda.sh

FUNCTION_NAME="dynamodb-shop-inventory-handler"
ROLE_NAME="dynamodb-shop-lambda-role"
TABLE_NAME="DynamoDBShop"

# 1. Lambda実行ロール作成
aws iam create-role \
    --role-name $ROLE_NAME \
    --assume-role-policy-document '{
        "Version": "2012-10-17",
        "Statement": [{
            "Effect": "Allow",
            "Principal": {"Service": "lambda.amazonaws.com"},
            "Action": "sts:AssumeRole"
        }]
    }'

# 2. ポリシーアタッチ
aws iam attach-role-policy \
    --role-name $ROLE_NAME \
    --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole

aws iam attach-role-policy \
    --role-name $ROLE_NAME \
    --policy-arn arn:aws:iam::aws:policy/AmazonDynamoDBReadOnlyAccess

# 3. ロールARN取得
ROLE_ARN=$(aws iam get-role --role-name $ROLE_NAME --query 'Role.Arn' --output text)
sleep 10  # ロール作成の反映を待つ

# 4. Lambda関数ビルド
cd lambda/inventory-stream-handler
make build

# 5. Lambda関数作成
aws lambda create-function \
    --function-name $FUNCTION_NAME \
    --runtime provided.al2023 \
    --handler bootstrap \
    --zip-file fileb://function.zip \
    --role $ROLE_ARN \
    --timeout 30 \
    --memory-size 128

# 6. Streams ARN取得
STREAM_ARN=$(aws dynamodb describe-table \
    --table-name $TABLE_NAME \
    --query 'Table.LatestStreamArn' \
    --output text)

# 7. イベントソースマッピング作成
aws lambda create-event-source-mapping \
    --function-name $FUNCTION_NAME \
    --event-source-arn $STREAM_ARN \
    --starting-position LATEST \
    --batch-size 100 \
    --maximum-batching-window-in-seconds 5

echo "Lambda setup complete!"
```

---

## イベントソースマッピング設定

| パラメータ | 説明 | 推奨値 |
|-----------|------|--------|
| batch-size | 1回のLambda呼び出しで処理するレコード数 | 100 |
| maximum-batching-window-in-seconds | バッチ収集の最大待機時間 | 5 |
| starting-position | 開始位置 | LATEST |
| parallelization-factor | シャードあたりの同時実行数 | 1-10 |

---

## エラーハンドリングとリトライ

### デフォルト動作

- Lambda失敗時、DynamoDB Streamsは自動リトライ
- 同じバッチを最大でレコードの有効期限（24時間）まで再試行

### 失敗時の設定

```bash
aws lambda update-event-source-mapping \
    --uuid <mapping-uuid> \
    --destination-config '{
        "OnFailure": {
            "Destination": "arn:aws:sqs:ap-northeast-1:123456789:dlq"
        }
    }' \
    --maximum-retry-attempts 3 \
    --maximum-record-age-in-seconds 3600
```

---

## Streamsレコード構造

```go
type DynamoDBEventRecord struct {
    EventID        string                       // ユニークID
    EventName      string                       // INSERT, MODIFY, REMOVE
    EventSource    string                       // aws:dynamodb
    EventVersion   string
    EventSourceArn string
    Change         DynamoDBStreamRecord
    UserIdentity   *events.DynamoDBUserIdentity // TTL削除の場合に設定
}

type DynamoDBStreamRecord struct {
    Keys           map[string]events.DynamoDBAttributeValue
    NewImage       map[string]events.DynamoDBAttributeValue  // 変更後
    OldImage       map[string]events.DynamoDBAttributeValue  // 変更前
    SequenceNumber string
    SizeBytes      int64
    StreamViewType string
}
```

---

## 実践的なパターン

### 1. 在庫アラート（このプロジェクト）

```go
if oldStock > threshold && newStock <= threshold {
    // SNS通知
}
```

### 2. データ同期（Elasticsearch等）

```go
if record.EventName == "INSERT" || record.EventName == "MODIFY" {
    indexToElasticsearch(record.Change.NewImage)
} else if record.EventName == "REMOVE" {
    deleteFromElasticsearch(record.Change.Keys)
}
```

### 3. 監査ログ

```go
auditLog := AuditLog{
    EventType: record.EventName,
    Timestamp: time.Now(),
    Before:    record.Change.OldImage,
    After:     record.Change.NewImage,
}
saveToAuditTable(auditLog)
```

### 4. 集計更新

```go
// 注文作成時に日次売上を更新
if entityType == "ORDER" && record.EventName == "INSERT" {
    updateDailySales(orderDate, orderAmount)
}
```

---

## デバッグ

### CloudWatch Logsの確認

```bash
aws logs tail /aws/lambda/dynamodb-shop-inventory-handler --follow
```

### ローカルテスト

```go
// test_event.json
{
  "Records": [
    {
      "eventName": "MODIFY",
      "dynamodb": {
        "Keys": {"PK": {"S": "PRODUCT#p001"}, "SK": {"S": "METADATA"}},
        "NewImage": {"Stock": {"N": "5"}, "Name": {"S": "テスト商品"}},
        "OldImage": {"Stock": {"N": "15"}, "Name": {"S": "テスト商品"}}
      }
    }
  ]
}

// ローカル実行
sam local invoke -e test_event.json
```

---

## 学習メモ

```
ここに学習中のメモを追記
```
