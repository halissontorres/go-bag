package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/halissontorres/go-bag/pkg/enum"
)

func main() {
	typeName := flag.String("type", "", "name of the type to generate the enum for")
	flag.Parse()

	if *typeName == "" {
		fmt.Fprintf(os.Stderr, "Usage: enumgen -type=<TypeName>\n")
		os.Exit(1)
	}

	// The current directory is where the command runs (typically via go generate).
	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	err = enum.Generate(dir, *typeName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate enum: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Enum successfully generated for %s\n", *typeName)
}
