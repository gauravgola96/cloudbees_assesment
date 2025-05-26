package storage

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

var ErrNilCache = errors.New("nil cache")

type DiskCache struct {
	filePath string
}

func NewDiskCache(directory string) (*DiskCache, error) {
	err := os.Mkdir(directory, 0755)
	if err != nil {
		return nil, fmt.Errorf("error creating cache directory %s : %w", directory, err)
	}
	dc := &DiskCache{
		filePath: directory,
	}
	return dc, nil
}

func (dc *DiskCache) GetPath(buildID string) string {
	return filepath.Join(dc.filePath, fmt.Sprintf("%s.log", buildID))
}

func (dc *DiskCache) StoreLog(data []byte, buildId string) error {
	return os.WriteFile(dc.GetPath(buildId), data, 0644)
}

func (dc *DiskCache) LogExists(buildId string) bool {
	_, err := os.Stat(dc.GetPath(buildId))
	return !os.IsNotExist(err)
}
