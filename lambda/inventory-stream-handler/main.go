// lambda/inventory-stream-handler/main.go
//
// DynamoDB Streams トリガーで起動し、在庫変動を検知して
// 閾値以下になった場合に LOW STOCK ALERT をログ出力する Lambda 関数
//
// 【学習ポイント】
// - Streams は INSERT/MODIFY/REMOVE の3種のイベントを送信する
// - NEW_AND_OLD_IMAGES 設定により、変更前後の両方の値を取得できる
// - DynamoDB の AttributeValue は events パッケージの型で扱う
// - Lambda 関数は events.DynamoDBEvent を受け取る
package main

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// 在庫が下回ったら LOW STOCK ALERTを出す閾値
const LowStockThreshold = 20

// handlerはDynamoDB streamsからのイベントを処理する
//
// 【events.DynamoDBEvent】
//
//	Records: 複数のレコード（最大1000件）が一度に渡される
//	各 Record は INSERT/MODIFY/REMOVE のいずれかのイベント
func handler(ctx context.Context, event events.DynamoDBEvent) error {
	log.Printf("Received %d records", len(event.Records))

	for _, record := range event.Records {
		// 1. 商品レコード（PRODUCT#xx, METADATA)以外は無視
		// 	keysはevents.DynamoDBAttributeValue型のマップ
		pk := record.Change.Keys["PK"].String()
		sk := record.Change.Keys["SK"].String()
		if !strings.HasPrefix(pk, "PRODUCT#") || sk != "METADATA" {
			continue
		}

		// 2. MODIFYイベントのみ処理（INSERT/REMOVEは対象外)
		if record.EventName != string(events.DynamoDBOperationTypeModify) {
			continue
		}

		// 3. 在庫の変動を検知
		oldStock, err := getIntAttr(record.Change.OldImage, "stock")
		if err != nil {
			log.Printf("Failed to get old stock: %v", err)
		}
		newStock, err := getIntAttr(record.Change.NewImage, "stock")
		if err != nil {
			log.Printf("Failed to get new stock: %v", err)
		}

		// 在庫に変動がない場合はスキップ
		if oldStock == newStock {
			continue
		}

		productID := strings.TrimPrefix(pk, "PRODUCT#")
		productName := record.Change.NewImage["name"].String()

		log.Printf("[STOCK CHANGE] product=%s name=%s %d -> %d", productID, productName, oldStock, newStock)

		// 4. 閾値判定：閾値以下になった瞬間にアラートを出す
		if oldStock > LowStockThreshold && newStock <= LowStockThreshold {
			log.Printf("LOW STOCK ALERT: product=%s name=%s stock=%d (threshold=%d)", productID, productName, newStock, LowStockThreshold)
		}

		// 5. 在庫切れアラート
		if newStock == 0 {
			log.Printf("OUT OF STOCK: product=%s name=%s", productID, productName)
		}
	}

	return nil
}

// getIntAttr は events.DynamoDBAttributeValue から数値を取り出すヘルパー
//
// DynamoDB の数値は文字列として送られてくるため、strconv.Atoi で変換する
// (events パッケージの仕様: Number は string で表現される)
func getIntAttr(image map[string]events.DynamoDBAttributeValue, key string) (int, error) {
	attr, ok := image[key]
	if !ok {
		return 0, nil
	}
	return strconv.Atoi(attr.Number())
}

func main() {
	lambda.Start(handler)
}
