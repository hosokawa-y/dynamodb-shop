#!/bin/bash
#
# Lambda 関数 (inventory-stream-handler) のデプロイスクリプト
#
# 処理内容:
#   1. IAM ロール作成 (なければ)
#   2. Lambda バイナリのビルド (Linux/amd64, custom runtime 用 bootstrap)
#   3. ZIP パッケージング
#   4. Lambda 関数の作成 / 更新
#   5. DynamoDB Streams のイベントソースマッピング作成
#
# Usage: ./setup-lambda.sh

set -e

# ====================================================================
# 設定
# ====================================================================
TABLE_NAME="${DYNAMODB_TABLE:-DynamoDBShop}"
REGION="${AWS_REGION:-ap-northeast-1}"
FUNCTION_NAME="inventory-stream-handler"
ROLE_NAME="dynamodb-shop-lambda-role"
RUNTIME="provided.al2023"
HANDLER="bootstrap"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LAMBDA_DIR="$PROJECT_ROOT/lambda/inventory-stream-handler"
BUILD_DIR="$LAMBDA_DIR/build"
ZIP_FILE="$BUILD_DIR/function.zip"

echo "=============================================="
echo "Lambda Setup: $FUNCTION_NAME"
echo "  Region: $REGION"
echo "  Table:  $TABLE_NAME"
echo "=============================================="

# ====================================================================
# 1. IAM ロール作成
# ====================================================================
echo ""
echo "[1/5] Setting up IAM role: $ROLE_NAME"

TRUST_POLICY='{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Principal": {"Service": "lambda.amazonaws.com"},
            "Action": "sts:AssumeRole"
        }
    ]
}'

if aws iam get-role --role-name "$ROLE_NAME" >/dev/null 2>&1; then
    echo "  Role $ROLE_NAME already exists. Skipping creation."
else
    echo "  Creating role $ROLE_NAME..."
    aws iam create-role \
        --role-name "$ROLE_NAME" \
        --assume-role-policy-document "$TRUST_POLICY" \
        >/dev/null

    # DynamoDB Streams 読み取り + CloudWatch Logs 書き込み権限
    aws iam attach-role-policy \
        --role-name "$ROLE_NAME" \
        --policy-arn arn:aws:iam::aws:policy/service-role/AWSLambdaDynamoDBExecutionRole

    echo "  Role created. Waiting for IAM propagation (10s)..."
    sleep 10
fi

ROLE_ARN=$(aws iam get-role --role-name "$ROLE_NAME" --query 'Role.Arn' --output text)
echo "  Role ARN: $ROLE_ARN"

# ====================================================================
# 2. Lambda バイナリをビルド
# ====================================================================
echo ""
echo "[2/5] Building Lambda binary (linux/amd64)..."

mkdir -p "$BUILD_DIR"
rm -f "$BUILD_DIR/bootstrap" "$ZIP_FILE"

cd "$LAMBDA_DIR"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 \
    go build -tags lambda.norpc -o "$BUILD_DIR/bootstrap" .

echo "  Built: $BUILD_DIR/bootstrap"

# ====================================================================
# 3. ZIP パッケージング
# ====================================================================
echo ""
echo "[3/5] Packaging zip..."

cd "$BUILD_DIR"
zip -j "$ZIP_FILE" bootstrap >/dev/null
echo "  Created: $ZIP_FILE"

# ====================================================================
# 4. Lambda 関数の作成 / 更新
# ====================================================================
echo ""
echo "[4/5] Creating/updating Lambda function: $FUNCTION_NAME"

if aws lambda get-function --function-name "$FUNCTION_NAME" --region "$REGION" >/dev/null 2>&1; then
    echo "  Function exists. Updating code..."
    aws lambda update-function-code \
        --function-name "$FUNCTION_NAME" \
        --zip-file "fileb://$ZIP_FILE" \
        --region "$REGION" \
        >/dev/null
    aws lambda wait function-updated \
        --function-name "$FUNCTION_NAME" \
        --region "$REGION"
    echo "  Code updated."
else
    echo "  Creating new function..."
    aws lambda create-function \
        --function-name "$FUNCTION_NAME" \
        --runtime "$RUNTIME" \
        --role "$ROLE_ARN" \
        --handler "$HANDLER" \
        --zip-file "fileb://$ZIP_FILE" \
        --timeout 30 \
        --memory-size 128 \
        --region "$REGION" \
        >/dev/null
    aws lambda wait function-active \
        --function-name "$FUNCTION_NAME" \
        --region "$REGION"
    echo "  Function created."
fi

# ====================================================================
# 5. DynamoDB Streams イベントソースマッピング
# ====================================================================
echo ""
echo "[5/5] Setting up event source mapping (DynamoDB Streams → Lambda)"

STREAM_ARN=$(aws dynamodb describe-table \
    --table-name "$TABLE_NAME" \
    --region "$REGION" \
    --query 'Table.LatestStreamArn' \
    --output text)

if [ -z "$STREAM_ARN" ] || [ "$STREAM_ARN" = "None" ]; then
    echo "ERROR: Stream is not enabled on table $TABLE_NAME."
    echo "       Run create-table.sh first or enable streams manually."
    exit 1
fi

echo "  Stream ARN: $STREAM_ARN"

# 既存のマッピングをチェック
EXISTING_UUID=$(aws lambda list-event-source-mappings \
    --function-name "$FUNCTION_NAME" \
    --region "$REGION" \
    --query "EventSourceMappings[?EventSourceArn=='$STREAM_ARN'].UUID | [0]" \
    --output text)

if [ -n "$EXISTING_UUID" ] && [ "$EXISTING_UUID" != "None" ]; then
    echo "  Event source mapping already exists (UUID: $EXISTING_UUID). Skipping."
else
    echo "  Creating event source mapping..."
    aws lambda create-event-source-mapping \
        --function-name "$FUNCTION_NAME" \
        --event-source-arn "$STREAM_ARN" \
        --starting-position LATEST \
        --batch-size 10 \
        --region "$REGION" \
        >/dev/null
    echo "  Event source mapping created."
fi

echo ""
echo "=============================================="
echo "Lambda setup completed successfully!"
echo "=============================================="
echo ""
echo "動作確認:"
echo "  1. 商品の在庫を更新 (例: stock を 11 → 5 に変更)"
echo "  2. CloudWatch Logs を確認:"
echo "     aws logs tail /aws/lambda/$FUNCTION_NAME --follow --region $REGION"
echo ""
