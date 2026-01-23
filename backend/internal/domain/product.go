package domain

import "time"

type Product struct {
	ID          string    `json:"id" dynamodbav:"ProductId"`
	Name        string    `json:"name" dynamodbav:"Name"`
	Description string    `json:"description" dynamodbav:"Description"`
	Price       int       `json:"price" dynamodbav:"Price"`
	Category    string    `json:"category" dynamodbav:"Category"`
	Stock       int       `json:"stock" dynamodbav:"Stock"`
	ImageURL    string    `json:"imageUrl" dynamodbav:"ImageUrl"`
	Version     int       `json:"version" dynamodbav:"Version"` // 楽観的ロック用
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"CreatedAt"`
	UpdatedAt   time.Time `json:"updatedAt" dynamodbav:"UpdatedAt"`
}

type CreateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Category    string `json:"category"`
	Stock       int    `json:"stock"`
	ImageURL    string `json:"imageUrl"`
}

type UpdateProductRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Category    string `json:"category"`
	ImageURL    string `json:"imageUrl"`
	Version     int    `json:"version"` // 楽観的ロック用
}

type PriceHistory struct {
	ProductID string    `json:"productId" dynamodbav:"ProductId"`
	Price     int       `json:"price" dynamodbav:"Price"`
	ChangedBy string    `json:"changedBy" dynamodbav:"ChangedBy"`
	Timestamp time.Time `json:"timestamp" dynamodbav:"CreatedAt"`
}

type InventoryLog struct {
	ProductID     string    `json:"productId" dynamodbav:"ProductId"`
	ChangeType    string    `json:"changeType" dynamodbav:"ChangeType"` // IN, OUT, ADJUST
	Quantity      int       `json:"quantity" dynamodbav:"Quantity"`
	PreviousStock int       `json:"previousStock" dynamodbav:"PreviousStock"`
	NewStock      int       `json:"newStock" dynamodbav:"NewStock"`
	Reason        string    `json:"reason" dynamodbav:"Reason"`
	OrderID       string    `json:"orderId,omitempty" dynamodbav:"OrderId,omitempty"`
	Timestamp     time.Time `json:"timestamp" dynamodbav:"CreatedAt"`
}
