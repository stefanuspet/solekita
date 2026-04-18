package payment

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/stefanuspet/solekita/backend/internal/config"
	"github.com/stefanuspet/solekita/backend/internal/service"
)

type TripayClientImpl struct {
	apiKey       string
	privateKey   string
	merchantCode string
	baseURL      string
	http         *http.Client
}

func NewTripay(cfg *config.Config) *TripayClientImpl {
	return &TripayClientImpl{
		apiKey:       cfg.TripayAPIKey,
		privateKey:   cfg.TripayPrivateKey,
		merchantCode: cfg.TripayMerchantCode,
		baseURL:      cfg.TripayBaseURL,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// ── CreateTransaction ─────────────────────────────────────────────────────────

type tripayOrderItem struct {
	Name     string `json:"name"`
	Price    int    `json:"price"`
	Quantity int    `json:"quantity"`
}

type tripayCreateRequest struct {
	Method        string            `json:"method"`
	MerchantRef   string            `json:"merchant_ref"`
	Amount        int               `json:"amount"`
	CustomerName  string            `json:"customer_name"`
	CustomerEmail string            `json:"customer_email,omitempty"`
	CustomerPhone string            `json:"customer_phone"`
	OrderItems    []tripayOrderItem `json:"order_items"`
	Signature     string            `json:"signature"`
	ExpiredTime   int64             `json:"expired_time"`
	ReturnURL     string            `json:"return_url,omitempty"`
}

type tripayCreateResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    struct {
		Reference  string `json:"reference"`
		PaymentURL string `json:"pay_url"`
	} `json:"data"`
}

// CreateTransaction membuat tagihan ke Tripay.
// Signature: HMAC-SHA256(merchantCode + merchantRef + amount, privateKey).
func (t *TripayClientImpl) CreateTransaction(ctx context.Context, req service.TripayTransactionRequest) (*service.TripayTransactionResult, error) {
	sig := t.signTransaction(req.MerchantRef, req.Amount)

	items := make([]tripayOrderItem, len(req.OrderItems))
	for i, it := range req.OrderItems {
		items[i] = tripayOrderItem{Name: it.Name, Price: it.Price, Quantity: it.Quantity}
	}

	payload := tripayCreateRequest{
		Method:        "QRIS",
		MerchantRef:   req.MerchantRef,
		Amount:        req.Amount,
		CustomerName:  req.CustomerName,
		CustomerEmail: req.CustomerEmail,
		CustomerPhone: req.CustomerPhone,
		OrderItems:    items,
		Signature:     sig,
		ExpiredTime:   req.ExpiredTime,
		ReturnURL:     req.ReturnURL,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("TripayClient.CreateTransaction: marshal: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, t.baseURL+"/transaction/create", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("TripayClient.CreateTransaction: buat request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+t.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := t.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("TripayClient.CreateTransaction: kirim request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("TripayClient.CreateTransaction: baca response: %w", err)
	}

	var tripayResp tripayCreateResponse
	if err := json.Unmarshal(respBody, &tripayResp); err != nil {
		return nil, fmt.Errorf("TripayClient.CreateTransaction: parse response: %w", err)
	}

	if !tripayResp.Success {
		return nil, fmt.Errorf("TripayClient.CreateTransaction: %s", tripayResp.Message)
	}

	return &service.TripayTransactionResult{
		Reference:  tripayResp.Data.Reference,
		PaymentURL: tripayResp.Data.PaymentURL,
	}, nil
}

// ── ValidateSignature ─────────────────────────────────────────────────────────

// ValidateSignature memvalidasi HMAC-SHA256 signature dari webhook Tripay.
// Tripay mengirim signature di header X-Callback-Signature.
// Dihitung dari: HMAC-SHA256(raw request body, privateKey).
func (t *TripayClientImpl) ValidateSignature(body []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(t.privateKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// ── helpers ───────────────────────────────────────────────────────────────────

// signTransaction membuat signature untuk create transaction.
// Format: HMAC-SHA256(merchantCode + merchantRef + amount, privateKey).
func (t *TripayClientImpl) signTransaction(merchantRef string, amount int) string {
	data := fmt.Sprintf("%s%s%d", t.merchantCode, merchantRef, amount)
	mac := hmac.New(sha256.New, []byte(t.privateKey))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
