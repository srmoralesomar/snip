package clipboard

import "github.com/atotto/clipboard"

// SystemReader reads from the real system clipboard.
type SystemReader struct{}

// ReadAll returns the current clipboard contents.
func (SystemReader) ReadAll() (string, error) {
	return clipboard.ReadAll()
}
