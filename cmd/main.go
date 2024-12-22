package main

import "medods-tz/internal/app"

const configPath = "config/config.yaml"

func main() {
	app.Run(configPath)
}
