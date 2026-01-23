package domain

import "time"

type Order struct {
	ID              string      `json:"id" dynamodbav:"OrderId"`
	UserID          string      `json:"userId" dynamodbav:"UserId"`
	Status          string      `json:"status" dynamodbav:"Status"` // PENDING, CONFIRMED, SHIPPED, DELIVERED, CANCELLED
	TotalAmount     int         `json:"totalAmount" dynamodbav:"TotalAmount"`
	ShippingAddress *Address    `json:"shippingAddress" dynamodbav:"ShippingAddress"`
	Items           []OrderItem `json:"items,omitempty"`
	OrderDate       time.Time   `json:"orderDate" dynamodbav:"OrderDate"`
	UpdatedAt       time.Time   `json:"updatedAt" dynamodbav:"UpdatedAt"`
}

type OrderItem struct {
	ProductID   string `json:"productId" dynamodbav:"ProductId"`
	ProductName string `json:"productName" dynamodbav:"ProductName"`
	Price       int    `json:"price" dynamodbav:"Price"`
	Quantity    int    `json:"quantity" dynamodbav:"Quantity"`
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
