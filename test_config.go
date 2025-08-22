package main

import (
    "fmt"
    "github.com/mohan-s-gopal/redis-valkey-tui/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        fmt.Printf("Config loaded successfully (error expected if no config exists): %v
", err)
    } else {
        fmt.Printf("Config loaded: %+v
", cfg)
    }
}
