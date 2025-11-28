package common

import (
	"context"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type debouncer struct {
	ctx    context.Context
	cancel context.CancelFunc
	signal chan struct{}
}

var (
	debounceMu sync.Mutex
	debouncers = map[string]*debouncer{}
)

// Debounce waits for the given duration before running cmd; newer calls with
// the same identifier cancel previous ones.
func Debounce(identifier string, duration time.Duration, cmd tea.Cmd) tea.Cmd {
	if cmd == nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	signal := make(chan struct{}, 1)
	state := &debouncer{ctx: ctx, cancel: cancel, signal: signal}

	debounceMu.Lock()
	if previous, ok := debouncers[identifier]; ok {
		previous.cancel()
	}
	debouncers[identifier] = state
	debounceMu.Unlock()

	go func(id string, current *debouncer) {
		timer := time.NewTimer(duration)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-current.ctx.Done():
			return
		}

		debounceMu.Lock()
		latest := debouncers[id]
		debounceMu.Unlock()
		if latest != current {
			return
		}

		select {
		case current.signal <- struct{}{}:
		default:
		}
	}(identifier, state)

	return func() tea.Msg {
		select {
		case <-state.signal:
			msg := cmd()

			debounceMu.Lock()
			latest := debouncers[identifier]
			debounceMu.Unlock()

			select {
			case <-state.ctx.Done():
				return nil
			default:
			}

			if latest != state {
				return nil
			}

			return msg
		case <-state.ctx.Done():
			return nil
		}
	}
}
