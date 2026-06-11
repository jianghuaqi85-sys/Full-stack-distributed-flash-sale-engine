package db

import (
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique;not null" json:"username"`
	Password string `gorm:"not null" json:"-"`
	Email    string `gorm:"unique;not null" json:"email"`
	Role     string `gorm:"default:user" json:"role"`
}

type Event struct {
	gorm.Model
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	Location    string    `gorm:"not null"`
	CoverImage  string
	StartTime   time.Time `gorm:"not null;index"`
	EndTime     time.Time `gorm:"not null;index"`
	Status      string    `gorm:"default:draft;not null;index"`
	TotalStock  int       `gorm:"default:0"`
}

type TicketType struct {
	gorm.Model
	EventID    uint    `gorm:"not null;index"`
	Name       string  `gorm:"not null"`
	Price      float64 `gorm:"not null"`
	Stock      int     `gorm:"not null;default:0"`
	MaxPerUser int     `gorm:"not null;default:1"`
	SortOrder  int     `gorm:"default:0"`
}

type Ticket struct {
	gorm.Model
	UserID       uint    `gorm:"not null;index"`
	EventID      uint    `gorm:"not null;index"`
	ShowID       uint    `gorm:"index"`
	TicketTypeID uint    `gorm:"not null;index"`
	OrderNo      string  `gorm:"uniqueIndex;not null"`
	Quantity     int     `gorm:"not null;default:1"`
	TotalPrice   float64
	Status       string  `gorm:"not null;default:reserved;index"`
	QRCode       string  `gorm:"type:text"`
	DiscountCode string  `gorm:"index"`
	// 实名制字段
	RealName      string `gorm:"index"`
	IDCard        string `gorm:"index"`
	Phone         string `gorm:"index"`
	TransferredTo uint   `gorm:"index"`
	TransferStatus string `gorm:"default:none"` // none, pending, approved, rejected
}

type TicketTransfer struct {
	gorm.Model
	TicketID     uint   `gorm:"not null;index"`
	FromUserID   uint   `gorm:"not null;index"`
	ToUserID     uint   `gorm:"not null;index"`
	Status       string `gorm:"not null;default:pending;index"` // pending, approved, rejected
	TransferType string `gorm:"not null;default:gift;index"`     // gift(转赠), marketplace(二手交易)
	Price        float64
	Reason       string
	ReviewedBy   uint
	ReviewedAt   *time.Time
}

type Show struct {
	gorm.Model
	EventID    uint      `gorm:"not null;index"`
	Name       string    `gorm:"not null"`
	ShowTime   time.Time `gorm:"not null;index"`
	EndTime    time.Time `gorm:"not null"`
	Status     string    `gorm:"default:draft;not null;index"` // draft, on_sale, off_sale, ended
	Stock      int       `gorm:"not null;default:0"`
	SoldCount  int       `gorm:"default:0"`
	SortOrder  int       `gorm:"default:0"`
}

type MarketplaceListing struct {
	gorm.Model
	TicketID    uint    `gorm:"not null;index"`
	SellerID    uint    `gorm:"not null;index"`
	Price       float64 `gorm:"not null"`
	Status      string  `gorm:"not null;default:active;index"` // active, sold, cancelled
	BuyerID     uint
	Description string  `gorm:"type:text"`
}

type PromoCode struct {
	gorm.Model
	Code         string    `gorm:"uniqueIndex;not null"`
	EventID      uint      `gorm:"index"`
	DiscountType string    `gorm:"not null"` // percent, fixed
	DiscountValue float64  `gorm:"not null"`
	MinAmount    float64   `gorm:"default:0"`
	MaxUses      int       `gorm:"default:0"` // 0 = unlimited
	UsedCount    int       `gorm:"default:0"`
	StartTime    time.Time
	EndTime      time.Time
	IsActive     bool      `gorm:"default:true"`
}

func NewConnection(dsn string, maxOpenConns, maxIdleConns, connMaxLifetime int) (*gorm.DB, error) {
	return NewConnectionWithIdleTime(dsn, maxOpenConns, maxIdleConns, connMaxLifetime, 0)
}

// NewConnectionWithIdleTime 创建数据库连接，支持 ConnMaxIdleTime 配置
func NewConnectionWithIdleTime(dsn string, maxOpenConns, maxIdleConns, connMaxLifetime, connMaxIdleTime int) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(time.Duration(connMaxLifetime) * time.Second)
	if connMaxIdleTime > 0 {
		sqlDB.SetConnMaxIdleTime(time.Duration(connMaxIdleTime) * time.Second)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}, &Event{}, &TicketType{}, &Ticket{}, &PromoCode{}, &TicketTransfer{}, &Show{}, &MarketplaceListing{}); err != nil {
		return err
	}

	compositeIndexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_tickets_user_status ON tickets(user_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_event_status ON tickets(event_id, status)",
		"CREATE INDEX IF NOT EXISTS idx_tickets_status_created ON tickets(status, created_at)",
		"CREATE INDEX IF NOT EXISTS idx_marketplace_status_created ON marketplace_listings(status, created_at)",
	}
	for _, idx := range compositeIndexes {
		if err := db.Exec(idx).Error; err != nil {
			return err
		}
	}

	return nil
}
