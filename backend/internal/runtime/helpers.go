package runtime

import (
	"os"
	"path/filepath"
)

func projectBatchTemplatePath() string {
	return filepath.Clean("./data/ad/templates/创建AD用户模板.xlsx")
}

func projectBatchUploadDir() string {
	return filepath.Clean("./data/ad/uploads")
}

func projectBatchExcelFiles() ([]string, error) {
	dir := projectBatchUploadDir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		ext := filepath.Ext(name)
		if ext == ".xlsx" || ext == ".xls" {
			files = append(files, name)
		}
	}
	return files, nil
}
