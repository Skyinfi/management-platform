package model

import "time"

type OperationLog struct {
	ID        string    `json:"id"`
	Operator  string    `json:"operator"`
	Action    string    `json:"action"`
	Target    string    `json:"target"`
	TargetType string   `json:"targetType"`
	Result    string    `json:"result"`
	Timestamp time.Time `json:"timestamp"`
}

type OperationLogResponse struct {
	Items []OperationLog `json:"items"`
}
