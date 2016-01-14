package main

import (
	"fmt"
	"os"
	"time"

	"hive"
)

func main() {
	_, err := hive.Connect(hive.Config{
		Username:        os.Getenv("HIVE_USER"),
		Password:        os.Getenv("HIVE_PASS"),
		RefreshInterval: 10 * time.Second,
	})

	if err != nil {
		fmt.Println(err)
	}

	for {
	}
}
