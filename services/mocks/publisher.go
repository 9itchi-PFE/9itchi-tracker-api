package mocks

import (
	"context"
	"sync"
)

// Publisher records published events; Err is returned from Publish when set.
type Publisher struct {
	mu     sync.Mutex
	Events []string
	Last   any
	Err    error
}

func (p *Publisher) Publish(_ context.Context, event string, payload any) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Events = append(p.Events, event)
	p.Last = payload
	return p.Err
}
