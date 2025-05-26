package client

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type LogProxyClient struct {
	client *http.Client
	url    string
}

func NewLogProxyClient(url string) *LogProxyClient {
	lpc := &LogProxyClient{}
	lpc.client = &http.Client{
		Timeout: 60 * time.Second,
	}
	lpc.url = url
	return lpc
}

// DownloadLog : Get logs from log proxy server (bytes)
func (lpc *LogProxyClient) DownloadLog(logURL string) ([]byte, error) {
	resp, err := lpc.client.Get(fmt.Sprintf("%s/%s", lpc.url, logURL))
	if err != nil {
		return nil, fmt.Errorf("[Log Proxy Client] log request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[Log Proxy Client] status received : %s", resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[Log Proxy Client] failed to read response: %w", err)
	}
	return data, nil
}

func (lpc *LogProxyClient) HeadLog(logURL string) (int64, error) {
	resp, err := lpc.client.Head(fmt.Sprintf("%s/%s", lpc.url, logURL))
	if err != nil {
		return 0, fmt.Errorf("[Log Proxy Client] log request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("[Log Proxy Client] status received : %s", resp.Status)
	}

	contentLengthStr := resp.Header.Get("Content-Length")
	if contentLengthStr == "" {
		return 0, fmt.Errorf("Content-Length header not found in response for %s", logURL)
	}

	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse Content-Length '%s': %w", contentLengthStr, err)
	}

	return contentLength, nil
}

func (lpc *LogProxyClient) DeDupLogs(logs []byte) (string, error) {
	scanner := bufio.NewScanner(bytes.NewReader(logs))
	var deduplicatedLines []string
	lastLine := ""

	// example format: "[2025-05-20T06:06:45.859Z] "
	timestampRegex := regexp.MustCompile(`^\[\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z\] `)

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := line

		// removing leading date-time stamps
		if matches := timestampRegex.FindStringSubmatchIndex(line); len(matches) > 0 {
			trimmedLine = line[matches[1]:]
		}

		// removing consecutive identical lines
		currentProcessedLine := strings.TrimSpace(trimmedLine)
		if currentProcessedLine != lastLine {
			deduplicatedLines = append(deduplicatedLines, trimmedLine)
			lastLine = currentProcessedLine
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("Error scanning logs: %w\n", err)
	}

	return strings.Join(deduplicatedLines, "\n"), nil
}
