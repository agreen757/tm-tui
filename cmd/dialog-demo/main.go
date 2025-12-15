package main

import (
	"flag"
	"fmt"

	"github.com/agreen757/tm-tui/internal/ui/dialog/demo"
)

func main() {
	// Define flags
	validateFlag := flag.Bool("validate", false, "Run validation instead of interactive demo")
	flag.Parse()
	
	if *validateFlag {
		// Run validation
		fmt.Println("Running Dialog Component Validation...")
		fmt.Println()
		
		results := demo.ValidateDialogComponents()
		fmt.Println(results)
	} else {
		// Run interactive demo
		fmt.Println("Starting Dialog Component Demo")
		fmt.Println("Press 1-6 to show different dialog types, ? for help, and q to quit")
		
		demo.Run()
	}
}