package pool

import (
	"fmt"
	"log"
	"math/rand"
	"sync"

	"github.com/aarzilli/golua/lua"
	"github.com/sosodev/heart/config"
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
	takeCount   map[*lua.State]int
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
		takeCount:   make(map[*lua.State]int),
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

	randIndex := rand.Intn(p.size())
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
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.empty() {
		state, err = p.newState()
		if err != nil {
			return nil, err
		}
	} else {
		state = p.randomTake()
	}

	p.takeCount[state] = p.takeCount[state] + 1

	return state, nil
}

// Return a *lua.State back to the pool
func (p *Pool) Return(state *lua.State) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if p.takeCount[state] > 10000 {
		go func() {
			state.Close()

			p.lock.Lock()
			defer p.lock.Unlock()

			delete(p.takeCount, state)

			nuState, err := p.newState()
			if err != nil {
				log.Printf("failed to allocate new state: %s\n", err)
			}

			p.stack = append(p.stack, nuState)
			p.top++
		}()

		return
	}

	p.stack = append(p.stack, state)
	p.top++
}

// Cleanup the pool and all of its state
func (p *Pool) Cleanup() {
	for _, state := range p.stack {
		state.Close()
	}
}
