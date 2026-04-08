## Docs
- API routes: see `specs/schema.md`
- Redis schema: see `specs/schema.md`

## Project Structure
```
/nocitationneeded
| -- cmd/
| ---- web/
| ------ main.go # entry point
| ------ handlers.go # api route handlers
| -- internal/ # helpers and internal logic for app
| ---- cache/
| ------ citation.go # cache for citations
| ------ redis.go # redis client
| -- ui/
| ---- views/ # htmx views for website
| ------ base.layout.gohtml
| ------ footer.partial.gohtml
| ------ nav.partial.gohtml
| ------ home.page.gohtml
| ------ citation.page.gohtml
| ---- static/ # static resources
| ------ img/
| ------ css/
| -------- style.css
```
