package intents

type AddMessage struct {
	Text      string
	Err       error
	NoTimeout bool
}

func (AddMessage) isIntent() {}

type DismissOldest struct{}

func (DismissOldest) isIntent() {}
