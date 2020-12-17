package pool

import (
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/sosodev/heart/config"
)

// Pool is a pool of *lua.State
type Pool struct {
	config      *config.Config
	stack       []*lua.State
	top         int
	lock        sync.Mutex
	initializer func(*lua.State) error
}

// New gets you a *Pool of fully initialized *lua.State
// Needs the initial size of the pool and an initializer function
// The initializer will be reused later when the pool grows to meet peak demand
func New(config *config.Config, initializer func(*lua.State) error) (*Pool, error) {
	pool := &Pool{
		config:      config,
		stack:       make([]*lua.State, config.InitialPoolSize),
		top:         config.InitialPoolSize - 1,
		initializer: initializer,
	}

	for i := 0; i < config.InitialPoolSize; i++ {
		state := lua.NewState()
		pool.stack[i] = state
		err := pool.initializer(state)
		if err != nil {
			return nil, err
		}
	}

	return pool, nil
}

func (p *Pool) empty() bool {
	return p.top == -1
}

func (p *Pool) peek() *lua.State {
	if p.empty() {
		return nil
	}

	return p.stack[p.top]
}

// Take a *lua.State from the pool
// Provisions and initializes a new one if the pool is empty
func (p *Pool) Take() (*lua.State, error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var state *lua.State
	if p.empty() {
		state = lua.NewState()
		err := p.initializer(state)
		if err != nil {
			return nil, err
		}
	} else {
		state = p.peek()
		p.stack = p.stack[:p.top]
		p.top--
	}

	return state, nil
}

// Return a *lua.State back to the pool
func (p *Pool) Return(state *lua.State) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.stack = append(p.stack, state)
	p.top++
}

// Cleanup the pool and all of its state
func (p *Pool) Cleanup() {
	for _, state := range p.stack {
		state.Close()
	}
}
