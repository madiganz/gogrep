package main

import (
	"os"
	"path/filepath"
)

const (
	smLine        = "test\n"
	mdLine        = "this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test\n"
	lgLine        = "this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test this is a test\n"
	searchPattern = "hello"
)

func generateFile(fileName string, numOfLines int, lineData string, patternDivisor int) {
	f, err := os.Create(filepath.Join("./test/", fileName))
	if err != nil {
		panic(err)
	}
	for i := 1; i <= numOfLines; i++ {
		if i%patternDivisor == 0 {
			f.WriteString(searchPattern + lineData)
		} else {
			f.WriteString(lineData)
		}
	}
	f.Close()
}

func main() {
	os.MkdirAll("./test", os.ModePerm)

	generateFile("sm_sm.txt", 100, smLine, 10)
	generateFile("sm_md.txt", 100, mdLine, 10)
	generateFile("sm_lg.txt", 100, lgLine, 10)

	generateFile("md_sm.txt", 10000, smLine, 1000)
	generateFile("md_md.txt", 10000, mdLine, 1000)
	generateFile("md_lg.txt", 10000, lgLine, 1000)

	generateFile("lg_sm.txt", 2000000, smLine, 100000)
	generateFile("lg_md.txt", 2000000, mdLine, 100000)
	generateFile("lg_lg.txt", 2000000, lgLine, 100000)

	generateFile("xl_sm.txt", 10000000, smLine, 500000)
	generateFile("xl_md.txt", 10000000, mdLine, 500000)
	generateFile("xl_lg.txt", 10000000, lgLine, 500000)
}
