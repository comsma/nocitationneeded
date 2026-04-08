package scraper

import (
	"encoding/json"
)

// parseNewsArticleLD parses a JSON-LD blob (object or array) and returns the
// first NewsArticle, Article, or BlogPosting entry found, or nil.
func parseNewsArticleLD(data []byte) *NewsArticle {
	isArticle := func(a *NewsArticle) bool {
		switch a.Type {
		case "NewsArticle", "Article", "BlogPosting", "ReportageNewsArticle", "ScholarlyArticle":
			return true
		}
		return false
	}

	var single NewsArticle
	if json.Unmarshal(data, &single) == nil && isArticle(&single) {
		return &single
	}

	var arr []NewsArticle
	json.Unmarshal(data, &arr) //nolint:errcheck — partial population is acceptable
	for i := range arr {
		if isArticle(&arr[i]) {
			return &arr[i]
		}
	}
	return nil
}

// NewsArticle represents the schema.org NewsArticle type as found in
// application/ld+json script tags. Polymorphic fields (author, image) have
// custom unmarshalers to handle the variants sites produce in practice.
type NewsArticle struct {
	Context          string       `json:"@context"`
	Type             string       `json:"@type"`
	Headline         string       `json:"headline"`
	AlternativeHdl   string       `json:"alternativeHeadline"`
	Description      string       `json:"description"`
	ArticleBody      string       `json:"articleBody"`
	ArticleSection   string       `json:"articleSection"`
	URL              string       `json:"url"`
	DatePublished    string       `json:"datePublished"`
	DateModified     string       `json:"dateModified"`
	Keywords         []string     `json:"keywords"`
	Author           Authors      `json:"author"`
	Publisher        Organization `json:"publisher"`
	Image            ArticleImage `json:"image"`
	MainEntityOfPage WebPage      `json:"mainEntityOfPage"`
}

// Person represents a schema.org Person.
type Person struct {
	Type string     `json:"@type"`
	Name PersonName `json:"name"`
	URL  string     `json:"url"`
}

// PersonName handles name as a plain string or an array of strings (e.g. NPR).
type PersonName string

func (n *PersonName) UnmarshalJSON(data []byte) error {
	var s string
	if json.Unmarshal(data, &s) == nil {
		*n = PersonName(s)
		return nil
	}
	var arr []string
	if json.Unmarshal(data, &arr) == nil && len(arr) > 0 {
		*n = PersonName(arr[0])
	}
	return nil
}

// Authors handles author as a single Person object or an array of Person objects.
type Authors []Person

func (a *Authors) UnmarshalJSON(data []byte) error {
	var single Person
	if json.Unmarshal(data, &single) == nil && single.Name != "" {
		*a = Authors{single}
		return nil
	}
	var multi []Person
	if json.Unmarshal(data, &multi) == nil {
		*a = Authors(multi)
	}
	return nil
}

// Organization represents a schema.org Organization.
type Organization struct {
	Type string `json:"@type"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

// ImageObject represents a schema.org ImageObject.
type ImageObject struct {
	Type   string `json:"@type"`
	URL    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	Name   string `json:"name"`
}

// ArticleImage handles image as an ImageObject, a plain URL string, or an
// array of either (e.g. The Guardian uses []string).
type ArticleImage struct {
	ImageObject
}

func (i *ArticleImage) UnmarshalJSON(data []byte) error {
	var s string
	if json.Unmarshal(data, &s) == nil {
		i.URL = s
		return nil
	}
	if json.Unmarshal(data, &i.ImageObject) == nil && i.URL != "" {
		return nil
	}
	// array — take the first element
	var arr []json.RawMessage
	if json.Unmarshal(data, &arr) == nil && len(arr) > 0 {
		return i.UnmarshalJSON(arr[0])
	}
	return nil
}

// WebPage represents a schema.org WebPage used in mainEntityOfPage.
type WebPage struct {
	Type string `json:"@type"`
	ID   string `json:"@id"`
}

func (w *WebPage) UnmarshalJSON(data []byte) error {
	var s string
	if json.Unmarshal(data, &s) == nil {
		w.ID = s
		return nil
	}
	type webPageAlias WebPage
	return json.Unmarshal(data, (*webPageAlias)(w))
}
