package events

import (
	"context"
	"sync"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type Publisher struct {
	mu  sync.RWMutex
	ctx context.Context
}

func NewPublisher() *Publisher {
	return &Publisher{}
}

func (p *Publisher) SetContext(ctx context.Context) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ctx = ctx
}

func (p *Publisher) Emit(eventName string, data interface{}) {
	p.mu.RLock()
	ctx := p.ctx
	p.mu.RUnlock()

	if ctx != nil {
		wailsRuntime.EventsEmit(ctx, eventName, data)
	}
}
