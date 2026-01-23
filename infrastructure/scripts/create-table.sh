#!/bin/bash
#
# DynamoDB テーブル作成スクリプト
# Usage: ./create-table.sh [--local]
#

set -e

TABLE_NAME="DynamoDBShop"
REGION="${AWS_REGION:-ap-northeast-1}"

# ローカル開発モードのチェック
if [ "$1" = "--local" ]; then
    ENDPOINT="--endpoint-url http://localhost:8000"
    echo "Creating table in DynamoDB Local..."
else
    ENDPOINT=""
    echo "Creating table in AWS DynamoDB (Region: $REGION)..."
fi

# テーブルが既に存在するかチェック
if aws dynamodb describe-table --table-name $TABLE_NAME $ENDPOINT --region $REGION 2>/dev/null; then
    echo "Table $TABLE_NAME already exists."
    exit 0
fi

# テーブル作成
echo "Creating table: $TABLE_NAME"

aws dynamodb create-table \
    --table-name $TABLE_NAME \
    --attribute-definitions \
        AttributeName=PK,AttributeType=S \
        AttributeName=SK,AttributeType=S \
        AttributeName=GSI1PK,AttributeType=S \
        AttributeName=GSI1SK,AttributeType=S \
        AttributeName=GSI2PK,AttributeType=S \
        AttributeName=GSI2SK,AttributeType=S \
    --key-schema \
        AttributeName=PK,KeyType=HASH \
        AttributeName=SK,KeyType=RANGE \
    --global-secondary-indexes '[
        {
            "IndexName": "GSI1",
            "KeySchema": [
                {"AttributeName": "GSI1PK", "KeyType": "HASH"},
                {"AttributeName": "GSI1SK", "KeyType": "RANGE"}
            ],
            "Projection": {"ProjectionType": "ALL"}
        },
        {
            "IndexName": "GSI2",
            "KeySchema": [
                {"AttributeName": "GSI2PK", "KeyType": "HASH"},
                {"AttributeName": "GSI2SK", "KeyType": "RANGE"}
            ],
            "Projection": {"ProjectionType": "ALL"}
        }
    ]' \
    --billing-mode PAY_PER_REQUEST \
    --stream-specification StreamEnabled=true,StreamViewType=NEW_AND_OLD_IMAGES \
    $ENDPOINT \
    --region $REGION

echo "Waiting for table to be active..."
aws dynamodb wait table-exists --table-name $TABLE_NAME $ENDPOINT --region $REGION

# TTL有効化
echo "Enabling TTL on TTL attribute..."
aws dynamodb update-time-to-live \
    --table-name $TABLE_NAME \
    --time-to-live-specification "Enabled=true,AttributeName=TTL" \
    $ENDPOINT \
    --region $REGION

echo "Table $TABLE_NAME created successfully!"
echo ""
echo "Table structure:"
echo "  Primary Key: PK (HASH), SK (RANGE)"
echo "  GSI1: GSI1PK (HASH), GSI1SK (RANGE)"
echo "  GSI2: GSI2PK (HASH), GSI2SK (RANGE)"
echo "  TTL: TTL attribute"
echo "  Streams: NEW_AND_OLD_IMAGES"
