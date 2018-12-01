package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"runtime"
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

type searcher interface {
	Search(s, expression string)
	SearchParallel(s, expression string)
}

type fileSearcher struct {
}

type lineData struct {
	number int
	data   string
}

func (fs fileSearcher) Search(fileName, expression string) {
	start := time.Now()
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Create scanner for file to process each line
	fileScanner := bufio.NewScanner(f)
	lineNumber := 1
	for fileScanner.Scan() {
		line := fileScanner.Text()
		if parsedLine, found := fs.parseLine(line, expression); found {
			fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(lineNumber), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsedLine)
		}
		lineNumber++
	}
	if err := fileScanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading from file:", err)
	}
	fmt.Println(time.Since(start))
}

func (fs fileSearcher) SearchParallel(fileName, expression string) {
	start := time.Now()
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	l := fs.generateLines(f)
	// pl1 := fs.searchLine(l, expression)
	// pl2 := fs.searchLine(l, expression)

	// for line := range merge(pl1, pl2) {
	// 	fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(line.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), line.data)
	// }
	var wg sync.WaitGroup
	parsedLines := make(chan lineData, 100)
	numWorkers := int(math.Max(1.0, float64(runtime.NumCPU()-1)))
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		fs.searchLine(l, parsedLines, &wg, expression)
	}
	// fs.searchLine(l, parsedLines, &wg, expression)
	// fs.searchLine(l, parsedLines, &wg, expression)

	errChan := fs.outputWriter(parsedLines)
	wg.Wait()
	close(parsedLines)
	<-errChan //Block

	// for line := range parsedLines {
	// 	fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(line.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), line.data)
	// }
	// close(parsedLines)

	// lines := make(chan lineData, 100)
	// parsedLines := make(chan lineData, 100)
	// var wg sync.WaitGroup
	// // numWorkers := int(math.Max(1.0, float64(runtime.NumCPU()-1)))

	// // for i := 0; i < numWorkers; i++ {
	// for i := 0; i < 2; i++ {
	// 	go fs.searchLine(i, lines, parsedLines, &wg, expression)
	// }

	// go func(lines <-chan lineData, wg *sync.WaitGroup) {
	// 	for parsed := range lines {
	// 		fmt.Println(parsed.data)
	// 		// fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(parsed.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsed.data)
	// 	}
	// }(parsedLines, &wg)

	// // Create scanner for file to process each line
	// fileScanner := bufio.NewScanner(f)
	// lineNumber := 1
	// for fileScanner.Scan() {
	// 	line := fileScanner.Text()
	// 	ld := lineData{
	// 		number: lineNumber,
	// 		data:   line,
	// 	}
	// 	wg.Add(1)
	// 	lines <- ld
	// 	// if parsedLine, found := fs.parseLine(line, expression); found {
	// 	// 	fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(lineNumber), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsedLine)
	// 	// }
	// 	lineNumber++
	// }
	// if err := fileScanner.Err(); err != nil {
	// 	fmt.Fprintln(os.Stderr, "reading from file:", err)
	// }

	// close(lines)
	// wg.Wait()
	fmt.Println(time.Since(start))
}

func (fs fileSearcher) generateLines(f io.Reader) <-chan lineData {
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

func (fs fileSearcher) searchLine(lines <-chan lineData, parsedLines chan<- lineData, wg *sync.WaitGroup, expression string) {
	go func() {
		for line := range lines {
			// fmt.Printf("worker: %d processing line: %d\n", id, line.number)
			if parsedLine, found := fs.parseLine(line.data, expression); found {
				// fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(line.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsedLine)
				// fmt.Printf("worker %d found success on line %s\n", id, parsedLine)
				parsedLines <- lineData{
					number: line.number,
					data:   parsedLine,
				}
			}
		}
		wg.Done()
	}()
}

func (fs fileSearcher) outputWriter(parsedLines <-chan lineData) <-chan error {
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

// func (fs fileSearcher) searchLine(lines <-chan lineData, expression string) <-chan lineData {
// 	parsedLines := make(chan lineData, 100)
// 	go func() {
// 		for line := range lines {
// 			// fmt.Printf("worker: %d processing line: %d\n", id, line.number)
// 			if parsedLine, found := fs.parseLine(line.data, expression); found {
// 				// fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(line.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsedLine)
// 				// fmt.Printf("worker %d found success on line %s\n", id, parsedLine)
// 				parsedLines <- lineData{
// 					number: line.number,
// 					data:   parsedLine,
// 				}
// 			}
// 		}
// 		close(parsedLines)
// 	}()
// 	return parsedLines
// }

// func (fs fileSearcher) searchLine(id int, lines <-chan lineData, output chan<- lineData, wg *sync.WaitGroup, expression string) {
// 	for line := range lines {
// 		// fmt.Printf("worker: %d processing line: %d\n", id, line.number)
// 		if parsedLine, found := fs.parseLine(line.data, expression); found {
// 			// fmt.Fprintf(out, "%s%s %s\n", fs.modifyPrintColor(fmt.Sprint(line.number), colorGreen), fs.modifyPrintColor(":", colorMagenta), parsedLine)
// 			// fmt.Printf("worker %d found success on line %s\n", id, parsedLine)
// 			output <- lineData{
// 				number: line.number,
// 				data:   parsedLine,
// 			}
// 		}
// 		wg.Done()
// 	}
// }

// parseLineRecursive looks recursively through a line to find every instance of the expression in the line
func (fs fileSearcher) parseLine(s, exp string) (string, bool) {
	var parsedString string
	i := strings.Index(s, exp)
	if i > -1 {
		expLength := len(exp)
		sString := s[0:i]
		expString := s[i : i+expLength]
		parsedString += fmt.Sprintf("%s%s", sString, fs.modifyPrintColor(expString, colorRed))
		if i+expLength < len(s) {
			out, _ := fs.parseLine(s[i+expLength:], exp)
			parsedString += out
		}
		return parsedString, true
	}
	return parsedString + s, false
}

func (fs fileSearcher) modifyPrintColor(s, color string) string {
	return fmt.Sprintf("%s%s%s", color, s, colorNormal)
}

func main() {
	var fileName string
	var filePath string
	var searchString string
	flag.StringVar(&searchString, "s", "", "-s hello world")
	flag.StringVar(&fileName, "f", "", "-f temp.go")
	flag.StringVar(&filePath, "fp", "./", "-fp ./temp/")
	flag.Parse()

	if searchString == "" || fileName == "" {
		panic("whoops") // TODO: Fix
	}

	searcher := fileSearcher{}
	// search(searcher, fileName, searchString)
	searcher.SearchParallel(fileName, searchString)
}

func search(searcher searcher, fileName, expression string) {
	searcher.Search(fileName, expression)
}

// https://blog.golang.org/pipelines
// merges multiple channels into 1
// func merge(cs ...<-chan lineData) <-chan lineData {
// 	var wg sync.WaitGroup
// 	out := make(chan lineData)

// 	// Start an output goroutine for each input channel in cs.  output
// 	// copies values from c to out until c is closed, then calls wg.Done.
// 	output := func(c <-chan lineData) {
// 		for n := range c {
// 			out <- n
// 		}
// 		wg.Done()
// 	}
// 	wg.Add(len(cs))
// 	for _, c := range cs {
// 		go output(c)
// 	}

// 	// Start a goroutine to close out once all the output goroutines are
// 	// done.  This must start after the wg.Add call.
// 	go func() {
// 		wg.Wait()
// 		close(out)
// 	}()
// 	return out
// }
