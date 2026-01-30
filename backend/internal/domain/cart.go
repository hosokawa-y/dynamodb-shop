package domain

import "time"

type CartItem struct {
	UserID      string    `json:"userId" dynamodbav:"UserId"`
	ProductID   string    `json:"productId" dynamodbav:"ProductId"`
	ProductName string    `json:"productName" dynamodbav:"ProductName"`
	Price       int       `json:"price" dynamodbav:"Price"`
	Quantity    int       `json:"quantity" dynamodbav:"Quantity"`
	Version     int       `json:"version" dynamodbav:"Version"` // 楽観的ロック用
	AddedAt     time.Time `json:"addedAt" dynamodbav:"AddedAt"`
	UpdatedAt   time.Time `json:"updatedAt" dynamodbav:"UpdatedAt"`
}

type AddToCartRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity"`
	Version  int `json:"version"` // 楽観的ロック用
}

type Cart struct {
	Items      []CartItem `json:"items"`
	TotalPrice int        `json:"totalPrice"`
	ItemCount  int        `json:"itemCount"`
}
