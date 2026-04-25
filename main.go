package main

import (
	"log"

	"purecore/cmd"

	"github.com/joho/godotenv"
)

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default ports")
	}

	cmd.Execute()
}
