package payment

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/skr1ms/PaymentAlphaBank.git/config"
)

type AlfaBankClient struct {
	config *config.Config
	client *http.Client
}

func NewAlfaBankClient(config *config.Config) *AlfaBankClient {
	return &AlfaBankClient{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *AlfaBankClient) RegisterOrder(ctx context.Context, req *AlfaBankRegisterRequest) (*AlfaBankRegisterResponse, error) {
	data := url.Values{}
	data.Set("userName", c.config.Username)
	data.Set("password", c.config.Password)
	data.Set("orderNumber", req.OrderNumber)
	data.Set("amount", strconv.FormatInt(req.Amount, 10))
	data.Set("returnUrl", req.ReturnUrl)

	if req.Currency != "" {
		data.Set("currency", req.Currency)
	} else {
		data.Set("currency", "810") // По умолчанию рубли
	}

	if req.FailUrl != "" {
		data.Set("failUrl", req.FailUrl)
	}
	if req.Description != "" {
		data.Set("description", req.Description)
	}
	if req.Language != "" {
		data.Set("language", req.Language)
	} else {
		data.Set("language", "ru")
	}
	if req.ClientId != "" {
		data.Set("clientId", req.ClientId)
	}
	if req.JsonParams != "" {
		data.Set("jsonParams", req.JsonParams)
	}
	if req.SessionTimeoutSecs > 0 {
		data.Set("sessionTimeoutSecs", strconv.Itoa(req.SessionTimeoutSecs))
	}

	if c.config.IsTest {
		log.Printf("Отправляем запрос в Альфа-Банк: %s", data.Encode())
	}

	resp, err := c.client.PostForm(c.config.BaseURL+"/payment/rest/register.do", data)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса к API: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
	}

	if c.config.IsTest {
		log.Printf("Ответ от Альфа-Банка: %s", string(body))
	}

	var result AlfaBankRegisterResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга ответа: %w", err)
	}

	return &result, nil
}

func (c *AlfaBankClient) GetOrderStatus(ctx context.Context, orderID string) (*AlfaBankStatusResponse, error) {
	data := url.Values{}
	data.Set("userName", c.config.Username)
	data.Set("password", c.config.Password)
	data.Set("orderId", orderID)
	data.Set("language", "ru")

	resp, err := c.client.PostForm(c.config.BaseURL+"/payment/rest/getOrderStatus.do", data)
	if err != nil {
		return nil, fmt.Errorf("ошибка запроса статуса: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения ответа статуса: %w", err)
	}

	if c.config.IsTest {
		log.Printf("Статус заказа %s: %s", orderID, string(body))
	}

	var result AlfaBankStatusResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга статуса: %w", err)
	}

	return &result, nil
}

type CouponService struct {
	couponRepo     *CouponRepository
	orderRepo      *OrderRepository
	userCouponRepo *UserCouponRepository
	alfaClient     *AlfaBankClient
}

func NewCouponService(
	couponRepo *CouponRepository,
	orderRepo *OrderRepository,
	userCouponRepo *UserCouponRepository,
	alfaClient *AlfaBankClient,
) *CouponService {
	return &CouponService{
		couponRepo:     couponRepo,
		orderRepo:      orderRepo,
		userCouponRepo: userCouponRepo,
		alfaClient:     alfaClient,
	}
}

func (s *CouponService) GetCoupons(ctx context.Context) ([]Coupon, error) {
	return s.couponRepo.GetActiveCoupons(ctx)
}

func (s *CouponService) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	// Получаем купон
	coupon, err := s.couponRepo.GetByID(ctx, req.CouponID)
	if err != nil {
		return &CreateOrderResponse{
			Success: false,
			Message: "Купон не найден",
		}, err
	}

	// Генерируем уникальный номер заказа
	orderNumber := fmt.Sprintf("COUPON_%d_%s_%d", req.CouponID, req.UserID, time.Now().Unix())
	amountInKopecks := int64(coupon.Price * 100)

	// Создаем заказ в базе данных
	order := &Order{
		OrderNumber: orderNumber,
		CouponID:    req.CouponID,
		UserID:      req.UserID,
		Amount:      amountInKopecks,
		Currency:    coupon.Currency,
		Status:      OrderStatusCreated,
		ReturnURL:   req.ReturnURL,
		FailURL:     req.FailURL,
		Description: fmt.Sprintf("Покупка купона: %s", coupon.Name),
	}

	err = s.orderRepo.Create(ctx, order)
	if err != nil {
		return &CreateOrderResponse{
			Success: false,
			Message: "Ошибка создания заказа",
		}, err
	}

	// Регистрируем заказ в Альфа-Банке
	alfaReq := &AlfaBankRegisterRequest{
		OrderNumber:        orderNumber,
		Amount:             amountInKopecks,
		Currency:           "810", // Рубли
		ReturnUrl:          req.ReturnURL,
		FailUrl:            req.FailURL,
		Description:        order.Description,
		Language:           "ru",
		ClientId:           req.UserID,
		JsonParams:         fmt.Sprintf(`{"couponId":"%d","userId":"%s","orderId":"%d"}`, req.CouponID, req.UserID, order.ID),
		SessionTimeoutSecs: 1200,
	}

	alfaResp, err := s.alfaClient.RegisterOrder(ctx, alfaReq)
	if err != nil {
		// Обновляем статус заказа на failed
		s.orderRepo.UpdateStatus(ctx, order.ID, OrderStatusFailed)
		return &CreateOrderResponse{
			Success: false,
			Message: "Ошибка регистрации платежа",
		}, err
	}

	if alfaResp.ErrorCode != "" && alfaResp.ErrorCode != "0" {
		// Обновляем статус заказа на failed
		s.orderRepo.UpdateStatus(ctx, order.ID, OrderStatusFailed)
		return &CreateOrderResponse{
			Success: false,
			Message: alfaResp.ErrorMessage,
		}, fmt.Errorf("ошибка API Альфа-Банка: %s", alfaResp.ErrorMessage)
	}

	// Обновляем заказ с данными от Альфа-Банка
	err = s.orderRepo.UpdateAlfaBankOrderID(ctx, order.ID, alfaResp.OrderId)
	if err != nil {
		log.Printf("Ошибка обновления AlfaBankOrderID: %v", err)
	}

	// Обновляем статус на pending
	err = s.orderRepo.UpdateStatus(ctx, order.ID, OrderStatusPending)
	if err != nil {
		log.Printf("Ошибка обновления статуса заказа: %v", err)
	}

	return &CreateOrderResponse{
		OrderID:    order.ID,
		PaymentURL: alfaResp.FormUrl,
		Success:    true,
		Message:    "Заказ успешно создан",
	}, nil
}

func (s *CouponService) CheckOrderStatus(ctx context.Context, orderNumber string) (*OrderStatusResponse, error) {
	// Получаем заказ из базы данных
	order, err := s.orderRepo.GetByOrderNumber(ctx, orderNumber)
	if err != nil {
		return &OrderStatusResponse{
			Success: false,
			Message: "Заказ не найден",
		}, err
	}

	// Если у нас нет ID заказа от Альфа-Банка, возвращаем текущий статус
	if order.AlfaBankOrderID == "" {
		return &OrderStatusResponse{
			OrderID:  order.ID,
			Status:   order.Status,
			Amount:   float64(order.Amount) / 100,
			Currency: order.Currency,
			Success:  true,
		}, nil
	}

	// Проверяем статус в Альфа-Банке
	alfaStatus, err := s.alfaClient.GetOrderStatus(ctx, order.AlfaBankOrderID)
	if err != nil {
		return &OrderStatusResponse{
			OrderID:  order.ID,
			Status:   order.Status,
			Amount:   float64(order.Amount) / 100,
			Currency: order.Currency,
			Success:  true,
			Message:  "Ошибка проверки статуса в банке",
		}, nil
	}

	// Обновляем статус заказа в зависимости от ответа банка
	var newStatus string
	switch alfaStatus.OrderStatus {
	case 2: // Успешно оплачен
		newStatus = OrderStatusPaid
		// Активируем купон для пользователя
		err = s.userCouponRepo.ActivateCoupon(ctx, order.UserID, order.CouponID, order.ID)
		if err != nil {
			log.Printf("Ошибка активации купона: %v", err)
		}
	case 1: // В процессе оплаты
		newStatus = OrderStatusPending
	case 6: // Отклонен
		newStatus = OrderStatusFailed
	default:
		newStatus = order.Status
	}

	// Обновляем статус в базе данных, если он изменился
	if newStatus != order.Status {
		err = s.orderRepo.UpdateStatus(ctx, order.ID, newStatus)
		if err != nil {
			log.Printf("Ошибка обновления статуса заказа: %v", err)
		}
	}

	couponName := ""
	if order.Coupon != nil {
		couponName = order.Coupon.Name
	}

	return &OrderStatusResponse{
		OrderID:    order.ID,
		Status:     newStatus,
		CouponName: couponName,
		Amount:     float64(order.Amount) / 100,
		Currency:   order.Currency,
		Success:    true,
	}, nil
}

func (s *CouponService) GetUserCoupons(ctx context.Context, userID string) ([]UserCoupon, error) {
	return s.userCouponRepo.GetUserCoupons(ctx, userID)
}

func (s *CouponService) GetUserOrders(ctx context.Context, userID string) ([]Order, error) {
	return s.orderRepo.GetUserOrders(ctx, userID)
}
