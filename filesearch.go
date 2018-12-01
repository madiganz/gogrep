package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	colorable "github.com/mattn/go-colorable"
)

const (
	colorRed     = "\x1b[31;1m"
	colorMagenta = "\x1b[35;1m"
	colorGreen   = "\x1b[32;1m"
	colorNormal  = "\x1b[0m"
)

var out = colorable.NewColorableStdout()

type fileSearcher struct {
	parallelWorkers int
}

type lineData struct {
	number int
	data   string
}

func (fs *fileSearcher) Search(fileName, pattern string) {
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if fs.parallelWorkers <= 1 {
		fs.SearchSequential(f, pattern)
	} else {
		fs.SearchParallel(f, pattern)
	}
}

func (fs *fileSearcher) SearchSequential(f *os.File, pattern string) {
	// Create scanner for file to process each line
	fileScanner := bufio.NewScanner(f)
	lineNumber := 1
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if parsedLine, found := fs.parseLine(line, pattern); found {
			fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(lineNumber), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsedLine)
		}
		lineNumber++
	}
	if err := fileScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading from file:", err)
	}
}

func (fs *fileSearcher) SearchParallel(f *os.File, pattern string) {
	var wg sync.WaitGroup
	parsedLines := make(chan lineData, 100)
	l := fs.generateLines(f)
	for i := 0; i < fs.parallelWorkers; i++ {
		wg.Add(1)
		fs.searchLine(l, parsedLines, &wg, pattern)
	}
	writeChan := fs.outputWriter(parsedLines)
	wg.Wait()
	close(parsedLines) // Done parsing all lines. It is safe to close the channel
	<-writeChan        // Block until write is done
}

func (fs *fileSearcher) generateLines(f io.Reader) <-chan lineData {
	lines := make(chan lineData, 100)
	go func() {
		// Create scanner for file to process each line
		fileScanner := bufio.NewScanner(f)
		lineNumber := 1
		for fileScanner.Scan() {
			line := fileScanner.Text()
			ld := lineData{
				number: lineNumber,
				data:   line,
			}
			lines <- ld
			lineNumber++
		}
		if err := fileScanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "reading from file:", err)
		}
		close(lines)
	}()
	return lines
}

func (fs *fileSearcher) searchLine(lines <-chan lineData, parsedLines chan<- lineData, wg *sync.WaitGroup, pattern string) {
	go func() {
		for line := range lines {
			if parsedLine, found := fs.parseLine(line.data, pattern); found {
				parsedLines <- lineData{
					number: line.number,
					data:   parsedLine,
				}
			}
		}
		wg.Done()
	}()
}

func (fs *fileSearcher) outputWriter(parsedLines <-chan lineData) <-chan error {
	err := make(chan error, 1)
	go func() {
		for line := range parsedLines {
			fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(line.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), line.data)
		}
		err <- nil
		close(err)
	}()
	return err
}

// parseLine looks recursively through a line to find every instance of the pattern in the line
func (fs *fileSearcher) parseLine(s, pattern string) (string, bool) {
	var parsedString string
	i := strings.Index(s, pattern)
	if i > -1 {
		pLength := len(pattern)
		sString := s[0:i]
		foundPattern := s[i : i+pLength]
		parsedString += fmt.Sprintf("%s%s", sString, fs.modifyPrintColor(foundPattern, colorRed))
		if i+pLength < len(s) {
			out, _ := fs.parseLine(s[i+pLength:], pattern)
			parsedString += out
		}
		return parsedString, true
	}
	return parsedString + s, false
}

func (fs *fileSearcher) modifyPrintColor(s, color string) string {
	return fmt.Sprintf("%s%s%s", color, s, colorNormal)
}

func main() {
	flag.Usage = func() {
		fmt.Printf("%s by Zachary Madigan\n", os.Args[0])
		fmt.Println("Usage:")
		fmt.Printf("\tgogrep [flags] pattern file\n")
		fmt.Println("Flags:")
		flag.PrintDefaults()
	}

	var numWorkers int
	flag.IntVar(&numWorkers, "p", 4, "Number of parallel workers. Specify 1 for sequential processing.")
	flag.Parse()

	if flag.NArg() < 2 {
		flag.Usage()
		os.Exit(1)
	}

	pattern := flag.Arg(0)
	fileName := flag.Arg(1)

	if _, err := os.Stat(fileName); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	searcher := &fileSearcher{
		parallelWorkers: numWorkers,
	}

	start := time.Now()
	searcher.Search(fileName, pattern)
	fmt.Println(time.Since(start))
}
