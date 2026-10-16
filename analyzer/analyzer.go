package analyzer

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	_ "golang.org/x/image/webp" // ให้เปิดไฟล์ .webp ได้
)

// Result คือผลวิเคราะห์ 1 รูป (ตกลงรูปแบบนี้กับคนที่ 2)
type Result struct {
	Tags []string `json:"tags"`
}

// Analyzer คือสิ่งที่วิเคราะห์รูปได้ (ตอนนี้มี Gemini ตัวเดียว)
type Analyzer interface {
	Analyze(ctx context.Context, path string) (Result, error)
}

var ErrUnsupported = errors.New("unsupported image format")

var supported = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true,
	".gif": true, ".bmp": true, ".tif": true, ".tiff": true,
	".jfif": true, // Windows มักบันทึกรูป JPEG จากเว็บเป็น .jfif
	".webp": true, // รูปที่โหลดจากเว็บ
}

func IsSupported(path string) bool {
	return supported[strings.ToLower(filepath.Ext(path))]
}

// prompt ที่ส่งให้ AI พร้อมรูป
const prompt = "Give 5 to 10 tags for this photo. " +
	"Each tag is a short lowercase singular English noun (e.g. dog, beach, car, food, sunset) " +
	"about the main objects, place, and scene. No sentences."

// prepareImage ย่อรูปให้ด้านยาวไม่เกิน maxSide แล้วแปลงเป็น JPEG
func prepareImage(path string, maxSide int) ([]byte, error) {
	img, err := imaging.Open(path, imaging.AutoOrientation(true))
	if err != nil {
		return nil, err
	}
	img = imaging.Fit(img, maxSide, maxSide, imaging.Lanczos)

	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(85)); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// NormalizeTags: ตัวเล็ก, ตัดช่องว่าง/#, ไม่ซ้ำ, ไม่เกิน max อัน
func NormalizeTags(tags []string, max int) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, t := range tags {
		t = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(t, "#")))
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
		if len(out) == max {
			break
		}
	}
	return out
}
