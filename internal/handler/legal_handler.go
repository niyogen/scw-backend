package handler

import (
	"fmt"
	"net/http"
	"strings"

	"delivery-backend/internal/legal"

	"github.com/gin-gonic/gin"
	"github.com/gomarkdown/markdown"
	mdhtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type LegalHandler struct{}

func NewLegalHandler() *LegalHandler { return &LegalHandler{} }

func (h *LegalHandler) Terms(c *gin.Context)   { h.renderDoc(c, legal.DocUserTC, "Terms & Conditions") }
func (h *LegalHandler) Privacy(c *gin.Context) { h.renderDoc(c, legal.DocUserPrivacy, "Privacy Policy") }

const legalPageTemplate = `<!DOCTYPE html>
<html lang="%s">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="index, follow">
<title>%s · Smart Sewa</title>
<style>
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, system-ui, sans-serif; max-width: 760px; margin: 0 auto; padding: 24px 20px 64px; line-height: 1.55; color: #1f2937; }
h1, h2, h3, h4 { line-height: 1.25; color: #111827; }
h1 { font-size: 1.75rem; margin-top: 0.5rem; }
h2 { font-size: 1.3rem; margin-top: 2rem; }
h3 { font-size: 1.1rem; margin-top: 1.5rem; }
a { color: #ea580c; }
code { background: #f3f4f6; padding: 2px 4px; border-radius: 3px; font-size: 0.9em; }
hr { border: 0; border-top: 1px solid #e5e7eb; margin: 2rem 0; }
.meta { font-size: 0.85rem; color: #6b7280; margin: 0 0 1.5rem; }
.lang-switcher { margin: 0 0 0.5rem; font-size: 0.9rem; }
.lang-switcher a { margin-right: 14px; }
.lang-switcher a.active { font-weight: 700; text-decoration: none; color: #111827; }
table { border-collapse: collapse; width: 100%%; }
th, td { border: 1px solid #e5e7eb; padding: 6px 10px; text-align: left; }
th { background: #f9fafb; }
</style>
</head>
<body>
<div class="lang-switcher">%s</div>
<div class="meta">%s · Version %s · Effective %s</div>
%s
</body>
</html>`

func (h *LegalHandler) renderDoc(c *gin.Context, docID, title string) {
	locale := strings.ToLower(strings.TrimSpace(c.Query("lang")))
	if locale == "" {
		locale = legal.AuthoritativeLang
	}
	body, servedLocale, err := legal.Body(docID, locale)
	if err != nil {
		c.String(http.StatusNotFound, "Document not available")
		return
	}
	version, _ := legal.Version(docID)
	effective, _ := legal.EffectiveDate(docID)

	parserExts := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(parserExts)
	renderer := mdhtml.NewRenderer(mdhtml.RendererOptions{
		Flags: mdhtml.CommonFlags | mdhtml.HrefTargetBlank,
	})
	htmlBody := markdown.ToHTML([]byte(body), p, renderer)

	var switcher strings.Builder
	path := c.Request.URL.Path
	for _, l := range legal.SupportedLocales() {
		cls := ""
		if l == servedLocale {
			cls = ` class="active"`
		}
		switcher.WriteString(fmt.Sprintf(`<a href="%s?lang=%s"%s>%s</a>`, path, l, cls, strings.ToUpper(l)))
	}

	page := fmt.Sprintf(legalPageTemplate,
		servedLocale, title, switcher.String(),
		legal.AppName(), version, effective,
		string(htmlBody),
	)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cache-Control", "public, max-age=300")
	c.String(http.StatusOK, page)
}
