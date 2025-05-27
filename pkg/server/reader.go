package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
)

type LineRangeReader struct {
	scanner       *bufio.Scanner
	currentLine   int
	startLine     int
	endLine       int
	buffer        *bytes.Buffer
	err           error
	streamStarted bool
}

func NewLineRangeReader(file *os.File, startLine, endLine int) *LineRangeReader {
	return &LineRangeReader{
		scanner:   bufio.NewScanner(file),
		startLine: startLine,
		endLine:   endLine,
		buffer:    &bytes.Buffer{},
	}
}

func (r *LineRangeReader) Read(p []byte) (int, error) {
	if r.err != nil {
		return 0, r.err
	}

	for r.buffer.Len() == 0 {
		if !r.scanner.Scan() {
			if err := r.scanner.Err(); err != nil {
				r.err = fmt.Errorf("error reading file lines: %w", err)
				return 0, r.err
			}
			r.err = io.EOF
			break
		}

		line := r.scanner.Text()
		r.currentLine++

		if r.currentLine < r.startLine {
			continue
		}

		if r.endLine != -1 && r.currentLine > r.endLine {
			r.err = io.EOF
			break
		}

		r.buffer.WriteString(line + "\n")
	}

	return r.buffer.Read(p)
}

func getTotalLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	var count int
	for scanner.Scan() {
		count++
	}
	if err := scanner.Err(); err != nil {

		return 0, fmt.Errorf("error in counting lines %w", err)
	}

	return count, nil
}
