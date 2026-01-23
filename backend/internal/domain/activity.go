package domain

import "time"

type UserActivity struct {
	UserID     string            `json:"userId" dynamodbav:"UserId"`
	ActionType string            `json:"actionType" dynamodbav:"ActionType"` // VIEW, CLICK, ADD_CART, PURCHASE
	ProductID  string            `json:"productId" dynamodbav:"ProductId"`
	Metadata   map[string]string `json:"metadata" dynamodbav:"Metadata"`
	TTL        int64             `json:"-" dynamodbav:"TTL"` // Unix Epoch秒
	Timestamp  time.Time         `json:"timestamp" dynamodbav:"CreatedAt"`
}

type LogActivityRequest struct {
	ActionType string            `json:"actionType"`
	ProductID  string            `json:"productId"`
	Metadata   map[string]string `json:"metadata"`
}
