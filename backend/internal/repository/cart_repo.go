// cart_repo.go
// カートデータのDynamoDB操作を担当するリポジトリ
//
// 【キー設計】
//   PK: USER#<ユーザーID>     - パーティションキー（ユーザー単位）
//   SK: CART#<商品ID>        - ソートキー（商品単位）
//
// 【アクセスパターン】
//   1. ユーザーのカート全件取得  → Query(PK = "USER#xxx" AND begins_with(SK, "CART#"))
//   2. カートアイテム1件取得    → GetItem(PK, SK)
//   3. カートにアイテム追加     → PutItem
//   4. 数量更新（楽観的ロック）  → UpdateItem + ConditionExpression
//   5. カートからアイテム削除   → DeleteItem

package repository

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"github.com/hosokawa-y/dynamodb-shop/backend/internal/domain"
	"github.com/hosokawa-y/dynamodb-shop/backend/pkg/timeutil"
)
