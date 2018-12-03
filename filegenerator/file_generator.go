package main

import (
	"os"
	"path/filepath"
)

const (
	smLine = "test\n"
	mdLine = "this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test\n"
	lgLine = "this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test\n"
)

func generateFile(fileName string, numOfLines int, lineData string) {
	f, err := os.Create(filepath.Join("./test/", fileName))
	if err != nil {
		panic(err)
	}
	for i := 1; i <= numOfLines; i++ {
		f.WriteString(lineData)
	}
	f.Close()
}

func main() {
	os.MkdirAll("./test", os.ModePerm)

	generateFile("sm_sm.txt", 100, smLine)
	generateFile("sm_md.txt", 100, mdLine)
	generateFile("sm_lg.txt", 100, lgLine)

	generateFile("md_sm.txt", 10000, smLine)
	generateFile("md_md.txt", 10000, mdLine)
	generateFile("md_lg.txt", 10000, lgLine)

	generateFile("lg_sm.txt", 2000000, smLine)
	generateFile("lg_md.txt", 2000000, mdLine)
	generateFile("lg_lg.txt", 2000000, lgLine)

	generateFile("xl_sm.txt", 10000000, smLine)
	generateFile("xl_md.txt", 10000000, mdLine)
	generateFile("xl_lg.txt", 10000000, lgLine)
}
