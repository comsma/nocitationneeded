package main

import (
	"crypto/md5"
	"encoding/hex"
	"html/template"
	"net/http"

	"github.com/comsma/nocitationneeded/internal/cache"
	"github.com/comsma/nocitationneeded/internal/scraper"
	"github.com/labstack/echo/v5"
)

var (
	siteKey = "055a0c5a-ccde-4c32-bd2a-912c6d88b8ca"
)

type Handler struct {
	e             *echo.Echo
	citationCache *cache.CitationCache
	scraper       *scraper.Scraper
}

func NewHandler(e *echo.Echo, citationCache *cache.CitationCache, s *scraper.Scraper) *Handler {
	return &Handler{
		e:             e,
		citationCache: citationCache,
		scraper:       s,
	}
}

func (h *Handler) GetHome(c *echo.Context) error {
	tmpl, err := template.ParseGlob("ui/views/*.gohtml")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	data := struct {
		Citation *cache.Citation
		SiteKey  string
	}{}

	data.SiteKey = siteKey
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
	tmpl, err := template.ParseGlob("ui/views/*.gohtml")
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	renderError := func(msg string) error {
		return tmpl.ExecuteTemplate(c.Response(), "error", struct{ Error string }{Error: msg})
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
