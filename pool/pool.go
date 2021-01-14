package pool

import (
	"fmt"
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/rs/zerolog/log"
	"github.com/sosodev/heart/config"
	"github.com/sosodev/heart/las"
	"github.com/valyala/fastrand"
)

// TODO: consider ways to make the pool self-optimizing
// an ideal pool would be able to provision just enough state for peak demand
// without taking the performance penatly of doing it JIT

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

func (p *Pool) size() int {
	return p.top + 1
}

func (p *Pool) peek() *lua.State {
	if p.empty() {
		return nil
	}

	return p.stack[p.top]
}

func (p *Pool) randomTake() *lua.State {
	if p.empty() {
		panic("interally tried to take from empty pool")
	}

	var state *lua.State
	if p.size() == 1 {
		state = p.peek()
		p.stack = p.stack[:p.top]
		p.top--
		return state
	}

	randIndex := fastrand.Uint32n(uint32(p.size()))
	state = p.stack[randIndex]
	p.stack[randIndex] = p.stack[p.top]
	p.stack = p.stack[:p.top]
	p.top--

	return state
}

func (p *Pool) newState() (*lua.State, error) {
	state := lua.NewState()
	if state == nil {
		return nil, fmt.Errorf("failed to allocate new state -- LuaJIT probably OOM")
	}

	err := p.initializer(state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

// Take a *lua.State from the pool
// Provisions and initializes a new one if the pool is empty
func (p *Pool) Take() (state *lua.State, err error) {
	updateStateTakeCount := func() {
		as, ok := las.Get(state)
		if ok {
			as.IncrementTakeCount()
		}
	}

	p.lock.Lock()
	if !p.empty() {
		state = p.randomTake()
		p.lock.Unlock()
		updateStateTakeCount()
		return state, nil
	}
	p.lock.Unlock()

	state, err = p.newState()
	if err != nil {
		return nil, err
	}
	updateStateTakeCount()
	return state, nil
}

// Return a *lua.State back to the pool
func (p *Pool) Return(state *lua.State) {
	as, ok := las.Get(state)
	if !ok {
		log.Fatal().Msg("Failed to get associated state on pool return")
	}

	if as.GetTakeCount() > 10000 {
		go func() {
			state.Close()
			las.Free(state)

			nuState, err := p.newState()
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to allocate new state")
			}

			p.lock.Lock()
			p.stack = append(p.stack, nuState)
			p.top++
			p.lock.Unlock()
		}()

		return
	}

	p.lock.Lock()
	p.stack = append(p.stack, state)
	p.top++
	p.lock.Unlock()
}

// Cleanup the pool and all of its state
func (p *Pool) Cleanup() {
	for _, state := range p.stack {
		state.Close()
		las.Free(state)
	}
}
