package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"order-system/internal/pkg/constants"
	"order-system/internal/pkg/db"
	"order-system/internal/pkg/redis/pkgredis"
	"order-system/internal/repository"
)

const ticketExpireDuration = 30 * time.Minute

func StartTicketExpireChecker(
	ctx context.Context,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	redis *pkgredis.RedisClientWrapper,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			checkExpiredTickets(ctx, ticketRepo, ticketTypeRepo, redis)
		}
	}
}

func checkExpiredTickets(
	ctx context.Context,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	redis *pkgredis.RedisClientWrapper,
) {
	deadline := time.Now().Add(-ticketExpireDuration)

	for {
		tickets, err := ticketRepo.FindExpiredReserved(deadline, 100)
		if err != nil {
			log.Printf("[TICKET EXPIRE] Failed to query expired tickets: %v", err)
			return
		}

		if len(tickets) == 0 {
			return
		}

		for i := range tickets {
			if err := expireTicket(ctx, &tickets[i], ticketRepo, ticketTypeRepo, redis); err != nil {
				log.Printf("[TICKET EXPIRE] Failed to expire ticket %d: %v", tickets[i].ID, err)
			} else {
				log.Printf("[TICKET EXPIRE] Ticket %d expired", tickets[i].ID)
			}
		}
	}
}

func expireTicket(
	ctx context.Context,
	ticket *db.Ticket,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	redis *pkgredis.RedisClientWrapper,
) error {
	if err := constants.IsValidTicketTransition(ticket.Status, constants.TicketStatusExpired); err != nil {
		return err
	}

	if err := ticketRepo.UpdateStatus(ticket.ID, constants.TicketStatusExpired); err != nil {
		return fmt.Errorf("failed to update ticket status: %w", err)
	}

	if err := ticketTypeRepo.UpdateStock(ticket.TicketTypeID, -ticket.Quantity); err != nil {
		return fmt.Errorf("failed to rollback stock: %w", err)
	}

	activityID := fmt.Sprintf("ticket:%d", ticket.EventID)
	redis.SeckillRollback(ctx, activityID,
		fmt.Sprint(ticket.TicketTypeID), fmt.Sprint(ticket.UserID))

	return nil
}
