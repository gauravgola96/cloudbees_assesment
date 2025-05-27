package main

import (
	"flag"
	"fmt"
	"github.com/gauravgola96/cloudbees_assesment/pkg/client"
	"log"
	"net/url"
	"os"
	"strconv"
)

func main() {

	var headRequest bool
	flag.BoolVar(&headRequest, "head", false, "Perform a HEAD request to get Content-Length only")
	flag.Parse()
	args := flag.Args()

	//build_id [offset] [limit] [output_file]
	if len(args) < 1 || len(args) > 4 {
		fmt.Println("Usage: go run cmd/client/main.go [-head] <build_id> [offset] [limit] [output_file]")
		fmt.Println("Examples:")
		fmt.Println("  go run cmd/client/main.go 7372                        # Download and deduplicate full log for build 100")
		fmt.Println("  go run cmd/client/main.go -head 7372                  # Get Content-Length for build 100")
		fmt.Println("  go run cmd/client/main.go 7372 500                    # Download from offset 500 for build 100")
		fmt.Println("  go run cmd/client/main.go 7372 0 1000                 # Download first 1000 bytes for build 100")
		fmt.Println("  go run cmd/client/main.go 7372 500 1000 output.txt    # Download with offset/limit and save to file")
		os.Exit(1)
	}

	buildID := args[0]

	logClient := client.NewLogProxyClient("http://localhost:8080")
	logURL := fmt.Sprintf("logs/%s", buildID)

	if headRequest {
		contentLength, err := logClient.HeadLog(logURL)
		if err != nil {
			log.Fatalf("Error getting Content-Length for build %s: %v", buildID, err)
		}
		fmt.Printf("Content-Length for build %s: %d bytes\n", buildID, contentLength)
		return
	}

	offset := -1
	limit := -1
	outputFile := ""

	// Parse optional arguments
	argIndex := 2
	if len(os.Args) > argIndex {
		parsedOffset, err := strconv.Atoi(os.Args[argIndex])
		if err == nil {
			offset = parsedOffset
			argIndex++
		} else {
			argIndex++
		}
	}

	if len(os.Args) > argIndex {
		parsedLimit, err := strconv.Atoi(os.Args[argIndex])
		if err == nil {
			limit = parsedLimit
			argIndex++
		} else {
			argIndex++
		}
	}

	if len(os.Args) > argIndex {
		outputFile = os.Args[argIndex]
	}

	//logClient = client.NewLogProxyClient("http://localhost:8080")

	queryParams := url.Values{}
	if offset != -1 {
		queryParams.Set("offset", strconv.Itoa(offset))
	}
	if limit != -1 {
		queryParams.Set("limit", strconv.Itoa(limit))
	}

	//logURL := fmt.Sprintf("logs/%s", buildID)
	if len(queryParams) > 0 {
		logURL = fmt.Sprintf("%s?%s", logURL, queryParams.Encode())
	}

	logContent, err := logClient.DownloadLog(logURL)
	if err != nil {
		log.Fatalf("Error downloading log for build %s: %v", buildID, err)
	}

	deduplicatedContent, _ := logClient.DeDupLogs(logContent)

	if outputFile != "" {
		err := os.WriteFile(outputFile, []byte(deduplicatedContent), 0644)
		if err != nil {
			log.Fatalf("Error writing deduplicated log to file %s: %v", outputFile, err)
			return
		}
		fmt.Printf("Deduplicated log for build %s saved to %s\n", buildID, outputFile)
	} else {
		fmt.Println("--- Deduplicated Log Output ---")
		fmt.Println(deduplicatedContent)
		fmt.Println("-----------------------------")
	}
}
