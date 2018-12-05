# gogrep
File searcher that has a sequential or parallel processing option. Optimized for large files.

## Get source code
go get -u github.com/madiganz/gogrep

## Build on Windows
- for Windows: go build
- cross combine for Linux (CS1) using git bash: $ env GOOS=linux GOARCH=386 go build

#### Run on Windows
./gogrep.exe [-flags] pattern filename

#### Run on Linux
./gogrep [-flags] pattern filename (may need to run chmod +x gogrep)

### Flags
- -cs - Perform case-sensitive search. Default is false.
- -p - Number of parallel workers to use. Default is 4. Specify 1 for sequential processing.
- -v - Output all lines that match the search pattern

## Testing
### Build and Run
#### On Windows
From inside ./gogrep/ - go build -o filegenerator.exe .\filegenerator

To run - ./filegenerator.exe

#### On Linux
Follow similar pattern as for gogrep.exe

This will create multiple files in a ./test directory that can be used to test the performance of the file searcher.