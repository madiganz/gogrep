package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	colorable "github.com/mattn/go-colorable"
	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

// Color constants used for showing the found data.
const (
	colorRed     = "\x1b[31;1m"
	colorMagenta = "\x1b[35;1m"
	colorGreen   = "\x1b[32;1m"
	colorNormal  = "\x1b[0m"
)

// Helper to allow color printing in console on Windows.
var out = colorable.NewColorableStdout()

// lineItem holds information for a line.
// index corresponds to the line number location of the file.
// value corresponds to the actual line of text.
type lineItem struct {
	index int
	value string
}

// fileSearcher provides functionality for searching a text file for a pattern.
// parallelWorkers can be set in order to parallelize the search.
// By default, the fileSearcher is not case-sensitive.
type fileSearcher struct {
	parallelWorkers int
	caseSensitive   bool
}

// Search searches a file for a specific pattern. It will either run a sequential or
// parallel search, and will either do a case-sensitive or insenstive searcching depending
// on the fileSearcher's properties. Once the search is complete, it will return a list of
// lines that contain the pattern.
func (fs *fileSearcher) Search(fileName, pattern string) []lineItem {
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// According to https://godoc.org/golang.org/x/text/search#Matcher.CompileString,
	// using Matcher and Pattern allows for faster searching.
	matcher := &search.Matcher{}
	if !fs.caseSensitive {
		matcher = search.New(language.English, search.IgnoreCase)
	} else {
		matcher = search.New(language.English)
	}
	p := matcher.CompileString(pattern)

	if fs.parallelWorkers <= 1 {
		return fs.SearchSequential(f, p)
	}
	return fs.SearchParallel(f, p)
}

// SearchSequential searches a file sequentially by reading one line at a time and determining if
// that line contains the pattern.
func (fs *fileSearcher) SearchSequential(f *os.File, pattern *search.Pattern) []lineItem {
	// Create scanner for file to process each line.
	fileScanner := bufio.NewScanner(f)
	lineNumber := 1
	items := []lineItem{}
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if parsedLine, found := fs.processLine(line, pattern); found {
			items = append(items, lineItem{value: parsedLine, index: lineNumber})
		}
		lineNumber++
	}
	if err := fileScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading from file:", err)
	}
	return items
}

// SearchParallel searches a file using parallel pipeline. The pipeline consists of three stages.
// The first stage is solely responsible for reading from the file.
// The second stage is responsible for determining if a line of text contains the pattern.
// The third stage aggregrates all the results from second stage workers.
// Based on https://blog.golang.org/pipelines, except only uses one output channel for all the second stage workers.
func (fs *fileSearcher) SearchParallel(f *os.File, p *search.Pattern) []lineItem {
	var wg sync.WaitGroup
	processedLines := make(chan lineItem, 200)
	l := fs.generateLines(f)
	for i := 0; i < fs.parallelWorkers; i++ {
		wg.Add(1)
		fs.searchLine(l, processedLines, &wg, p)
	}
	writeChan := fs.outputWriter(processedLines)
	wg.Wait()             // Wait for all stage two workers to be done.
	close(processedLines) // Done parsing all lines. It is safe to close the channel.
	output := <-writeChan // Block until third stage is done.
	return output
}

// generateLines is first stage in the pipeline. Reads each line from the file and uses a channel
// to pass the line information to the second stage workers. Once all lines are read, close channel
// to signal workers that there will be no more incomming lines.
func (fs *fileSearcher) generateLines(f io.Reader) <-chan lineItem {
	lines := make(chan lineItem, 200)
	go func() {
		// Create scanner for file to process each line.
		fileScanner := bufio.NewScanner(f)
		lineNumber := 1
		for fileScanner.Scan() {
			line := fileScanner.Text()
			ld := lineItem{
				index: lineNumber,
				value: line,
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

// searchLine is the second stage in the pipeline. Processing incoming lines from the line channel and outputs them to the processedLines
// channel once the line has been processed.
func (fs *fileSearcher) searchLine(lines <-chan lineItem, processedLines chan<- lineItem, wg *sync.WaitGroup, p *search.Pattern) {
	go func() {
		for line := range lines {
			if parsedLine, found := fs.processLine(line.value, p); found {
				processedLines <- lineItem{
					index: line.index,
					value: parsedLine,
				}
			}
		}
		wg.Done()
	}()
}

// outputWriter is the third stage in the pipeline. Aggregates all the incoming processedLines into a list.
func (fs *fileSearcher) outputWriter(processedLines <-chan lineItem) <-chan []lineItem {
	output := make(chan []lineItem, 1)
	go func() {
		d := []lineItem{}
		for line := range processedLines {
			d = append(d, line)
		}
		output <- d
		close(output)
	}()
	return output
}

// processLine looks recursively through a line to find every instance of the pattern in the line.
func (fs *fileSearcher) processLine(s string, p *search.Pattern) (string, bool) {
	var parsedString string
	start, end := p.IndexString(s)
	if start > -1 && end > -1 {
		sString := s[:start]
		foundPattern := s[start:end]
		parsedString += fmt.Sprintf("%s%s", sString, fs.modifyPrintColor(foundPattern, colorRed))
		if end < len(s) {
			out, _ := fs.processLine(s[end:], p)
			parsedString += out
		}
		return parsedString, true
	}
	return parsedString + s, false
}

// modifyPrintColor changes the color of the string based on the provided color.
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
	var caseSensitive bool
	var verbose bool
	flag.IntVar(&numWorkers, "p", 4, "Number of parallel workers. Specify 1 for sequential processing.")
	flag.BoolVar(&caseSensitive, "cs", false, "Perform a case-sensitive search.")
	flag.BoolVar(&verbose, "v", false, "Output will show each line with match.")
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
		caseSensitive:   caseSensitive,
	}

	start := time.Now()
	items := searcher.Search(fileName, pattern)
	elapsed := time.Since(start)

	// Need to use different printing technique to write color in console on Windows, which results in the printing being the bottleneck. This makes it
	// worthwile to do printing after processing has been done and not during processing to get accurate measurements.
	if verbose {
		for _, l := range items {
			fmt.Fprintf(out, "%s%s %s\n", searcher.modifyPrintColor(fmt.Sprint(l.index), colorGreen), searcher.modifyPrintColor(":", colorMagenta), l.value)
		}
	}

	fmt.Printf("Total number of lines with match: %d\n", len(items))
	fmt.Printf("Total time to process file: %-8v\n", elapsed)
}
