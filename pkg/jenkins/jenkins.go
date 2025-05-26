package jenkins

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type JClient struct {
	client *http.Client
	url    string
}

func NewJenkinsClient(url string) *JClient {
	jc := &JClient{}
	jc.client = &http.Client{
		Timeout: 120 * time.Second,
	}
	jc.url = url
	return jc
}

func (j *JClient) GetLogs(buildId string) ([]byte, error) {
	url := fmt.Sprintf("%s%s/consoleText", j.url, buildId)
	resp, err := j.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("[jenkins] log request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("[jenkins] status received %s: %s", buildId, resp.Status)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("[jenkins] failed to read response: %w", err)
	}

	return data, nil
}
