package domain

import "time"

// Order は注文ヘッダー
// 【キー設計】
//
//	PK: USER#<userId>
//	SK: ORDER#<orderId>
type Order struct {
	ID          string      `json:"id" dynamodbav:"OrderId"`
	UserID      string      `json:"userId" dynamodbav:"UserId"`
	Status      string      `json:"status" dynamodbav:"Status"` // PENDING, CONFIRMED, SHIPPED, DELIVERED, CANCELLED
	TotalAmount int         `json:"totalAmount" dynamodbav:"TotalAmount"`
	ItemCount   int         `json:"itemCount" dynamodbav:"ItemCount"`
	Items       []OrderItem `json:"items,omitempty"`
	CreatedAt   time.Time   `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time   `json:"updatedAt" dynamodbav:"UpdatedAt"`
}

// OrderItem は注文明細
// 【キー設計】
//
//	PK: ORDER#<orderId>
//	SK: ITEM#<productId>
type OrderItem struct {
	OrderID     string `json:"orderId" dynamodbav:"OrderId"`
	ProductID   string `json:"productId" dynamodbav:"ProductId"`
	ProductName string `json:"productName" dynamodbav:"ProductName"`
	Price       int    `json:"price" dynamodbav:"Price"` // 注文時の価格（スナップショット）
	Quantity    int    `json:"quantity" dynamodbav:"Quantity"`
	Subtotal    int    `json:"subtotal" dynamodbav:"Subtotal"` // Price * Quantity
}

type Address struct {
	ZipCode    string `json:"zipCode" dynamodbav:"ZipCode"`
	Prefecture string `json:"prefecture" dynamodbav:"Prefecture"`
	City       string `json:"city" dynamodbav:"City"`
	Address    string `json:"address" dynamodbav:"Address"`
}

type CreateOrderRequest struct {
	ShippingAddress *Address `json:"shippingAddress"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

const (
	OrderStatusPending   = "PENDING"
	OrderStatusConfirmed = "CONFIRMED"
	OrderStatusShipped   = "SHIPPED"
	OrderStatusDelivered = "DELIVERED"
	OrderStatusCancelled = "CANCELLED"
)
