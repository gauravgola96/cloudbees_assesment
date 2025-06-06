package server

import (
	"fmt"
	"github.com/gauravgola96/cloudbees_assesment/pkg/jenkins"
	"github.com/gauravgola96/cloudbees_assesment/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type LogProxyServer struct {
	jenkins *jenkins.JClient
	cache   *storage.DiskCache
}

func NewLogProxy(url string, cacheDir string) *LogProxyServer {
	lps := &LogProxyServer{}
	lps.jenkins = jenkins.NewJenkinsClient(url)
	lps.cache, _ = storage.NewDiskCache(cacheDir)
	return lps
}

func LogProxyRoutes() *chi.Mux {
	mux := chi.NewMux()
	lps := NewLogProxy(JENKINBASELOGURL, "cache_logs")
	mux.Get("/logs/{build_id}", lps.HandleGetLog)
	mux.Head("/logs/{build_id}", lps.HandleHeadLog)
	return mux
}

func (lps *LogProxyServer) HandleGetLog(w http.ResponseWriter, r *http.Request) {
	buildID := chi.URLParam(r, "build_id")
	if buildID == "" {
		http.Error(w, "Build ID cannot be empty string", http.StatusBadRequest)
		return
	}
	subLogger := log.With().Str("module", "server.handler.HandleGetLog").Str("build_id", buildID).Logger()

	offSetParam := r.URL.Query().Get("offset")
	limitParam := r.URL.Query().Get("limit")

	offset := 0
	if offSetParam != "" {
		var err error
		offset, err = strconv.Atoi(offSetParam)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error in offset param : %s", err.Error()), http.StatusBadRequest)
			return
		}

		if offset < 0 {
			http.Error(w, fmt.Sprintf("Offset cannot be negative : %s", offSetParam), http.StatusBadRequest)
			return
		}
	}

	//Get file from jenkins server
	if !lps.cache.LogExists(buildID) {
		subLogger.Info().Msgf("Getting buildID=%s logs from jenkins server", buildID)
		logsBts, err := lps.jenkins.GetLogs(buildID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				subLogger.Error().Err(err).Msg("buildId not found")
				http.Error(w, fmt.Sprintf("BuildID=%s logs not present on jenkins server", buildID), http.StatusNotFound)
				return
			}
			subLogger.Error().Err(err).Msg("Error in fetching logs from jenkins server")
			http.Error(w, fmt.Sprintf("Failed to get logs from jenkins server : %s", err.Error()), http.StatusInternalServerError)
			return
		}

		err = lps.cache.StoreLog(logsBts, buildID)
		if err != nil {
			subLogger.Error().Err(err).Msg("Error in storing logs from jenkins server")
			http.Error(w, fmt.Sprintf("Failed to store logs for buildID=%s on disk : %s", buildID, err.Error()), http.StatusInternalServerError)
			return
		}
		subLogger.Info().Msgf("Cached buildID=%s successfully", buildID)
	}

	filePath := lps.cache.GetPath(buildID)
	file, err := os.Open(filePath)
	if err != nil {
		subLogger.Error().Err(err).Msgf("Error in opening logs from disk : %s", filePath)
		http.Error(w, fmt.Sprintf("Failed to open logs for buildID=%s from disk : %s", buildID, err.Error()), http.StatusInternalServerError)
		return
	}

	defer file.Close()
	totalLines, err := getTotalLines(filePath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error in counting lines : %s", err.Error()), http.StatusBadRequest)
		return
	}
	linesToServe := totalLines - offset

	//read limit param
	limit := -1
	if limitParam != "" {
		limit, err = strconv.Atoi(limitParam)
		if err != nil {
			http.Error(w, fmt.Sprintf("Error in offset param : %s", err.Error()), http.StatusBadRequest)
			return
		}
		if limit < linesToServe {
			linesToServe = limit
		}
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.Itoa(linesToServe-offset))

	reader := NewLineRangeReader(file, offset, limit)
	_, err = io.Copy(w, reader)
	if err != nil && err != io.EOF {
		subLogger.Error().Err(err).Msgf("Error copying file content %s", filePath)
	}

}

func (lps *LogProxyServer) HandleHeadLog(w http.ResponseWriter, r *http.Request) {
	buildID := chi.URLParam(r, "build_id")
	if buildID == "" {
		http.Error(w, "Build ID cannot be empty string", http.StatusBadRequest)
		return
	}
	subLogger := log.With().Str("module", "server.handler.HandleHeadLog").Str("build_id", buildID).Logger()

	if !lps.cache.LogExists(buildID) {
		subLogger.Info().Msgf("Getting buildID=%s logs from jenkins server", buildID)
		logsBts, err := lps.jenkins.GetLogs(buildID)
		if err != nil {
			if strings.Contains(err.Error(), "404") {
				subLogger.Error().Err(err).Msg("buildId not found")
				http.Error(w, fmt.Sprintf("BuildID=%s logs not present on jenkins server ", buildID), http.StatusNotFound)
				return
			}
			subLogger.Error().Err(err).Msg("Error in fetching logs from jenkins server")
			http.Error(w, fmt.Sprintf("Failed to get logs from jenkins server : %s", err.Error()), http.StatusInternalServerError)
			return
		}

		err = lps.cache.StoreLog(logsBts, buildID)
		if err != nil {
			subLogger.Error().Err(err).Msg("Error in storing logs from jenkins server")
			http.Error(w, fmt.Sprintf("Failed to store logs for buildID=%s on disk : %s", buildID, err.Error()), http.StatusInternalServerError)
			return
		}
		subLogger.Info().Msgf("Cached buildID=%s successfully", buildID)
	}

	filePath := lps.cache.GetPath(buildID)
	file, err := os.Open(filePath)
	if err != nil {
		subLogger.Error().Err(err).Msgf("Error in opening logs from disk : %s", filePath)
		http.Error(w, fmt.Sprintf("Failed to open logs for buildID=%s from disk : %s", buildID, err.Error()), http.StatusInternalServerError)
		return
	}

	defer file.Close()
	fileInfo, err := file.Stat()
	if err != nil {
		subLogger.Error().Err(err).Msgf("Error getting file info for %s: %v", filePath, err)
		http.Error(w, "Failed to get file info", http.StatusInternalServerError)
		return
	}
	totalSize := fileInfo.Size()
	w.Header().Set("Content-Type", "text/plain")
	w.Header().Set("Content-Length", strconv.FormatInt(totalSize, 10))
	w.WriteHeader(http.StatusOK)
}
