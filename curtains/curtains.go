package curtains

type CurtainState string

const (
	CurtainStateOpening CurtainState = "opening"
	CurtainStateClosing              = "closing"
	CurtainStateStopped              = "stopped"
)

type Curtain interface {
	// Initialises the curtain controller. If any data is written to the error channel
	// returned by this function it should be assumed that the curtain is no longer in
	// operation and should be re-initialized.
	// An error written to this channel means the curtain controller is no longer
	// connected. A nil value written to this channel means the controller has shut down
	// (as requested by the Shutdown() function)
	Init() <-chan error

	// Move the curtain to target position (0-100). 0 = fully open, 100 = fully closed
	SetTargetPosition(int)

	// Ask the curtain controller to provide its state and position. These will be
	// communicated via the channels returned by Position and State when the data
	// becomes available.
	Query()

	// Whenever the curtain controller provides an update to its current position
	// it will be written into this channel.
	Position() <-chan int

	// Whenever the curtain controller provides an update to its current state
	// it will be written into this channel.
	State() <-chan CurtainState

	// Requests the controller connection be gracefully terminated. A nil value
	// will be written to the error channel returned by Init() when shutdown has
	// completed.
	Shutdown()
}
