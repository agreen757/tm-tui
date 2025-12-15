package main

import (
	"fmt"

	"github.com/agreen757/tm-tui/internal/ui/dialog/demo"
)

func main() {
	fmt.Println("Running Dialog Component Validation...")
	fmt.Println()

	results := demo.ValidateDialogComponents()

	fmt.Println(results)
}
