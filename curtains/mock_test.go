package curtains

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockStartAndStop(t *testing.T) {
	t.Parallel()
	c := NewMockCurtain()
	ec := c.Init()
	c.Shutdown()
	err := <-ec
	require.Nil(t, err)
}

func TestInitIdempotency(t *testing.T) {
	t.Parallel()
	c := NewMockCurtain()
	c.Init()
	ec := c.Init()
	c.Shutdown()
	err := <-ec
	require.Nil(t, err)
}

var timeoutExceededError = errors.New("Timeout exceeded")

func waitForPosition(t *testing.T, c Curtain, targetPos int, timeout time.Duration) ([]int, error) {
	t.Helper()
	positions := make([]int, 0)
	for {
		select {
		case newPos := <-c.Position():
			positions = append(positions, newPos)
			if newPos == targetPos {
				return positions, nil
			}
		case <-time.Tick(timeout):
			return positions, timeoutExceededError
		}
	}
}

func waitForState(t *testing.T, c Curtain, targetState CurtainState, timeout time.Duration) ([]CurtainState, error) {
	t.Helper()
	states := make([]CurtainState, 0)
	for {
		select {
		case newState := <-c.State():
			states = append(states, newState)
			if newState == targetState {
				return states, nil
			}
		case <-time.Tick(timeout):
			return states, timeoutExceededError
		}
	}
}

func TestMockMoveToPosition(t *testing.T) {
	t.Parallel()
	mockTickerTime = time.Millisecond
	c := NewMockCurtain()
	errC := c.Init()
	c.SetTargetPosition(100)

	seenPositions, err := waitForPosition(t, c, 100, time.Second)
	assert.NoError(t, err)

	seenStates, err := waitForState(t, c, CurtainStateStopped, time.Second)
	assert.NoError(t, err)

	expectPositions := make([]int, 100)
	for i := range expectPositions {
		expectPositions[i] = i + 1
	}

	assert.Equal(t, expectPositions, seenPositions)
	assert.Equal(t, []CurtainState{CurtainStateClosing, CurtainStateStopped}, seenStates)

	c.Shutdown()
	<-errC
}

func TestQuery(t *testing.T) {
	t.Parallel()
	c := NewMockCurtain()
	errC := c.Init()
	c.Query()

	seenPositions, err := waitForPosition(t, c, 0, time.Second)
	assert.NoError(t, err)

	seenStates, err := waitForState(t, c, CurtainStateStopped, time.Second)
	assert.NoError(t, err)

	assert.Equal(t, []int{0}, seenPositions)
	assert.Equal(t, []CurtainState{CurtainStateStopped}, seenStates)

	c.Shutdown()
	<-errC
}
