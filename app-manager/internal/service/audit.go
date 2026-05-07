package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/Skyinfi/management-platform/app-manager/internal/model"
)

type AuditLog struct {
	mu   sync.RWMutex
	logs []model.OperationLog
}

func NewAuditLog() *AuditLog {
	return &AuditLog{
		logs: make([]model.OperationLog, 0),
	}
}

func (a *AuditLog) Record(operator, action, target, targetType, result string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	entry := model.OperationLog{
		ID:         fmt.Sprintf("%d", time.Now().UnixNano()),
		Operator:   operator,
		Action:     action,
		Target:     target,
		TargetType: targetType,
		Result:     result,
		Timestamp:  time.Now(),
	}

	a.logs = append(a.logs, entry)
	if len(a.logs) > 1000 {
		a.logs = a.logs[len(a.logs)-1000:]
	}
}

func (a *AuditLog) List(limit, offset int) []model.OperationLog {
	a.mu.RLock()
	defer a.mu.RUnlock()

	total := len(a.logs)
	if offset >= total {
		return nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	result := make([]model.OperationLog, end-offset)
	copy(result, a.logs[offset:end])

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}
