package main

import app "url-shortener/internal/app/url-shortener"

func main() {
	us := app.New()
	app.Run(us)
}
