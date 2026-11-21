package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	hash := "$2a$10$zYNLZs6izQZOl79sBs9Zq.oAwcy84oGE30gYDauh4nTzq7V.NhY.G"
	candidates := []string{"password", "admin", "123456", "12345678", "secret", "orazbakhov"}

	for _, p := range candidates {
		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(p))
		if err == nil {
			fmt.Printf("MATCH FOUND: %s\n", p)
			return
		}
	}
	fmt.Println("NO MATCH")
}
