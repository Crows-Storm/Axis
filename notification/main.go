package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/gommon/log"
)

//TIP <p>To run your code, right-click the code and select <b>Run</b>.</p> <p>Alternatively, click
// the <icon src="AllIcons.Actions.Execute"/> icon in the gutter and select the <b>Run</b> menu item from here.</p>

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Warn("No .env.example file found, using system environment variables")
	}
}
