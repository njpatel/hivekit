package main

import (
	"fmt"
	"os"

	"hive"
)

func main() {
	_, err := hive.Connect(hive.Config{
		Username: os.Getenv("HIVE_USER"),
		Password: os.Getenv("HIVE_PASS"),
	})
	fmt.Println(err)
}
