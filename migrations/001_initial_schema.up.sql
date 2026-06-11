-- 初始数据库 Schema
-- 对应 internal/pkg/db/db.go 中的 AutoMigrate 模型

-- 用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    username VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) DEFAULT 'user'
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- 活动表
CREATE TABLE IF NOT EXISTS events (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255) NOT NULL,
    cover_image VARCHAR(500),
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    total_stock INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_events_deleted_at ON events(deleted_at);
CREATE INDEX IF NOT EXISTS idx_events_start_time ON events(start_time);
CREATE INDEX IF NOT EXISTS idx_events_end_time ON events(end_time);
CREATE INDEX IF NOT EXISTS idx_events_status ON events(status);

-- 票种表
CREATE TABLE IF NOT EXISTS ticket_types (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    event_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    max_per_user INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_ticket_types_deleted_at ON ticket_types(deleted_at);
CREATE INDEX IF NOT EXISTS idx_ticket_types_event_id ON ticket_types(event_id);

-- 票务表（订单）
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    user_id INTEGER NOT NULL,
    event_id INTEGER NOT NULL,
    show_id INTEGER,
    ticket_type_id INTEGER NOT NULL,
    order_no VARCHAR(255) UNIQUE NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 1,
    total_price DECIMAL(10,2),
    status VARCHAR(50) NOT NULL DEFAULT 'reserved',
    qr_code TEXT,
    discount_code VARCHAR(255),
    real_name VARCHAR(255),
    id_card VARCHAR(255),
    phone VARCHAR(255),
    transferred_to INTEGER,
    transfer_status VARCHAR(50) DEFAULT 'none'
);

CREATE INDEX IF NOT EXISTS idx_tickets_deleted_at ON tickets(deleted_at);
CREATE INDEX IF NOT EXISTS idx_tickets_user_id ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_event_id ON tickets(event_id);
CREATE INDEX IF NOT EXISTS idx_tickets_show_id ON tickets(show_id);
CREATE INDEX IF NOT EXISTS idx_tickets_ticket_type_id ON tickets(ticket_type_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);
CREATE INDEX IF NOT EXISTS idx_tickets_discount_code ON tickets(discount_code);
CREATE INDEX IF NOT EXISTS idx_tickets_real_name ON tickets(real_name);
CREATE INDEX IF NOT EXISTS idx_tickets_id_card ON tickets(id_card);
CREATE INDEX IF NOT EXISTS idx_tickets_phone ON tickets(phone);
CREATE INDEX IF NOT EXISTS idx_tickets_transferred_to ON tickets(transferred_to);

-- 复合索引
CREATE INDEX IF NOT EXISTS idx_tickets_user_status ON tickets(user_id, status);
CREATE INDEX IF NOT EXISTS idx_tickets_event_status ON tickets(event_id, status);
CREATE INDEX IF NOT EXISTS idx_tickets_status_created ON tickets(status, created_at);

-- 票务转让表
CREATE TABLE IF NOT EXISTS ticket_transfers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    ticket_id INTEGER NOT NULL,
    from_user_id INTEGER NOT NULL,
    to_user_id INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    transfer_type VARCHAR(50) NOT NULL DEFAULT 'gift',
    price DECIMAL(10,2),
    reason TEXT,
    reviewed_by INTEGER,
    reviewed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX IF NOT EXISTS idx_ticket_transfers_deleted_at ON ticket_transfers(deleted_at);
CREATE INDEX IF NOT EXISTS idx_ticket_transfers_ticket_id ON ticket_transfers(ticket_id);
CREATE INDEX IF NOT EXISTS idx_ticket_transfers_from_user_id ON ticket_transfers(from_user_id);
CREATE INDEX IF NOT EXISTS idx_ticket_transfers_to_user_id ON ticket_transfers(to_user_id);
CREATE INDEX IF NOT EXISTS idx_ticket_transfers_status ON ticket_transfers(status);
CREATE INDEX IF NOT EXISTS idx_ticket_transfers_transfer_type ON ticket_transfers(transfer_type);

-- 场次表
CREATE TABLE IF NOT EXISTS shows (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    event_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    show_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(50) DEFAULT 'draft' NOT NULL,
    stock INTEGER NOT NULL DEFAULT 0,
    sold_count INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_shows_deleted_at ON shows(deleted_at);
CREATE INDEX IF NOT EXISTS idx_shows_event_id ON shows(event_id);
CREATE INDEX IF NOT EXISTS idx_shows_show_time ON shows(show_time);
CREATE INDEX IF NOT EXISTS idx_shows_status ON shows(status);

-- 二手市场挂单表
CREATE TABLE IF NOT EXISTS marketplace_listings (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    ticket_id INTEGER NOT NULL,
    seller_id INTEGER NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    buyer_id INTEGER,
    description TEXT
);

CREATE INDEX IF NOT EXISTS idx_marketplace_listings_deleted_at ON marketplace_listings(deleted_at);
CREATE INDEX IF NOT EXISTS idx_marketplace_listings_ticket_id ON marketplace_listings(ticket_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_listings_seller_id ON marketplace_listings(seller_id);
CREATE INDEX IF NOT EXISTS idx_marketplace_listings_status ON marketplace_listings(status);

-- 复合索引
CREATE INDEX IF NOT EXISTS idx_marketplace_status_created ON marketplace_listings(status, created_at);

-- 促销码表
CREATE TABLE IF NOT EXISTS promo_codes (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    code VARCHAR(255) UNIQUE NOT NULL,
    event_id INTEGER,
    discount_type VARCHAR(50) NOT NULL,
    discount_value DECIMAL(10,2) NOT NULL,
    min_amount DECIMAL(10,2) DEFAULT 0,
    max_uses INTEGER DEFAULT 0,
    used_count INTEGER DEFAULT 0,
    start_time TIMESTAMP WITH TIME ZONE,
    end_time TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_promo_codes_deleted_at ON promo_codes(deleted_at);
CREATE INDEX IF NOT EXISTS idx_promo_codes_code ON promo_codes(code);
CREATE INDEX IF NOT EXISTS idx_promo_codes_event_id ON promo_codes(event_id);
