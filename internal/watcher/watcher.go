package watcher

import (
	"github.com/fsnotify/fsnotify"
	"github.com/aakash-a-dev/blackboard/internal/logger"
)

// Watch calls onChange whenever the file at path is written.
// Blocks until the watcher is closed or encounters a fatal error.
func Watch(path string, onChange func()) error {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer w.Close()

	if err := w.Add(path); err != nil {
		return err
	}

	for {
		select {
		case event, ok := <-w.Events:
			if !ok {
				return nil
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				logger.Info("  reloading " + path + " ...")
				onChange()
			}
		case err, ok := <-w.Errors:
			if !ok {
				return nil
			}
			logger.Error("watcher: " + err.Error())
		}
	}
}
