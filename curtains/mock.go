package curtains

import (
	"sync"
	"time"
)

var mockTickerTime = time.Millisecond * 250

type mockCurtain struct {
	targetPos       int
	currentPos      int
	state           CurtainState
	posC            chan int
	stateC          chan CurtainState
	requestExitChan chan struct{}
	confirmExitChan chan error
	shutdown        bool
	m               *sync.Mutex
}

func NewMockCurtain() Curtain {
	return &mockCurtain{
		state: CurtainStateStopped,
	}
}

func (c *mockCurtain) setup() {
	c.m = &sync.Mutex{}
	c.posC = make(chan int)
	c.stateC = make(chan CurtainState)
	c.requestExitChan = make(chan struct{})
	c.confirmExitChan = make(chan error)
}

func (c *mockCurtain) Init() <-chan error {
	if c.m != nil {
		return c.confirmExitChan
	}
	c.setup()
	ticker := time.NewTicker(mockTickerTime)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.tick()
			case <-c.requestExitChan:
				c.m.Lock()
				defer c.m.Unlock()
				ticker.Stop()
				c.shutdown = true
				defer close(c.confirmExitChan)
				defer close(c.posC)
				defer close(c.stateC)
				return
			}
		}
	}()
	return c.confirmExitChan
}

func (c *mockCurtain) SetTargetPosition(p int) {
	c.m.Lock()
	defer c.m.Unlock()
	c.targetPos = p
}

func (c *mockCurtain) Query() {
	c.informState(c.state)
	c.informPos(c.currentPos)
}

func (c *mockCurtain) Position() <-chan int {
	return c.posC
}

func (c *mockCurtain) State() <-chan CurtainState {
	return c.stateC
}

func (c *mockCurtain) Shutdown() {
	go func() {
		close(c.requestExitChan)
	}()
}

func (c *mockCurtain) updateState(v CurtainState) {
	if c.state == v {
		return
	}
	c.state = v
	c.informState(v)
}

func (c *mockCurtain) informState(v CurtainState) {
	go func() {
		select {
		case c.stateC <- v:
		case <-c.requestExitChan:
		}
	}()
}

func (c *mockCurtain) updatePos(v int) {
	if c.currentPos == v {
		return
	}
	c.currentPos = v
	c.informPos(v)
}

func (c *mockCurtain) informPos(v int) {
	go func() {
		select {
		case c.posC <- v:
		case <-c.requestExitChan:
		}
	}()
}

func (c *mockCurtain) tick() {
	c.m.Lock()
	defer c.m.Unlock()
	switch {
	case c.targetPos > c.currentPos:
		c.updateState(CurtainStateClosing)
		c.updatePos(c.currentPos + 1)
	case c.targetPos < c.currentPos:
		c.updateState(CurtainStateOpening)
		c.updatePos(c.currentPos - 1)
	default:
		c.updateState(CurtainStateStopped)
	}
}
