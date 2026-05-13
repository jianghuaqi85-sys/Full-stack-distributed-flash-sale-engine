package mq

import "time"

// TicketMessage 票务消息
type TicketMessage struct {
	UserID       uint   `json:"user_id"`
	EventID      uint   `json:"event_id"`
	TicketTypeID uint   `json:"ticket_type_id"`
	Quantity     int    `json:"quantity"`
	Timestamp    int64  `json:"timestamp"`
}

// TicketResult 票务处理结果
type TicketResult struct {
	TicketID  uint   `json:"ticket_id"`
	UserID    uint   `json:"user_id"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	OrderNo   string `json:"order_no"`
	Timestamp int64  `json:"timestamp"`
}

const (
	TicketStreamKey     = "ticket:orders"
	TicketConsumerGroup = "ticket-processor"
)

func NewTicketMessage(userID, eventID, ticketTypeID uint, quantity int) *TicketMessage {
	return &TicketMessage{
		UserID:       userID,
		EventID:      eventID,
		TicketTypeID: ticketTypeID,
		Quantity:     quantity,
		Timestamp:    time.Now().UnixNano(),
	}
}
