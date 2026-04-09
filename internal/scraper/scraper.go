package scraper

import (
	"context"
	"crypto/tls"
	"fmt"
	htmlesc "html"
	"net"
	"net/http"

	"strings"

	"github.com/comsma/nocitationneeded/internal/cache"
	utls "github.com/refraction-networking/utls"
	"golang.org/x/net/html"
	"golang.org/x/net/http2"
)

type Scraper struct {
	client *http.Client
}

// TODO fix resource leak warning
func dialTLS(ctx context.Context, network, addr string, _ *tls.Config) (net.Conn, error) {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	conn, err := (&net.Dialer{}).DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}
	uconn := utls.UClient(conn, &utls.Config{ServerName: host}, utls.HelloChrome_Auto)
	if err := uconn.HandshakeContext(ctx); err != nil {
		conn.Close()
		return nil, err
	}
	return uconn, nil
}

func New() *Scraper {
	return &Scraper{
		client: &http.Client{
			Transport: &http2.Transport{
				DialTLSContext: dialTLS,
			},
		},
	}
}

func (s *Scraper) Scrape(ctx context.Context, rawURL string) (*cache.Citation, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"macOS"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parsing HTML: %w", err)
	}

	meta := extractMeta(doc)
	ld := extractJSONLD(doc)
	if ld == nil {
		ld = &NewsArticle{}
	}

	ldAuthor := ""
	if len(ld.Author) > 0 {
		ldAuthor = string(ld.Author[0].Name)
	}

	citation := &cache.Citation{URL: rawURL}

	// title: JSON-LD headline -> og:title -> name:title -> <title> tag
	switch {
	case ld.Headline != "":
		citation.Title = ld.Headline
	case meta["property:og:title"] != "":
		citation.Title = meta["property:og:title"]
	case meta["name:title"] != "":
		citation.Title = meta["name:title"]
	default:
		citation.Title = meta["_title"]
	}

	// author: JSON-LD author -> dcterms.creator -> sailthru.author -> name:author -> property:article:author
	switch {
	case ldAuthor != "":
		citation.Author = ldAuthor
	case meta["name:dcterms.creator"] != "":
		citation.Author = meta["name:dcterms.creator"]
	case meta["name:sailthru.author"] != "":
		citation.Author = meta["name:sailthru.author"]
	case meta["name:author"] != "":
		citation.Author = meta["name:author"]
	default:
		citation.Author = meta["property:article:author"]
	}

	// publication_date: JSON-LD datePublished -> article:published_time -> dcterms.date -> name:date
	switch {
	case ld.DatePublished != "":
		citation.PublicationDate = ld.DatePublished
	case meta["property:article:published_time"] != "":
		citation.PublicationDate = meta["property:article:published_time"]
	case meta["name:dcterms.date"] != "":
		citation.PublicationDate = meta["name:dcterms.date"]
	default:
		citation.PublicationDate = meta["name:date"]
	}

	// publisher: JSON-LD publisher -> og:site_name -> name:publisher
	switch {
	case ld.Publisher.Name != "":
		citation.Publisher = ld.Publisher.Name
	case meta["property:og:site_name"] != "":
		citation.Publisher = meta["property:og:site_name"]
	default:
		citation.Publisher = meta["name:publisher"]
	}

	citation.Version = meta["name:version"]

	citation.Title = htmlesc.UnescapeString(citation.Title)
	citation.Author = htmlesc.UnescapeString(citation.Author)
	citation.Publisher = htmlesc.UnescapeString(citation.Publisher)

	return citation, nil
}

// extractJSONLD finds the first NewsArticle/Article/BlogPosting JSON-LD block
// in the HTML tree and returns it as a NewsArticle struct.
func extractJSONLD(n *html.Node) *NewsArticle {
	var walk func(*html.Node) *NewsArticle
	walk = func(node *html.Node) *NewsArticle {
		if node.Type == html.ElementNode && node.Data == "script" {
			for _, attr := range node.Attr {
				if attr.Key == "type" && attr.Val == "application/ld+json" && node.FirstChild != nil {
					if a := parseNewsArticleLD([]byte(node.FirstChild.Data)); a != nil {
						return a
					}
					break
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			if result := walk(c); result != nil {
				return result
			}
		}
		return nil
	}
	return walk(n)
}

// extractMeta walks the HTML node tree and collects meta tag content values
// keyed as "property:{value}" or "name:{value}", plus "_title" for <title> text.
func extractMeta(n *html.Node) map[string]string {
	result := make(map[string]string)
	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
			case "meta":
				var propKey, nameKey, content string
				for _, attr := range node.Attr {
					switch attr.Key {
					case "property":
						propKey = "property:" + strings.ToLower(attr.Val)
					case "name":
						nameKey = "name:" + strings.ToLower(attr.Val)
					case "content":
						content = attr.Val
					}
				}
				if propKey != "" {
					result[propKey] = content
				}
				if nameKey != "" {
					result[nameKey] = content
				}
			case "title":
				if node.FirstChild != nil {
					result["_title"] = node.FirstChild.Data
				}
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return result
}
