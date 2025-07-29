package payment

import (
	"github.com/uptrace/bun"
	"time"
)

// Статусы заказов
const (
	OrderStatusCreated   = "created"
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusFailed    = "failed"
	OrderStatusCancelled = "cancelled"
)

// Модель купона
type Coupon struct {
	bun.BaseModel `bun:"table:coupons"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	Name        string    `bun:"name,notnull" json:"name"`
	Description string    `bun:"description" json:"description"`
	Price       float64   `bun:"price,notnull" json:"price"`
	Currency    string    `bun:"currency,notnull,default:'RUB'" json:"currency"`
	IsActive    bool      `bun:"is_active,notnull,default:true" json:"is_active"`
	CreatedAt   time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`
}

// Модель заказа
type Order struct {
	bun.BaseModel `bun:"table:orders"`

	ID              int64     `bun:"id,pk,autoincrement" json:"id"`
	OrderNumber     string    `bun:"order_number,notnull,unique" json:"order_number"`
	AlfaBankOrderID string    `bun:"alfabank_order_id" json:"alfabank_order_id"`
	CouponID        int64     `bun:"coupon_id,notnull" json:"coupon_id"`
	UserID          string    `bun:"user_id,notnull" json:"user_id"`
	Amount          int64     `bun:"amount,notnull" json:"amount"` // в копейках
	Currency        string    `bun:"currency,notnull,default:'RUB'" json:"currency"`
	Status          string    `bun:"status,notnull,default:'created'" json:"status"`
	PaymentURL      string    `bun:"payment_url" json:"payment_url"`
	ReturnURL       string    `bun:"return_url" json:"return_url"`
	FailURL         string    `bun:"fail_url" json:"fail_url"`
	Description     string    `bun:"description" json:"description"`
	CreatedAt       time.Time `bun:"created_at,nullzero,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time `bun:"updated_at,nullzero,notnull,default:current_timestamp" json:"updated_at"`

	// Связи
	Coupon *Coupon `bun:"rel:belongs-to,join:coupon_id=id" json:"coupon,omitempty"`
}

// Модель активированных купонов пользователя
type UserCoupon struct {
	bun.BaseModel `bun:"table:user_coupons"`

	ID          int64     `bun:"id,pk,autoincrement" json:"id"`
	UserID      string    `bun:"user_id,notnull" json:"user_id"`
	CouponID    int64     `bun:"coupon_id,notnull" json:"coupon_id"`
	OrderID     int64     `bun:"order_id,notnull" json:"order_id"`
	ActivatedAt time.Time `bun:"activated_at,nullzero,notnull,default:current_timestamp" json:"activated_at"`
	IsUsed      bool      `bun:"is_used,notnull,default:false" json:"is_used"`
	UsedAt      time.Time `bun:"used_at,nullzero" json:"used_at,omitempty"`

	// Связи
	Coupon *Coupon `bun:"rel:belongs-to,join:coupon_id=id" json:"coupon,omitempty"`
	Order  *Order  `bun:"rel:belongs-to,join:order_id=id" json:"order,omitempty"`
}
