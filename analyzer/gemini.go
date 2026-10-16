package analyzer

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"
)

// DefaultGeminiModel เป็นชื่อย่อที่ Google ชี้ไปยังโมเดล Flash รุ่นล่าสุดให้เอง
// ถ้าใช้ไม่ได้ ให้ตั้ง environment variable GEMINI_MODEL เป็นชื่อรุ่นอื่น
const DefaultGeminiModel = "gemini-flash-latest"

// DefaultGeminiFallback ใช้เมื่อโมเดลหลักคนใช้เยอะจนตอบ 503
const DefaultGeminiFallback = "gemini-flash-lite-latest"

type Gemini struct {
	APIKey   string
	Model    string
	Fallback string
	BaseURL  string // เว้นว่าง = ใช้ของ Google
	Client   *http.Client
}

func NewGemini(apiKey, model string) *Gemini {
	if model == "" {
		model = DefaultGeminiModel
	}
	return &Gemini{
		APIKey:   apiKey,
		Model:    model,
		Fallback: DefaultGeminiFallback,
		Client:   &http.Client{Timeout: 2 * time.Minute},
	}
}

func (g *Gemini) Analyze(ctx context.Context, path string) (Result, error) {
	if !IsSupported(path) {
		return Result{}, fmt.Errorf("%w (%s)", ErrUnsupported, filepath.Ext(path))
	}

	img, err := prepareImage(path, 768)
	if err != nil {
		return Result{}, fmt.Errorf("read image: %w", err)
	}

	body, _ := json.Marshal(map[string]any{
		"contents": []any{map[string]any{
			"parts": []any{
				map[string]any{"text": prompt + ` Reply only with JSON like {"tags":["dog","beach"]}.`},
				map[string]any{"inline_data": map[string]any{
					"mime_type": "image/jpeg",
					"data":      base64.StdEncoding.EncodeToString(img),
				}},
			},
		}},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
		},
	})

	// ลองโมเดลหลักก่อน ถ้าเซิร์ฟเวอร์ Google ไม่ว่าง ค่อยสลับไปตัวสำรอง
	models := []string{g.Model}
	if g.Fallback != "" && g.Fallback != g.Model {
		models = append(models, g.Fallback)
	}

	for i, model := range models {
		if i > 0 {
			fmt.Printf("  (switch to fallback model %s)\n", model)
		}
		text, status, err := g.callWithRetry(ctx, model, body)
		if err == nil {
			return parseTags(text)
		}
		if !isTemporary(status) || ctx.Err() != nil {
			return Result{}, err
		}
		if i == len(models)-1 {
			return Result{}, err
		}
	}
	return Result{}, errors.New("unreachable")
}

// isTemporary: error ที่รอแล้วลองใหม่มีโอกาสหาย
// 429 = เกินโควตาต่อนาที, 500/502/503/504 = เซิร์ฟเวอร์ Google ไม่ว่าง/ขัดข้องชั่วคราว
func isTemporary(status int) bool {
	switch status {
	case 429, 500, 502, 503, 504:
		return true
	}
	return false
}

// callWithRetry ลองใหม่สูงสุด 2 ครั้ง โดยรอนานขึ้นเรื่อย ๆ (5s, 10s)
func (g *Gemini) callWithRetry(ctx context.Context, model string, body []byte) (string, int, error) {
	base := g.BaseURL
	if base == "" {
		base = "https://generativelanguage.googleapis.com"
	}
	url := base + "/v1beta/models/" + model + ":generateContent"
	wait := 5 * time.Second

	for attempt := 0; ; attempt++ {
		text, status, err := g.call(ctx, url, body)
		if err == nil || !isTemporary(status) || attempt == 2 {
			return text, status, err
		}

		fmt.Printf("  (gemini %d, retry in %v...)\n", status, wait)
		select {
		case <-time.After(wait):
			wait *= 2
		case <-ctx.Done():
			return "", status, ctx.Err()
		}
	}
}

func (g *Gemini) call(ctx context.Context, url string, body []byte) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", g.APIKey)

	resp, err := g.Client.Do(req)
	if err != nil {
		return "", 0, fmt.Errorf("call gemini (ต่อเน็ตอยู่ไหม?): %w", err)
	}
	defer resp.Body.Close()

	raw, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		hint := ""
		switch resp.StatusCode {
		case 400, 401, 403:
			hint = " → เช็ก GEMINI_API_KEY ว่าถูกต้อง"
		case 404:
			hint = " → ไม่พบชื่อโมเดลนี้ ลองตั้ง GEMINI_MODEL เป็นรุ่นอื่น"
		case 429:
			hint = " → ใช้เกินโควตาฟรี รอสักครู่แล้วลองใหม่"
		case 500, 502, 503, 504:
			hint = " → เซิร์ฟเวอร์ Google ไม่ว่างชั่วคราว ลองใหม่ภายหลัง"
		}
		msg := string(raw)
		if len(msg) > 300 {
			msg = msg[:300] + "..."
		}
		return "", resp.StatusCode, fmt.Errorf("gemini status %d%s: %s", resp.StatusCode, hint, msg)
	}

	var out struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text    string `json:"text"`
					Thought bool   `json:"thought"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.Unmarshal(raw, &out); err != nil {
		return "", resp.StatusCode, fmt.Errorf("decode gemini response: %w", err)
	}
	if len(out.Candidates) == 0 {
		return "", resp.StatusCode, errors.New("gemini returned no result (รูปอาจถูกบล็อกโดย safety filter)")
	}

	var sb strings.Builder
	for _, p := range out.Candidates[0].Content.Parts {
		if !p.Thought {
			sb.WriteString(p.Text)
		}
	}
	return sb.String(), resp.StatusCode, nil
}

// parseTags ดึง JSON {"tags":[...]} ออกจากข้อความที่โมเดลตอบ
// (เผื่อกรณีโมเดลห่อด้วย ```json ... ``` มา)
func parseTags(text string) (Result, error) {
	start := strings.Index(text, "{")
	end := strings.LastIndex(text, "}")
	if start < 0 || end <= start {
		return Result{}, fmt.Errorf("model did not return JSON: %q", text)
	}

	var r Result
	if err := json.Unmarshal([]byte(text[start:end+1]), &r); err != nil {
		return Result{}, fmt.Errorf("model did not return valid JSON: %w", err)
	}

	r.Tags = NormalizeTags(r.Tags, 10)
	if len(r.Tags) == 0 {
		return Result{}, errors.New("model returned no tags")
	}
	return r, nil
}
