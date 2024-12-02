module backend

go 1.23.2

require (
	github.com/joho/godotenv v1.5.1
	github.com/lib/pq v1.10.9
)

// DB Migrations
require (
	github.com/golang-migrate/migrate/v4 v4.18.1 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	go.uber.org/atomic v1.7.0 // indirect
)

// Chromedp Web Scraper
require (
	github.com/chromedp/cdproto v0.0.0-20241003230502-a4a8f7c660df // indirect
	github.com/chromedp/chromedp v0.11.0
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	golang.org/x/sys v0.26.0 // indirect
)
