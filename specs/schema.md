# Data Schema

## Redis

### Citation Cache
**Key:** `cite:{md5 of url}`
**TTL:** 1 hour
**Type:** Hash

| Field              | Type     | Description                   |
|--------------------|----------|-------------------------------|
| `title`            | `string` | Article title                 |
| `author`           | `string` | Article author                |
| `publication_date` | `string` | Publication date (ISO 8601)   |
| `url`              | `string` | Original article URL          |
| `publisher`        | `string` | Publishing institution        |
| `version`          | `string` | Version number of publication |

**Example:**
```json
{
  "title": "The Future of Unix",
  "author": "John Smith",
  "date": "2026-01-15",
  "url": "https://example.com/article",
  "publisher": "Bell Labs Technical Journal",
  "version": "Volume 1,"
}
```

## API Routes

| Method | Path      | Description                                     | Response     |
|--------|-----------|-------------------------------------------------|--------------|
| GET    | `/`       | Serve index, optionally restore cached citation | `text/html`  |
| POST   | `/cite`   | Scrape URL, cache result, return partial        | HTML partial |
| GET    | `/health` | Health check                                    | `200 OK`     |

---

### GET /
**Query Params**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `ref` | `string` | ❌ | Redis key to restore cached citation |

### POST /cite

**Form Params**
| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `url` | `string` | ✅ | Article URL to scrape |
| `style` | `string` | ❌ | Citation style: `apa`, `mla`, `chicago`, `ieee`, ' (default: `apa`) |

**Response**
- Success: renders `citation.html` partial, HTMX swaps into `#result`
- Failure: renders `error.html` partial, HTMX swaps into `#result`

