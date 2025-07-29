package payment

import (
	"log"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type PaymentHandlerDeps struct {
	CouponService *CouponService
}

type PaymentHandler struct {
	fiber.Router
	deps *PaymentHandlerDeps
}

func NewPaymentHandler(router fiber.Router, deps *PaymentHandlerDeps) {

	handler := &PaymentHandler{
		Router: router,
		deps:   deps,
	}

	// API маршруты
	router.Get("/coupons", handler.GetCoupons)
	router.Post("/orders", handler.CreateOrder)
	router.Get("/orders/{orderNumber}/status", handler.GetOrderStatus)
	router.Get("/users/{userID}/coupons", handler.GetUserCoupons)
	router.Get("/users/{userID}/orders", handler.GetUserOrders)

	// Платежные маршруты
	router.Get("/payment/return", handler.PaymentReturn)
	router.Post("/payment/notification", handler.PaymentNotification)

	// Тестовые маршруты
	router.Get("/test", handler.TestPage)
	router.Get("/test/payment", handler.TestPayment)

	// Статические файлы
	router.Static("/", "./static/")

}

func (h *PaymentHandler) GetCoupons(c *fiber.Ctx) error {
	coupons, err := h.deps.CouponService.GetCoupons(c.Context())
	if err != nil {
		log.Printf("Ошибка получения купонов: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка получения купонов",
		})
	}

	return c.JSON(coupons)
}

func (h *PaymentHandler) CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный формат запроса",
		})
	}

	response, err := h.deps.CouponService.CreateOrder(c.Context(), &req)
	if err != nil {
		log.Printf("Ошибка создания заказа: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка создания заказа",
		})
	}

	return c.JSON(response)
}

func (h *PaymentHandler) GetOrderStatus(c *fiber.Ctx) error {
	orderNumber := c.Params("orderNumber")

	if orderNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не указан номер заказа",
		})
	}

	response, err := h.deps.CouponService.CheckOrderStatus(c.Context(), orderNumber)
	if err != nil {
		log.Printf("Ошибка проверки статуса заказа: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка проверки статуса заказа",
		})
	}

	return c.JSON(response)
}

func (h *PaymentHandler) GetUserCoupons(c *fiber.Ctx) error {
	userID := c.Params("userID")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не указан ID пользователя",
		})
	}

	coupons, err := h.deps.CouponService.GetUserCoupons(c.Context(), userID)
	if err != nil {
		log.Printf("Ошибка получения купонов пользователя: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка получения купонов",
		})
	}

	return c.JSON(coupons)
}

func (h *PaymentHandler) GetUserOrders(c *fiber.Ctx) error {
	userID := c.Params("userID")

	if userID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не указан ID пользователя",
		})
	}

	orders, err := h.deps.CouponService.GetUserOrders(c.Context(), userID)
	if err != nil {
		log.Printf("Ошибка получения заказов пользователя: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка получения заказов",
		})
	}

	return c.JSON(orders)
}

func (h *PaymentHandler) PaymentReturn(c *fiber.Ctx) error {
	orderNumber := c.Query("orderNumber")
	if orderNumber == "" {
		// Пробуем получить из orderId
		orderId := c.Query("orderId")
		if orderId != "" {
			// Можно добавить логику поиска по alfaBankOrderId
			log.Printf("Возврат с платежной страницы для orderId: %s", orderId)
		}
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Не указан номер заказа",
		})
	}

	status, err := h.deps.CouponService.CheckOrderStatus(c.Context(), orderNumber)
	if err != nil {
		log.Printf("Ошибка проверки статуса при возврате: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Ошибка проверки статуса платежа",
		})
	}

	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Результат платежа</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; text-align: center; }
        .success { color: green; }
        .error { color: red; }
        .pending { color: orange; }
    </style>
</head>
<body>
    <h1>Результат платежа</h1>
`

	switch status.Status {
	case OrderStatusPaid:
		html += `<div class="success">
            <h2>✓ Платеж успешно завершен!</h2>
            <p>Купон "` + status.CouponName + `" активирован в вашем аккаунте.</p>
        </div>`
	case OrderStatusFailed:
		html += `<div class="error">
            <h2>✗ Платеж отклонен</h2>
            <p>Попробуйте еще раз или выберите другой способ оплаты.</p>
        </div>`
	case OrderStatusPending:
		html += `<div class="pending">
            <h2>⏳ Платеж обрабатывается</h2>
            <p>Статус платежа будет обновлен в ближайшее время.</p>
        </div>`
	default:
		html += `<div class="error">
            <h2>? Неизвестный статус платежа</h2>
        </div>`
	}

	html += `
    <p><a href="/">Вернуться на главную</a></p>
</body>
</html>`

	return c.SendString(html)
}

func (h *PaymentHandler) PaymentNotification(c *fiber.Ctx) error {
	// Webhook от Альфа-Банка
	orderNumber := c.FormValue("orderNumber")
	orderId := c.FormValue("orderId")

	log.Printf("Получено уведомление о платеже: orderNumber=%s, orderId=%s", orderNumber, orderId)

	if orderNumber != "" {
		_, err := h.deps.CouponService.CheckOrderStatus(c.Context(), orderNumber)
		if err != nil {
			log.Printf("Ошибка обработки уведомления: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Error",
			})
		}
	}

	return c.SendString("OK")
}

func (h *PaymentHandler) TestPage(c *fiber.Ctx) error {
	html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Тест системы купонов</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .section { margin: 20px 0; padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; margin: 5px; }
        button:hover { background: #0056b3; }
        .result { margin: 10px 0; padding: 10px; background: #f8f9fa; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Тест системы купонов</h1>
    
    <div class="section">
        <h2>Доступные купоны</h2>
        <button onclick="loadCoupons()">Загрузить купоны</button>
        <div id="coupons-result" class="result"></div>
    </div>
    
    <div class="section">
        <h2>Создать тестовый заказ</h2>
        <button onclick="createTestOrder()">Создать заказ</button>
        <div id="order-result" class="result"></div>
    </div>

    <script>
        function loadCoupons() {
            fetch('/api/coupons')
                .then(response => response.json())
                .then(data => {
                    const div = document.getElementById('coupons-result');
                    div.innerHTML = '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
                })
                .catch(error => {
                    document.getElementById('coupons-result').innerHTML = 'Ошибка: ' + error.message;
                });
        }

        function createTestOrder() {
            const orderData = {
                coupon_id: 1,
                user_id: 'test_user_123',
                return_url: window.location.origin + '/payment/return',
                fail_url: window.location.origin + '/payment/return'
            };

            fetch('/api/orders', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(orderData)
            })
            .then(response => response.json())
            .then(data => {
                const div = document.getElementById('order-result');
                div.innerHTML = '<pre>' + JSON.stringify(data, null, 2) + '</pre>';
                if (data.success && data.payment_url) {
                    div.innerHTML += '<br><a href="' + data.payment_url + '" target="_blank">Перейти к оплате</a>';
                }
            })
            .catch(error => {
                document.getElementById('order-result').innerHTML = 'Ошибка: ' + error.message;
            });
        }
    </script>
</body>
</html>`

	return c.Type("html").SendString(html)
}

func (h *PaymentHandler) TestPayment(c *fiber.Ctx) error {
	if c.Method() == "GET" {
		// Показываем форму
		html := `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>Тест платежа</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .form-group { margin: 10px 0; }
        label { display: block; margin-bottom: 5px; }
        input { width: 300px; padding: 5px; }
        button { background: #007bff; color: white; padding: 10px 20px; border: none; border-radius: 5px; cursor: pointer; }
        .result { margin-top: 20px; padding: 10px; background: #f8f9fa; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>Тест платежа</h1>
    <form method="POST">
        <div class="form-group">
            <label>ID купона:</label>
            <input type="number" name="coupon_id" value="1" required>
        </div>
        <div class="form-group">
            <label>ID пользователя:</label>
            <input type="text" name="user_id" value="test_user_123" required>
        </div>
        <button type="submit">Создать платеж</button>
    </form>
</body>
</html>`
		return c.SendString(html)
	}

	// Обрабатываем POST
	couponIDStr := c.FormValue("coupon_id")
	userID := c.FormValue("user_id")

	couponID, err := strconv.ParseInt(couponIDStr, 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Неверный ID купона",
		})
	}

	req := &CreateOrderRequest{
		CouponID:  couponID,
		UserID:    userID,
		ReturnURL: "http://" + c.Hostname() + "/payment/return",
		FailURL:   "http://" + c.Hostname() + "/payment/return",
	}

	response, err := h.deps.CouponService.CreateOrder(c.Context(), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	if response.Success {
		return c.Redirect(response.PaymentURL, fiber.StatusSeeOther)
	} else {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": response.Message,
		})
	}
}
