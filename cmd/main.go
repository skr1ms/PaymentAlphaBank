package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	_ "github.com/lib/pq"
	"github.com/skr1ms/PaymentAlphaBank.git/config"
	"github.com/skr1ms/PaymentAlphaBank.git/internal/payment"
	"github.com/skr1ms/PaymentAlphaBank.git/migration"
	"github.com/skr1ms/PaymentAlphaBank.git/pkg/db"
)

func main() {
	config := config.NewTestConfig()

	db, err := db.NewDb(config)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	defer db.Close()

	migration.Init(db, config)

	app := fiber.New()
	app.Use(cors.New(
		cors.Config{
			AllowOrigins: "*",
			AllowHeaders: "Origin, Content-Type, Accept",
			AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
		},
	))
	app.Use(recover.New())

	api := app.Group("/api")

	// repository
	couponRepo := payment.NewCouponRepository(db.DB)
	orderRepo := payment.NewOrderRepository(db.DB)
	userCouponRepo := payment.NewUserCouponRepository(db.DB)

	// service
	alfaClient := payment.NewAlfaBankClient(config)
	couponService := payment.NewCouponService(couponRepo, orderRepo, userCouponRepo, alfaClient)

	// handler
	payment.NewPaymentHandler(api, &payment.PaymentHandlerDeps{
		CouponService: couponService,
	})

	log.Printf("Тестовая страница: http://localhost:%s/api/test", config.Port)
	log.Fatal(app.Listen(":" + config.Port))
}
