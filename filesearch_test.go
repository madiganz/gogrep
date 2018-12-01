package main

import (
	"fmt"
	"testing"
)

func TestParseLine(t *testing.T) {
	fs := fileSearcher{}

	expected := fmt.Sprintf("%s", fs.modifyPrintColor("test", colorRed))
	actual, found := fs.parseLine("test", "test")
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s", "this is a ", fs.modifyPrintColor("test", colorRed))
	actual, found = fs.parseLine("this is a test", "test")
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s%s", "this is a ", fs.modifyPrintColor("test", colorRed), " this is a ", fs.modifyPrintColor("test", colorRed))
	actual, found = fs.parseLine("this is a test this is a test", "test")
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s%s%s", "this is a ", fs.modifyPrintColor("test", colorRed), " ", fs.modifyPrintColor("test", colorRed), " a is this")
	actual, found = fs.parseLine("this is a test test a is this", "test")
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s", fs.modifyPrintColor("test", colorRed), fs.modifyPrintColor("test", colorRed), fs.modifyPrintColor("test", colorRed))
	actual, found = fs.parseLine("testtesttest", "test")
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s%s%s%s%s", fs.modifyPrintColor("test", colorRed), " ", fs.modifyPrintColor("test", colorRed), " ", fs.modifyPrintColor("test", colorRed))
	actual, found = fs.parseLine("test test test", "test")
	if !found {
		t.Errorf("expected %t, but got %t instead", true, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}

	expected = fmt.Sprintf("%s", "expect nothing out of this")
	actual, found = fs.parseLine("expect nothing out of this", "test")
	if found {
		t.Errorf("expected %t, but got %t instead", false, found)
	}
	if expected != actual {
		t.Errorf("expected %q, but got %q instead", expected, actual)
	}
}
