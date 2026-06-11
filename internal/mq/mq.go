package mq

import "time"

// TicketMessage 票务消息
type TicketMessage struct {
	UserID       uint   `json:"user_id"`
	EventID      uint   `json:"event_id"`
	ShowID       uint   `json:"show_id,omitempty"`
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
	// Kafka Topics
	TicketOrderTopic     = "ticket.orders"
	TicketOrderDLQTopic  = "ticket.orders.dlq"
	TicketResultTopic    = "ticket.results"

	// Kafka Consumer Groups
	TicketOrderConsumerGroup = "ticket-order-processor"

	// Legacy Redis Stream keys (保留兼容，后续移除)
	TicketStreamKey     = "ticket:orders"
	TicketConsumerGroup = "ticket-processor"
)

// NewTicketMessage 创建票务消息
func NewTicketMessage(userID, eventID, ticketTypeID uint, quantity int) *TicketMessage {
	return &TicketMessage{
		UserID:       userID,
		EventID:      eventID,
		TicketTypeID: ticketTypeID,
		Quantity:     quantity,
		Timestamp:    time.Now().UnixNano(),
	}
}

// NewTicketMessageWithShow 创建带场次的票务消息
func NewTicketMessageWithShow(userID, eventID, showID, ticketTypeID uint, quantity int) *TicketMessage {
	return &TicketMessage{
		UserID:       userID,
		EventID:      eventID,
		ShowID:       showID,
		TicketTypeID: ticketTypeID,
		Quantity:     quantity,
		Timestamp:    time.Now().UnixNano(),
	}
}
