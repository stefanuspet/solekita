package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/stefanuspet/solekita/backend/internal/config"
)

const fonnteBaseURL = "https://api.fonnte.com/send"

type FonnteClient struct {
	token  string
	sender string
	http   *http.Client
}

func NewFonnte(cfg *config.Config) *FonnteClient {
	return &FonnteClient{
		token:  cfg.FonnteToken,
		sender: cfg.FonnteSender,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type fonnteResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Detail  string `json:"detail"`
}

// Send mengirim pesan WhatsApp ke satu nomor via Fonnte.
// phone: nomor tujuan format internasional tanpa '+' (contoh: "6281234567890").
// message: isi pesan.
func (f *FonnteClient) Send(ctx context.Context, phone, message string) error {
	if f.token == "" {
		return fmt.Errorf("FonnteClient.Send: token belum dikonfigurasi")
	}

	form := url.Values{}
	form.Set("target", phone)
	form.Set("message", message)
	if f.sender != "" {
		form.Set("sender", f.sender)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fonnteBaseURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("FonnteClient.Send: buat request: %w", err)
	}
	req.Header.Set("Authorization", f.token)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := f.http.Do(req)
	if err != nil {
		return fmt.Errorf("FonnteClient.Send: kirim request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("FonnteClient.Send: baca response: %w", err)
	}

	var result fonnteResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("FonnteClient.Send: parse response: %w", err)
	}

	if !result.Status {
		return fmt.Errorf("FonnteClient.Send: gagal kirim WA ke %s: %s", phone, result.Detail)
	}

	return nil
}
