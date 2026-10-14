package task

import "time"

type Status string

const (
	StatusQueued     Status = "queued"
	StatusRunning    Status = "running"
	StatusCancelling Status = "cancelling"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type Task struct {
	ID           int64      `json:"id"`
	TaskType     string     `json:"taskType"`
	Status       Status     `json:"status"`
	Progress     int        `json:"progress"`
	Message      string     `json:"message"`
	ErrorCode    *string    `json:"errorCode"`
	ErrorMessage *string    `json:"errorMessage"`
	PayloadJSON  string     `json:"payloadJson"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	FinishedAt   *time.Time `json:"finishedAt"`
}

type Accepted struct {
	TaskID int64  `json:"taskId"`
	Status Status `json:"status"`
}
