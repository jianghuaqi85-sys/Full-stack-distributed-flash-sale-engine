package constants

import "fmt"

const (
	EventStatusDraft   = "draft"
	EventStatusOnSale  = "on_sale"
	EventStatusOffSale = "off_sale"
	EventStatusEnded   = "ended"
)

var validEventStatuses = map[string]bool{
	EventStatusDraft:   true,
	EventStatusOnSale:  true,
	EventStatusOffSale: true,
	EventStatusEnded:   true,
}

var eventAllowedTransitions = map[string]map[string]bool{
	EventStatusDraft: {
		EventStatusOnSale: true,
	},
	EventStatusOnSale: {
		EventStatusOffSale: true,
		EventStatusEnded:   true,
	},
	EventStatusOffSale: {
		EventStatusOnSale: true,
		EventStatusEnded:  true,
	},
	EventStatusEnded: {},
}

func IsValidEventStatus(status string) bool {
	return validEventStatuses[status]
}

func IsValidEventTransition(from, to string) error {
	if !validEventStatuses[to] {
		return fmt.Errorf("invalid event status: %s", to)
	}

	transitions, ok := eventAllowedTransitions[from]
	if !ok {
		return fmt.Errorf("invalid current event status: %s", from)
	}

	if !transitions[to] {
		return fmt.Errorf("cannot transition event from %s to %s", from, to)
	}

	return nil
}
