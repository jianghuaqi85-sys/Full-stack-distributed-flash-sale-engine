package constants

import "fmt"

const (
	TicketStatusReserved  = "reserved"
	TicketStatusPaid      = "paid"
	TicketStatusUsed      = "used"
	TicketStatusExpired   = "expired"
	TicketStatusCancelled = "cancelled"
)

var validTicketStatuses = map[string]bool{
	TicketStatusReserved:  true,
	TicketStatusPaid:      true,
	TicketStatusUsed:      true,
	TicketStatusExpired:   true,
	TicketStatusCancelled: true,
}

var ticketAllowedTransitions = map[string]map[string]bool{
	TicketStatusReserved: {
		TicketStatusPaid:      true,
		TicketStatusExpired:   true,
		TicketStatusCancelled: true,
	},
	TicketStatusPaid: {
		TicketStatusUsed:      true,
		TicketStatusCancelled: true,
	},
	TicketStatusUsed:      {},
	TicketStatusExpired:   {},
	TicketStatusCancelled: {},
}

func IsValidTicketStatus(status string) bool {
	return validTicketStatuses[status]
}

func IsValidTicketTransition(from, to string) error {
	if !validTicketStatuses[to] {
		return fmt.Errorf("invalid ticket status: %s", to)
	}

	transitions, ok := ticketAllowedTransitions[from]
	if !ok {
		return fmt.Errorf("invalid current ticket status: %s", from)
	}

	if !transitions[to] {
		return fmt.Errorf("cannot transition ticket from %s to %s", from, to)
	}

	return nil
}
