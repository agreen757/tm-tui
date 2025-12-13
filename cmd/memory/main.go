package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/adriangreen/tm-tui/internal/memory"
)

func main() {
	// Define command flags
	storeCmd := flag.NewFlagSet("store", flag.ExitOnError)
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	logCmd := flag.NewFlagSet("log", flag.ExitOnError)
	
	// store command flags
	storeKey := storeCmd.String("key", "", "Key for the memory to store")
	storeFile := storeCmd.String("file", "", "File path to read content from (use '-' for stdin)")
	storeVal := storeCmd.String("value", "", "Value to store (alternative to file)")
	storeJSON := storeCmd.Bool("json", false, "Treat input as JSON")
	
	// get command flags
	getKey := getCmd.String("key", "", "Key for the memory to retrieve")
	
	// delete command flags
	deleteKey := deleteCmd.String("key", "", "Key for the memory to delete")
	
	// list command flags
	listPrefix := listCmd.String("prefix", "", "Prefix for filtering keys")
	listJSON := listCmd.Bool("json", false, "Output as JSON")
	
	// log command flags
	logTaskID := logCmd.String("task", "", "Task ID to log activity for")
	logActivity := logCmd.String("message", "", "Activity message to log")
	
	// Create helper instance
	helper, err := memory.DefaultHelper()
	if err != nil {
		fmt.Printf("Error creating memory helper: %v\n", err)
		os.Exit(1)
	}
	defer helper.Close()
	
	// Check if no arguments provided
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}
	
	// Determine which command is being run
	ctx := context.Background()
	switch os.Args[1] {
	
	case "store":
		storeCmd.Parse(os.Args[2:])
		if *storeKey == "" {
			fmt.Println("Error: Key is required for store command")
			storeCmd.PrintDefaults()
			os.Exit(1)
		}
		
		var data []byte
		
		// Get data from file or value
		if *storeFile != "" {
			var err error
			if *storeFile == "-" {
				data, err = io.ReadAll(os.Stdin)
			} else {
				data, err = os.ReadFile(*storeFile)
			}
			
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				os.Exit(1)
			}
		} else if *storeVal != "" {
			data = []byte(*storeVal)
		} else {
			fmt.Println("Error: Either -file or -value must be provided")
			storeCmd.PrintDefaults()
			os.Exit(1)
		}
		
		// Handle JSON format
		if *storeJSON {
			var jsonObj interface{}
			if err := json.Unmarshal(data, &jsonObj); err != nil {
				fmt.Printf("Error parsing JSON: %v\n", err)
				os.Exit(1)
			}
			
			if err := helper.StoreJSON(ctx, *storeKey, jsonObj); err != nil {
				fmt.Printf("Error storing JSON: %v\n", err)
				os.Exit(1)
			}
		} else {
			if err := helper.Store.Store(ctx, *storeKey, data); err != nil {
				fmt.Printf("Error storing data: %v\n", err)
				os.Exit(1)
			}
		}
		
		fmt.Printf("Successfully stored memory with key: %s\n", *storeKey)
		
	case "get":
		getCmd.Parse(os.Args[2:])
		if *getKey == "" {
			fmt.Println("Error: Key is required for get command")
			getCmd.PrintDefaults()
			os.Exit(1)
		}
		
		data, err := helper.Store.Retrieve(ctx, *getKey)
		if err != nil {
			if err == memory.ErrKeyNotFound {
				fmt.Printf("Key not found: %s\n", *getKey)
			} else {
				fmt.Printf("Error retrieving data: %v\n", err)
			}
			os.Exit(1)
		}
		
		// Print data to stdout
		fmt.Print(string(data))
		
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteKey == "" {
			fmt.Println("Error: Key is required for delete command")
			deleteCmd.PrintDefaults()
			os.Exit(1)
		}
		
		if err := helper.Store.Delete(ctx, *deleteKey); err != nil {
			fmt.Printf("Error deleting data: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Successfully deleted memory with key: %s\n", *deleteKey)
		
	case "list":
		listCmd.Parse(os.Args[2:])
		
		keys, err := helper.Store.List(ctx, *listPrefix)
		if err != nil {
			fmt.Printf("Error listing keys: %v\n", err)
			os.Exit(1)
		}
		
		if *listJSON {
			jsonData, err := json.MarshalIndent(keys, "", "  ")
			if err != nil {
				fmt.Printf("Error converting to JSON: %v\n", err)
				os.Exit(1)
			}
			fmt.Println(string(jsonData))
		} else {
			if len(keys) == 0 {
				fmt.Println("No keys found")
			} else {
				for _, key := range keys {
					fmt.Println(key)
				}
			}
		}
		
	case "log":
		logCmd.Parse(os.Args[2:])
		if *logTaskID == "" || *logActivity == "" {
			fmt.Println("Error: Both task ID and activity message are required")
			logCmd.PrintDefaults()
			os.Exit(1)
		}
		
		if err := helper.LogTaskActivity(ctx, *logTaskID, *logActivity); err != nil {
			fmt.Printf("Error logging activity: %v\n", err)
			os.Exit(1)
		}
		
		fmt.Printf("Successfully logged activity for task: %s\n", *logTaskID)
		
	case "readmes":
		readmes, err := helper.ListReadmes(ctx)
		if err != nil {
			fmt.Printf("Error listing READMEs: %v\n", err)
			os.Exit(1)
		}
		
		if len(readmes) == 0 {
			fmt.Println("No READMEs found")
		} else {
			fmt.Println("Available READMEs:")
			for _, name := range readmes {
				fmt.Println("-", name)
			}
		}
		
	case "help":
		printUsage()
		
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	progName := filepath.Base(os.Args[0])
	fmt.Printf(`Memory Storage for AI Agents - Usage:

  %s store -key <key> [-file <file>|-value <value>] [-json]
  %s get -key <key>
  %s delete -key <key>
  %s list [-prefix <prefix>] [-json]
  %s log -task <task_id> -message <activity>
  %s readmes
  %s help

Commands:
  store     Store data in memory
  get       Retrieve data from memory
  delete    Remove data from memory
  list      List all stored memory keys
  log       Log activity for a task
  readmes   List all stored README files
  help      Show this help message

Examples:
  %s store -key "readme:main" -file README.md
  %s store -key "task:1.2" -value "Implement user auth" -json
  %s get -key "readme:main"
  %s list -prefix "task:"
  %s log -task "1.2" -message "Started implementation"

`, progName, progName, progName, progName, progName, progName, progName, progName, progName, progName, progName, progName)
}