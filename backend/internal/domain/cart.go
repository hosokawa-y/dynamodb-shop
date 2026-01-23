package domain

import "time"

type CartItem struct {
	UserID      string    `json:"userId" dynamodbav:"UserId"`
	ProductID   string    `json:"productId" dynamodbav:"ProductId"`
	ProductName string    `json:"productName" dynamodbav:"ProductName"`
	Price       int       `json:"price" dynamodbav:"Price"`
	Quantity    int       `json:"quantity" dynamodbav:"Quantity"`
	AddedAt     time.Time `json:"addedAt" dynamodbav:"AddedAt"`
}

type AddToCartRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type UpdateCartRequest struct {
	Quantity int `json:"quantity"`
}

type Cart struct {
	Items      []CartItem `json:"items"`
	TotalPrice int        `json:"totalPrice"`
	ItemCount  int        `json:"itemCount"`
}
