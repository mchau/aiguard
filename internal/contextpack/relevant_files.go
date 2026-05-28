package contextpack

import (
	"os"
)

// RelevantFiles returns up to maxFiles source files whose paths appear in changedFiles,
// capped so no single file exceeds maxBytes.
func RelevantFiles(rootDir string, changedFiles []string, maxFiles, maxBytes int) []ContextFile {
	var result []ContextFile
	for _, path := range changedFiles {
		if len(result) >= maxFiles {
			break
		}
		absPath := rootDir + "/" + path
		info, err := os.Stat(absPath)
		if err != nil || info.IsDir() {
			continue
		}
		if int(info.Size()) > maxBytes {
			continue
		}
		content, err := os.ReadFile(absPath)
		if err != nil {
			continue
		}
		result = append(result, ContextFile{
			Path:    path,
			Content: string(content),
			Bytes:   len(content),
		})
	}
	return result
}
