package jenkins

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDownloadLog_NotFound(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Not Found"))
	}))
	defer testServer.Close()

	client := &JClient{
		client: testServer.Client(),
		url:    testServer.URL + "/job/Core/job/jenkins/job/master/",
	}

	buildID := "nonexistent"
	_, err := client.GetLogs(buildID)
	if err == nil {
		t.Fatal("DownloadLog expected an error, got nil")
	}
	expectedError := "Jenkins returned: 404 Not Found"
	if err.Error() != expectedError {
		t.Errorf("Expected error %q, got %q", expectedError, err.Error())
	}
}
