package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/darleet/iu9-tofl-regex/internal/service/parser"
)

func main() {
	s := parser.New()

	f, err := os.Open("regex.txt")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer f.Close()

	sc := bufio.NewScanner(f)

	for sc.Scan() {
		line := sc.Text()
		_, err := s.Parse(context.Background(), line)
		if err != nil {
			fmt.Printf("Incorrect regex: %s\n", line)
		} else {
			fmt.Printf("Correct regex: %s\n", line)
		}
	}

	if err := sc.Err(); err != nil {
		fmt.Println("Error reading file:", err)
	}
}
