package main

import (
	"fmt"
	"os"
	"sshpky/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// str := "                          \t   Last login: Tue Sep 30 13:46:42 2025 from 192.168.82.102"
	// str = strings.Trim(str, " ")
	// fmt.Println(str)
}
