package payment

type CreateOrderRequest struct {
	CouponID  int64  `json:"coupon_id"`
	UserID    string `json:"user_id"`
	ReturnURL string `json:"return_url"`
	FailURL   string `json:"fail_url,omitempty"`
}

type CreateOrderResponse struct {
	OrderID    int64  `json:"order_id"`
	PaymentURL string `json:"payment_url"`
	Success    bool   `json:"success"`
	Message    string `json:"message,omitempty"`
}

type OrderStatusResponse struct {
	OrderID    int64   `json:"order_id"`
	Status     string  `json:"status"`
	CouponName string  `json:"coupon_name,omitempty"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
	Success    bool    `json:"success"`
	Message    string  `json:"message,omitempty"`
}

type AlfaBankRegisterRequest struct {
	OrderNumber        string `json:"orderNumber"`
	Amount             int64  `json:"amount"`
	Currency           string `json:"currency,omitempty"`
	ReturnUrl          string `json:"returnUrl"`
	FailUrl            string `json:"failUrl,omitempty"`
	Description        string `json:"description,omitempty"`
	Language           string `json:"language,omitempty"`
	ClientId           string `json:"clientId,omitempty"`
	JsonParams         string `json:"jsonParams,omitempty"`
	SessionTimeoutSecs int    `json:"sessionTimeoutSecs,omitempty"`
}

type AlfaBankRegisterResponse struct {
	OrderId      string `json:"orderId"`
	FormUrl      string `json:"formUrl"`
	ErrorCode    string `json:"errorCode,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type AlfaBankStatusResponse struct {
	ErrorCode             string `json:"errorCode"`
	ErrorMessage          string `json:"errorMessage,omitempty"`
	OrderNumber           string `json:"orderNumber"`
	OrderStatus           int    `json:"orderStatus"`
	ActionCode            int    `json:"actionCode"`
	ActionCodeDescription string `json:"actionCodeDescription"`
	Amount                int64  `json:"amount"`
	Currency              string `json:"currency"`
	Date                  int64  `json:"date"`
	Ip                    string `json:"ip"`
	OrderDescription      string `json:"orderDescription"`
}
