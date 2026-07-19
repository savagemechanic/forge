package runtime

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

func generateID() string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return hex.EncodeToString(h[:])[:12]
}

func folderID(path string) string {
	h := sha256.Sum256([]byte(path))
	return hex.EncodeToString(h[:])[:12]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
