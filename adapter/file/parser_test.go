package file

import (
	"fmt"
	"testing"
	"time"
)

func TestNewRawFileParser(t *testing.T) {
	parser := NewParser(&Config{
		RootPath:      ".",
		StartDatePath: "",
	})

	parser.FlushFileList()
	fmt.Println(len(parser.Files), parser.Files)
}

func TestGetNextDaysReadPath(t *testing.T) {
	now := time.Now()
	fmt.Println(getNextDaysReadPath(now))
}
