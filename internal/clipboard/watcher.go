package clipboard

import (
	"context"
	"time"
)

// Reader abstracts clipboard read access for testability.
type Reader interface {
	ReadAll() (string, error)
}

// Clip represents a newly detected clipboard entry.
type Clip struct {
	Content   string
	Timestamp time.Time
}

// Watcher polls the clipboard and emits new clips over a channel.
type Watcher struct {
	reader   Reader
	interval time.Duration
	clips    chan Clip
}

// NewWatcher creates a Watcher using the given Reader and poll interval.
func NewWatcher(r Reader, interval time.Duration) *Watcher {
	return &Watcher{
		reader:   r,
		interval: interval,
		clips:    make(chan Clip, 16),
	}
}

// Clips returns the read-only channel of detected clips.
func (w *Watcher) Clips() <-chan Clip {
	return w.clips
}

// Start begins polling the clipboard. It blocks until ctx is cancelled,
// then closes the clips channel.
func (w *Watcher) Start(ctx context.Context) {
	defer close(w.clips)

	var last string
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			content, err := w.reader.ReadAll()
			if err != nil || content == "" || content == last {
				continue
			}
			last = content
			select {
			case w.clips <- Clip{Content: content, Timestamp: time.Now()}:
			case <-ctx.Done():
				return
			}
		}
	}
}
