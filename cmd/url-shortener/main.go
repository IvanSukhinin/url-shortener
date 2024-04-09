package main

import (
	"url-shortener/internal/app/us"
)

func main() {
	app := us.New()
	app.Run()
}
