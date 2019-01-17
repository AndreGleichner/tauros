package server

import (
	"crypto/sha256"
	"io"
	"os"
)

type TaurosServer struct {
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// For non-existing files this returns nil.
func fileSha256(path string) (hash []byte) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return
	}
	hash = h.Sum(nil)
	return
}
