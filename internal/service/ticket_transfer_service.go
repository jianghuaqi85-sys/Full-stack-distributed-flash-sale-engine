package service

import (
	"fmt"
	"time"

	"order-system/internal/pkg/db"
	"order-system/internal/repository"
)

type TicketTransferService struct {
	ticketRepo       repository.TicketRepository
	transferRepo     repository.TicketTransferRepository
	userRepo         repository.UserRepository
}

func NewTicketTransferService(
	ticketRepo repository.TicketRepository,
	transferRepo repository.TicketTransferRepository,
	userRepo repository.UserRepository,
) *TicketTransferService {
	return &TicketTransferService{
		ticketRepo:   ticketRepo,
		transferRepo: transferRepo,
		userRepo:     userRepo,
	}
}

type RequestTransferInput struct {
	TicketID uint
	ToUserID uint
	Reason   string
}

func (s *TicketTransferService) RequestTransfer(userID uint, input RequestTransferInput) (*db.TicketTransfer, error) {
	// 检查票务是否存在
	ticket, err := s.ticketRepo.FindByID(input.TicketID)
	if err != nil || ticket == nil {
		return nil, fmt.Errorf("票务不存在")
	}

	// 检查是否是票务所有者
	if ticket.UserID != userID {
		return nil, fmt.Errorf("无权转让此票务")
	}

	// 检查票务状态是否可以转让
	if ticket.Status != "paid" {
		return nil, fmt.Errorf("只有已支付的票务才能转让")
	}

	// 检查是否已有待审核的转让请求
	existing, _ := s.transferRepo.FindByTicketID(input.TicketID)
	if existing != nil {
		return nil, fmt.Errorf("已有待审核的转让请求")
	}

	// 检查目标用户是否存在
	toUser, err := s.userRepo.FindByID(input.ToUserID)
	if err != nil || toUser == nil {
		return nil, fmt.Errorf("目标用户不存在")
	}

	// 创建转让请求
	transfer := &db.TicketTransfer{
		TicketID:   input.TicketID,
		FromUserID: userID,
		ToUserID:   input.ToUserID,
		Status:     "pending",
		Reason:     input.Reason,
	}

	if err := s.transferRepo.Create(transfer); err != nil {
		return nil, fmt.Errorf("创建转让请求失败: %w", err)
	}

	return transfer, nil
}

func (s *TicketTransferService) ApproveTransfer(transferID, adminID uint) error {
	transfer, err := s.transferRepo.FindByID(transferID)
	if err != nil || transfer == nil {
		return fmt.Errorf("转让请求不存在")
	}

	if transfer.Status != "pending" {
		return fmt.Errorf("该转让请求已处理")
	}

	// 更新转让状态
	now := time.Now()
	transfer.Status = "approved"
	transfer.ReviewedBy = adminID
	transfer.ReviewedAt = &now

	if err := s.transferRepo.Update(transfer); err != nil {
		return fmt.Errorf("更新转让状态失败: %w", err)
	}

	// 更新票务所有者
	ticket, err := s.ticketRepo.FindByID(transfer.TicketID)
	if err != nil || ticket == nil {
		return fmt.Errorf("票务不存在")
	}

	ticket.UserID = transfer.ToUserID
	ticket.TransferStatus = "approved"

	if err := s.ticketRepo.UpdateStatus(ticket.ID, ticket.Status); err != nil {
		return fmt.Errorf("更新票务失败: %w", err)
	}

	return nil
}

func (s *TicketTransferService) RejectTransfer(transferID, adminID uint, reason string) error {
	transfer, err := s.transferRepo.FindByID(transferID)
	if err != nil || transfer == nil {
		return fmt.Errorf("转让请求不存在")
	}

	if transfer.Status != "pending" {
		return fmt.Errorf("该转让请求已处理")
	}

	// 更新转让状态
	now := time.Now()
	transfer.Status = "rejected"
	transfer.ReviewedBy = adminID
	transfer.ReviewedAt = &now
	transfer.Reason = reason

	if err := s.transferRepo.Update(transfer); err != nil {
		return fmt.Errorf("更新转让状态失败: %w", err)
	}

	return nil
}

func (s *TicketTransferService) GetPendingTransfers() ([]db.TicketTransfer, error) {
	return s.transferRepo.FindPending()
}

func (s *TicketTransferService) GetTransferHistory(userID uint) ([]db.TicketTransfer, error) {
	return s.transferRepo.FindByUserID(userID)
}

func (s *TicketTransferService) DirectGift(userID uint, input RequestTransferInput) (*db.TicketTransfer, error) {
	ticket, err := s.ticketRepo.FindByID(input.TicketID)
	if err != nil || ticket == nil {
		return nil, fmt.Errorf("票务不存在")
	}

	if ticket.UserID != userID {
		return nil, fmt.Errorf("无权转让此票务")
	}

	if ticket.Status != "paid" {
		return nil, fmt.Errorf("只有已支付的票务才能转让")
	}

	toUser, err := s.userRepo.FindByID(input.ToUserID)
	if err != nil || toUser == nil {
		return nil, fmt.Errorf("目标用户不存在")
	}

	if input.ToUserID == userID {
		return nil, fmt.Errorf("不能转让给自己")
	}

	transfer := &db.TicketTransfer{
		TicketID:     input.TicketID,
		FromUserID:   userID,
		ToUserID:     input.ToUserID,
		Status:       "approved",
		TransferType: "gift",
		Reason:       input.Reason,
	}

	if err := s.transferRepo.Create(transfer); err != nil {
		return nil, fmt.Errorf("创建转让记录失败: %w", err)
	}

	ticket.UserID = input.ToUserID
	ticket.TransferStatus = "approved"
	ticket.TransferredTo = input.ToUserID
	if err := s.ticketRepo.UpdateStatus(ticket.ID, ticket.Status); err != nil {
		return nil, fmt.Errorf("更新票务失败: %w", err)
	}

	return transfer, nil
}
