package main

import (
	"fmt"
	"testing"

	"golang.org/x/text/language"
	"golang.org/x/text/search"
)

func TestProcessLine(t *testing.T) {
	fs := fileSearcher{}

	matcher := search.New(language.English)
	p := matcher.CompileString("test")

	expected := fmt.Sprintf("%s", fs.modifyPrintColor("test", colorRed))
	actual, found := fs.processLine("test", p)
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s", "this is a ", fs.modifyPrintColor("test", colorRed))
	actual, found = fs.processLine("this is a test", p)
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s%s", "this is a ", fs.modifyPrintColor("test", colorRed), " this is a ", fs.modifyPrintColor("test", colorRed))
	actual, found = fs.processLine("this is a test this is a test", p)
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s%s%s", "this is a ", fs.modifyPrintColor("test", colorRed), " ", fs.modifyPrintColor("test", colorRed), " a is this")
	actual, found = fs.processLine("this is a test test a is this", p)
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s", fs.modifyPrintColor("test", colorRed), fs.modifyPrintColor("test", colorRed), fs.modifyPrintColor("test", colorRed))
	actual, found = fs.processLine("testtesttest", p)
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s%s%s", fs.modifyPrintColor("test", colorRed), " ", fs.modifyPrintColor("test", colorRed), " ", fs.modifyPrintColor("test", colorRed))
	actual, found = fs.processLine("test test test", p)
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s", "expect nothing out of this")
	actual, found = fs.processLine("expect nothing out of this", p)
	if found {
		t.Errorf("expected %t, but got %t instead", false, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}
}
