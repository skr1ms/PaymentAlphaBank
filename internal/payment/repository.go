package payment

import (
    "context"
    "fmt"
    "time"

    "github.com/uptrace/bun"
)

type CouponRepository struct {
    db *bun.DB
}

func NewCouponRepository(db *bun.DB) *CouponRepository {
    return &CouponRepository{db: db}
}

func (r *CouponRepository) GetActiveCoupons(ctx context.Context) ([]Coupon, error) {
    var coupons []Coupon
    err := r.db.NewSelect().
        Model(&coupons).
        Where("is_active = ?", true).
        Order("created_at DESC").
        Scan(ctx)
    return coupons, err
}

func (r *CouponRepository) GetByID(ctx context.Context, id int64) (*Coupon, error) {
    coupon := &Coupon{}
    err := r.db.NewSelect().
        Model(coupon).
        Where("id = ? AND is_active = ?", id, true).
        Scan(ctx)
    if err != nil {
        return nil, err
    }
    return coupon, nil
}

type OrderRepository struct {
    db *bun.DB
}

func NewOrderRepository(db *bun.DB) *OrderRepository {
    return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *Order) error {
    order.CreatedAt = time.Now()
    order.UpdatedAt = time.Now()
    _, err := r.db.NewInsert().Model(order).Exec(ctx)
    return err
}

func (r *OrderRepository) GetByOrderNumber(ctx context.Context, orderNumber string) (*Order, error) {
    order := &Order{}
    err := r.db.NewSelect().
        Model(order).
        Relation("Coupon").
        Where("order_number = ?", orderNumber).
        Scan(ctx)
    return order, err
}

func (r *OrderRepository) GetByAlfaBankOrderID(ctx context.Context, alfaBankOrderID string) (*Order, error) {
    order := &Order{}
    err := r.db.NewSelect().
        Model(order).
        Relation("Coupon").
        Where("alfabank_order_id = ?", alfaBankOrderID).
        Scan(ctx)
    return order, err
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, orderID int64, status string) error {
    _, err := r.db.NewUpdate().
        Model((*Order)(nil)).
        Set("status = ?", status).
        Set("updated_at = ?", time.Now()).
        Where("id = ?", orderID).
        Exec(ctx)
    return err
}

func (r *OrderRepository) UpdateAlfaBankOrderID(ctx context.Context, orderID int64, alfaBankOrderID string) error {
    _, err := r.db.NewUpdate().
        Model((*Order)(nil)).
        Set("alfabank_order_id = ?", alfaBankOrderID).
        Set("updated_at = ?", time.Now()).
        Where("id = ?", orderID).
        Exec(ctx)
    return err
}

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID string) ([]Order, error) {
    var orders []Order
    err := r.db.NewSelect().
        Model(&orders).
        Relation("Coupon").
        Where("user_id = ?", userID).
        Order("created_at DESC").
        Scan(ctx)
    return orders, err
}

type UserCouponRepository struct {
    db *bun.DB
}

func NewUserCouponRepository(db *bun.DB) *UserCouponRepository {
    return &UserCouponRepository{db: db}
}

func (r *UserCouponRepository) ActivateCoupon(ctx context.Context, userID string, couponID, orderID int64) error {
    userCoupon := &UserCoupon{
        UserID:      userID,
        CouponID:    couponID,
        OrderID:     orderID,
        ActivatedAt: time.Now(),
        IsUsed:      false,
    }
    
    _, err := r.db.NewInsert().Model(userCoupon).Exec(ctx)
    return err
}

func (r *UserCouponRepository) GetUserCoupons(ctx context.Context, userID string) ([]UserCoupon, error) {
    var userCoupons []UserCoupon
    err := r.db.NewSelect().
        Model(&userCoupons).
        Relation("Coupon").
        Relation("Order").
        Where("user_id = ?", userID).
        Order("activated_at DESC").
        Scan(ctx)
    return userCoupons, err
}

func (r *UserCouponRepository) UseCoupon(ctx context.Context, userCouponID int64) error {
    _, err := r.db.NewUpdate().
        Model((*UserCoupon)(nil)).
        Set("is_used = ?", true).
        Set("used_at = ?", time.Now()).
        Where("id = ?", userCouponID).
        Exec(ctx)
    return err
}

// Создание таблиц
func CreateTables(ctx context.Context, db *bun.DB) error {
    models := []interface{}{
        (*Coupon)(nil),
        (*Order)(nil),
        (*UserCoupon)(nil),
    }
    
    for _, model := range models {
        _, err := db.NewCreateTable().Model(model).IfNotExists().Exec(ctx)
        if err != nil {
            return fmt.Errorf("ошибка создания таблицы: %w", err)
        }
    }
    
    return nil
}

// Создание индексов
func CreateIndexes(ctx context.Context, db *bun.DB) error {
    indexes := []string{
        "CREATE INDEX IF NOT EXISTS idx_orders_order_number ON orders(order_number)",
        "CREATE INDEX IF NOT EXISTS idx_orders_alfabank_order_id ON orders(alfabank_order_id)",
        "CREATE INDEX IF NOT EXISTS idx_orders_user_id ON orders(user_id)",
        "CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status)",
        "CREATE INDEX IF NOT EXISTS idx_user_coupons_user_id ON user_coupons(user_id)",
        "CREATE INDEX IF NOT EXISTS idx_user_coupons_coupon_id ON user_coupons(coupon_id)",
        "CREATE INDEX IF NOT EXISTS idx_coupons_is_active ON coupons(is_active)",
    }
    
    for _, indexSQL := range indexes {
        _, err := db.ExecContext(ctx, indexSQL)
        if err != nil {
            return fmt.Errorf("ошибка создания индекса: %w", err)
        }
    }
    
    return nil
}