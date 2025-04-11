package main

import (
	"fmt"

	"github.com/Cmolloy36/blog_aggregator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Println(fmt.Errorf("error: %w", err))
	}

	cfg.SetUser("Cooper")

	new_cfg, err := config.Read()
	if err != nil {
		fmt.Println(fmt.Errorf("error: %w", err))
	}
	fmt.Printf("Read config: %+v\n", new_cfg)
}
