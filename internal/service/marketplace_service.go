package service

import (
	"fmt"

	"order-system/internal/pkg/db"
	"order-system/internal/repository"
)

type MarketplaceService struct {
	listingRepo   repository.MarketplaceRepository
	ticketRepo    repository.TicketRepository
	ticketTypeRepo repository.TicketTypeRepository
	userRepo      repository.UserRepository
}

func NewMarketplaceService(
	listingRepo repository.MarketplaceRepository,
	ticketRepo repository.TicketRepository,
	ticketTypeRepo repository.TicketTypeRepository,
	userRepo repository.UserRepository,
) *MarketplaceService {
	return &MarketplaceService{
		listingRepo:   listingRepo,
		ticketRepo:    ticketRepo,
		ticketTypeRepo: ticketTypeRepo,
		userRepo:      userRepo,
	}
}

type CreateListingInput struct {
	TicketID    uint
	Price       float64
	Description string
}

type ListingOutput struct {
	ID          uint    `json:"id"`
	TicketID    uint    `json:"ticket_id"`
	EventID     uint    `json:"event_id"`
	TicketName  string  `json:"ticket_name"`
	EventTitle  string  `json:"event_title"`
	SellerID    uint    `json:"seller_id"`
	SellerName  string  `json:"seller_name"`
	Price       float64 `json:"price"`
	Status      string  `json:"status"`
	BuyerID     uint    `json:"buyer_id,omitempty"`
	Description string  `json:"description"`
	CreatedAt   string  `json:"created_at"`
}

func (s *MarketplaceService) CreateListing(userID uint, input CreateListingInput) (*db.MarketplaceListing, error) {
	ticket, err := s.ticketRepo.FindByID(input.TicketID)
	if err != nil || ticket == nil {
		return nil, fmt.Errorf("票务不存在")
	}

	if ticket.UserID != userID {
		return nil, fmt.Errorf("无权出售此票务")
	}

	if ticket.Status != "paid" {
		return nil, fmt.Errorf("只有已支付的票务才能上架")
	}

	if input.Price <= 0 {
		return nil, fmt.Errorf("价格必须大于0")
	}

	existing, _ := s.listingRepo.FindByTicketID(input.TicketID)
	if existing != nil {
		return nil, fmt.Errorf("该票务已在转让市场中")
	}

	listing := &db.MarketplaceListing{
		TicketID:    input.TicketID,
		SellerID:    userID,
		Price:       input.Price,
		Status:      "active",
		Description: input.Description,
	}

	if err := s.listingRepo.Create(listing); err != nil {
		return nil, fmt.Errorf("上架失败: %w", err)
	}

	return listing, nil
}

func (s *MarketplaceService) BuyListing(buyerID, listingID uint) error {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return fmt.Errorf("商品不存在")
	}

	if listing.Status != "active" {
		return fmt.Errorf("该商品已下架或已售出")
	}

	if listing.SellerID == buyerID {
		return fmt.Errorf("不能购买自己的票")
	}

	ticket, err := s.ticketRepo.FindByID(listing.TicketID)
	if err != nil || ticket == nil {
		return fmt.Errorf("票务不存在")
	}

	if ticket.UserID != listing.SellerID {
		return fmt.Errorf("票务信息不一致")
	}

	// 转移票务所有权
	if err := s.ticketRepo.UpdateOwner(ticket.ID, buyerID); err != nil {
		return fmt.Errorf("转移票务失败: %w", err)
	}

	// 更新市场记录
	listing.Status = "sold"
	listing.BuyerID = buyerID
	if err := s.listingRepo.Update(listing); err != nil {
		return fmt.Errorf("更新市场记录失败: %w", err)
	}

	return nil
}

func (s *MarketplaceService) CancelListing(userID, listingID uint) error {
	listing, err := s.listingRepo.FindByID(listingID)
	if err != nil || listing == nil {
		return fmt.Errorf("商品不存在")
	}

	if listing.SellerID != userID {
		return fmt.Errorf("无权取消此商品")
	}

	if listing.Status != "active" {
		return fmt.Errorf("该商品已下架或已售出")
	}

	listing.Status = "cancelled"
	return s.listingRepo.Update(listing)
}

func (s *MarketplaceService) GetListing(id uint) (*ListingOutput, error) {
	listing, err := s.listingRepo.FindByID(id)
	if err != nil || listing == nil {
		return nil, fmt.Errorf("商品不存在")
	}

	ticket, err := s.ticketRepo.FindByID(listing.TicketID)
	if err != nil || ticket == nil {
		return nil, fmt.Errorf("票务不存在")
	}

	tt, _ := s.ticketTypeRepo.FindByID(ticket.TicketTypeID)
	ticketName := ""
	if tt != nil {
		ticketName = tt.Name
	}

	return &ListingOutput{
		ID:          listing.ID,
		TicketID:    listing.TicketID,
		EventID:     ticket.EventID,
		TicketName:  ticketName,
		SellerID:    listing.SellerID,
		Price:       listing.Price,
		Status:      listing.Status,
		BuyerID:     listing.BuyerID,
		Description: listing.Description,
		CreatedAt:   listing.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}

func (s *MarketplaceService) ListActiveListings(page, limit int) ([]ListingOutput, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	listings, total, err := s.listingRepo.FindActiveListings(page, limit)
	if err != nil {
		return nil, 0, err
	}

	return s.buildListingOutputs(listings), total, nil
}

func (s *MarketplaceService) ListByEvent(eventID uint, page, limit int) ([]ListingOutput, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}

	listings, total, err := s.listingRepo.FindActiveByEventID(eventID, page, limit)
	if err != nil {
		return nil, 0, err
	}

	return s.buildListingOutputs(listings), total, nil
}

func (s *MarketplaceService) ListMyListings(userID uint) ([]ListingOutput, error) {
	listings, err := s.listingRepo.FindBySellerID(userID)
	if err != nil {
		return nil, err
	}

	return s.buildListingOutputs(listings), nil
}

func (s *MarketplaceService) ListMyPurchases(userID uint) ([]ListingOutput, error) {
	listings, err := s.listingRepo.FindByBuyerID(userID)
	if err != nil {
		return nil, err
	}

	return s.buildListingOutputs(listings), nil
}

// buildListingOutputs 批量加载关联数据，避免 N+1 查询
func (s *MarketplaceService) buildListingOutputs(listings []db.MarketplaceListing) []ListingOutput {
	if len(listings) == 0 {
		return []ListingOutput{}
	}

	// 收集所有 ticket ID，批量查询
	ticketIDs := make([]uint, 0, len(listings))
	seen := make(map[uint]bool)
	for _, l := range listings {
		if !seen[l.TicketID] {
			ticketIDs = append(ticketIDs, l.TicketID)
			seen[l.TicketID] = true
		}
	}

	ticketMap := make(map[uint]*db.Ticket)
	for _, id := range ticketIDs {
		t, err := s.ticketRepo.FindByID(id)
		if err == nil && t != nil {
			ticketMap[id] = t
		}
	}

	// 收集所有 ticket type ID，批量查询
	ttIDs := make([]uint, 0)
	ttSeen := make(map[uint]bool)
	for _, t := range ticketMap {
		if !ttSeen[t.TicketTypeID] {
			ttIDs = append(ttIDs, t.TicketTypeID)
			ttSeen[t.TicketTypeID] = true
		}
	}

	ttMap := make(map[uint]string)
	tts, _ := s.ticketTypeRepo.FindByIDs(ttIDs)
	for _, tt := range tts {
		ttMap[tt.ID] = tt.Name
	}

	output := make([]ListingOutput, 0, len(listings))
	for _, listing := range listings {
		ticket, ok := ticketMap[listing.TicketID]
		if !ok {
			continue
		}

		output = append(output, ListingOutput{
			ID:          listing.ID,
			TicketID:    listing.TicketID,
			EventID:     ticket.EventID,
			TicketName:  ttMap[ticket.TicketTypeID],
			SellerID:    listing.SellerID,
			Price:       listing.Price,
			Status:      listing.Status,
			BuyerID:     listing.BuyerID,
			Description: listing.Description,
			CreatedAt:   listing.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	return output
}
