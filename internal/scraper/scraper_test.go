package scraper

import (
	"context"
	"testing"
	"time"
)

// sites contains a representative sample of news/article sites with known
// bot detection. Each entry has the URL to scrape and the minimum fields
// expected in a successful response.
var sites = []struct {
	name string
	url  string
}{
	{
		name: "The Hill",
		url:  "https://thehill.com/homenews/administration/5820244-vance-warns-iran-war-chaos/",
	},
	{
		name: "Medium",
		url:  "https://medium.com/@yawgyamfiprempeh27/i-built-a-production-grade-kubernetes-platform-in-48-hours-db5629fba0e3",
	},
	{
		name: "BBC News",
		url:  "https://www.bbc.com/news/articles/czjwp1vjn9lo",
	},
	{
		name: "Reuters",
		url:  "https://www.reuters.com/world/asia-pacific/trump-agrees-two-week-ceasefire-iran-says-safe-passage-through-hormuz-possible-2026-04-08/",
	},
	{
		name: "The Guardian",
		url:  "https://www.theguardian.com/world/2026/apr/08/palestinian-girl-who-lost-arm-in-israeli-missile-attack-on-gaza-arrives-in-uk",
	},
	{
		name: "NPR",
		url:  "https://www.npr.org/2026/04/08/nx-s1-5771648/sarcasm-origin-history-etymology",
	},
	{
		name: "AP News",
		url:  "https://apnews.com/live/iran-war-israel-trump-04-08-2026",
	},
	{
		name: "NY Times (no paywall section)",
		url:  "https://www.nytimes.com/2026/04/07/theater/cats-jellicle-ball-review-broadway.html",
	},
}

func TestScrapeBot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in short mode")
	}

	s := New()

	for _, tc := range sites {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			citation, err := s.Scrape(ctx, tc.url)
			if err != nil {
				t.Errorf("FAIL  %s: %v", tc.name, err)
				return
			}

			t.Logf("OK    %s", tc.name)
			t.Logf("      title:     %q", citation.Title)
			t.Logf("      author:    %q", citation.Author)
			t.Logf("      publisher: %q", citation.Publisher)
			t.Logf("      date:      %q", citation.PublicationDate)

			if citation.Title == "" {
				t.Errorf("      WARN: no title extracted")
			}
		})
	}
}
