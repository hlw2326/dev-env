package events

import (
	"context"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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

func (p *Publisher) Emit(event string, payload interface{}) {
	p.mu.RLock()
	ctx := p.ctx
	p.mu.RUnlock()
	if ctx == nil {
		return
	}
	defer func() {
		if recovered := recover(); recovered != nil {
			return
		}
	}()
	runtime.EventsEmit(ctx, event, payload)
}
