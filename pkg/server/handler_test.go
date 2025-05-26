package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
)

func TestHandleHeadLog_DownloadAndCache(t *testing.T) {

	server := NewLogProxy(JENKINBASELOGURL, "cached_logs")
	buildID := "7372"
	jenkinsContent := "Content from Jenkins for HEAD new build."

	req := httptest.NewRequest(http.MethodHead, "http://localhost:8080/logs/"+buildID, nil)
	rr := httptest.NewRecorder()
	server.HandleHeadLog(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expectedLen := len(jenkinsContent)
	if rr.Header().Get("Content-Length") != strconv.Itoa(expectedLen) {
		t.Errorf("Content-Length header mismatch. Expected %d, got %s", expectedLen, rr.Header().Get("Content-Length"))
	}

	if !server.cache.LogExists(buildID) {
		t.Errorf("Log was not cached after HEAD download")
	}

	cachedData, _ := os.ReadFile(server.cache.GetPath(buildID))
	if string(cachedData) != jenkinsContent {
		t.Errorf("Cached content mismatch. Expected %q, got %q", jenkinsContent, string(cachedData))
	}
}
