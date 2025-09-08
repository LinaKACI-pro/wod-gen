package main

import (
	"fmt"
	"os"
	"time"

	"github.com/LinaKACI-pro/wod-gen/pkg"
)

func main() {
	duration := 24 * time.Hour
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		fmt.Fprintln(os.Stderr, "missing AUTH_JWT_SECRET")
		os.Exit(1)
	}

	sub := "demo-user"
	if len(os.Args) > 1 {
		sub = os.Args[1]
	}

	jwtManager := pkg.NewJWTManager(secret, duration)
	token, err := jwtManager.Generate(sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(token)
}
