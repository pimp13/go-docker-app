package model

import (
	"regexp"
	"strings"
)

func generateSlug(title string) string {
	// حذف کاراکترهای غیرمجاز و تبدیل به lowercase
	slug := strings.ToLower(title)
	slug = regexp.MustCompile(`[^a-z0-9\s-]`).ReplaceAllString(slug, "")
	slug = regexp.MustCompile(`[\s]+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")

	// حذف کاراکترهای فارسی اضافی اگر لازم باشد (یا از کتابخانه transliteration استفاده کنید)
	return slug
}
