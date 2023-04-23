package main

import (
	"fmt"
	"os"
)
const exitSuccess = 0

func main() {
	// Print a greeting message
	fmt.Println("hello world")

	// Exit with a success status code
	if err := os.Exit(exitSuccess); 
        err != nil {
		panic(err)
	}
}
	
