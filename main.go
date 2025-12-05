package main

import "fmt"

// Função auxiliar (não é o entrypoint main) para evitar conflito.
func Hello() {
	fmt.Println("Hello from dummy package")
}