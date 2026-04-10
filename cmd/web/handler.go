package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/comsma/nocitationneeded/internal/cache"
	"github.com/comsma/nocitationneeded/internal/config"
	"github.com/comsma/nocitationneeded/internal/scraper"
	"github.com/labstack/echo/v5"
)

var (
	siteKey = "055a0c5a-ccde-4c32-bd2a-912c6d88b8ca"
)

//TODO global template registration
//TODO rate limit middleware

var funcs = template.FuncMap{
	"formatDate": func(s string) string {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return s
		}
		return t.Format("January, 02, 2006")
	},
}

type Handler struct {
	e             *echo.Echo
	citationCache *cache.CitationCache
	scraper       *scraper.Scraper
	cfg           *config.Config
}

func NewHandler(e *echo.Echo, citationCache *cache.CitationCache, s *scraper.Scraper, cfg *config.Config) *Handler {
	return &Handler{
		e:             e,
		citationCache: citationCache,
		scraper:       s,
		cfg:           cfg,
	}
}

func (h *Handler) GetHome(c *echo.Context) error {
	tmpl, err := template.New("").Funcs(funcs).ParseGlob("ui/views/*.gohtml")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data := struct {
		Citation *cache.Citation
		SiteKey  string
	}{}

	data.SiteKey = h.cfg.HCaptcha.SiteKey
	ref := c.QueryParam("ref")
	if ref != "" {
		citation, err := h.citationCache.Get(c.Request().Context(), ref)
		if err == nil {
			data.Citation = citation
		}
	}

	return tmpl.ExecuteTemplate(c.Response(), "base", data)
}

func (h *Handler) PostCite(c *echo.Context) error {
	tmpl, err := template.New("").Funcs(funcs).ParseGlob("ui/views/*.gohtml")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	renderError := func(msg string) error {
		return tmpl.ExecuteTemplate(c.Response(), "error", struct{ Error string }{Error: msg})
	}

	captchaToken := c.FormValue("h-captcha-response")

	verified, _, err := h.verifyToken(captchaToken, c.RealIP())
	if err != nil {
		return renderError(err.Error())
	}
	if !verified {
		return renderError("captcha verification failed")
	}

	rawURL := c.FormValue("url")
	if rawURL == "" {
		c.Response().WriteHeader(http.StatusBadRequest)
		return renderError("url is required")
	}

	style := c.FormValue("style")
	if style == "" {
		style = "apa"
	}
	_ = style

	sum := md5.Sum([]byte(rawURL))
	key := "cite:" + hex.EncodeToString(sum[:])

	citation, err := h.scraper.Scrape(c.Request().Context(), rawURL)
	if err != nil {
		return renderError(err.Error())
	}

	if err := h.citationCache.Set(c.Request().Context(), key, *citation); err != nil {
		return renderError(err.Error())
	}

	data := struct {
		Citation *cache.Citation
		Key      string
	}{
		Citation: citation,
		Key:      key,
	}

	return tmpl.ExecuteTemplate(c.Response(), "citation", data)
}

func (h *Handler) GetHealth(c *echo.Context) error {
	return c.String(http.StatusOK, "OK")
}

func (h *Handler) verifyToken(token, ip string) (bool, []string, error) {
	form := url.Values{
		"secret":   {h.cfg.HCaptcha.Secret},
		"response": {token},
		"remoteip": {ip},
		"sitekey":  {h.cfg.HCaptcha.SiteKey},
	}
	resp, err := http.PostForm(
		"https://api.hcaptcha.com/siteverify",
		form,
	)
	if err != nil {
		return false, nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Success    bool     `json:"success"`
		ErrorCodes []string `json:"error-codes"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return false, nil, err
	}
	if out.Success {
		return true, []string{}, nil
	}
	return false, out.ErrorCodes, nil
}
