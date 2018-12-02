package main

import (
	"fmt"
	"os"
)

func main() {
	f, err := os.Create("large_test_file.txt")
	if err != nil {
		panic(err)
	}
	for i := 1; i <= 10000000; i++ {
		f.WriteString(fmt.Sprintf("%s %d %s\n", "this is a test ", i, " this is a test"))
	}
	f.Close()

	f, err = os.Create("test_longer_lines.txt")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 500000; i++ {
		f.WriteString("This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line. This is a longer line.This is a longer line.This is a longer line.\n")
	}
	f.Close()

	f, err = os.Create("test_error_appears_once.txt")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 500000; i++ {
		if i != 123456 {
			f.WriteString("[INFO] This is an example log info statement.\n")
		} else {
			f.WriteString("[ERROR] Houston we have a problem!\n")
		}
	}
	f.Close()
}
