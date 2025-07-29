package migration

import (
	"context"
	"fmt"
	"log"

	"github.com/skr1ms/PaymentAlphaBank.git/config"
	"github.com/skr1ms/PaymentAlphaBank.git/internal/payment"
	"github.com/skr1ms/PaymentAlphaBank.git/pkg/db"
	"github.com/uptrace/bun"
)

func Init(db *db.Db, config *config.Config) {
	// Создаем таблицы и индексы
	ctx := context.Background()
	if err := payment.CreateTables(ctx, db.DB); err != nil {
		log.Fatalf("Ошибка создания таблиц: %v", err)
		return
	}

	if err := payment.CreateIndexes(ctx, db.DB); err != nil {
		log.Fatalf("Ошибка создания индексов: %v", err)
		return
	}

	// Создаем тестовые данные
	if config.IsTest {
		if err := createTestData(ctx, db.DB); err != nil {
			log.Printf("Ошибка создания тестовых данных: %v", err)
			return
		}
	}
}

func createTestData(ctx context.Context, db *bun.DB) error {
	// Проверяем, есть ли уже купоны
	count, err := db.NewSelect().Model((*payment.Coupon)(nil)).Count(ctx)
	if err != nil {
		return fmt.Errorf("ошибка проверки количества купонов: %w", err)
	}

	if count > 0 {
		log.Println("Тестовые данные уже существуют")
		return nil
	}

	// Создаем тестовые купоны
	testCoupons := []payment.Coupon{
		{
			Name:        "Скидка 10%",
			Description: "Скидка 10% на любую покупку в магазине",
			Price:       100.00,
			Currency:    "RUB",
			IsActive:    true,
		},
		{
			Name:        "Скидка 20%",
			Description: "Скидка 20% на любую покупку в магазине",
			Price:       200.00,
			Currency:    "RUB",
			IsActive:    true,
		},
		{
			Name:        "Бесплатная доставка",
			Description: "Бесплатная доставка для заказов от 500 рублей",
			Price:       50.00,
			Currency:    "RUB",
			IsActive:    true,
		},
		{
			Name:        "Скидка 50%",
			Description: "Скидка 50% на выбранные товары",
			Price:       500.00,
			Currency:    "RUB",
			IsActive:    true,
		},
	}

	_, err = db.NewInsert().Model(&testCoupons).Exec(ctx)
	if err != nil {
		return fmt.Errorf("ошибка создания тестовых купонов: %w", err)
	}

	log.Printf("Создано %d тестовых купонов", len(testCoupons))
	return nil
}
