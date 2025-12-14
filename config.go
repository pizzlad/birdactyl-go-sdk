package birdactyl

import (
	"log"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type HotConfig[T any] struct {
	path         string
	config       T
	mu           sync.RWMutex
	lastModified time.Time
	onChange     func(T)
	stopCh       chan struct{}
}

func NewHotConfig[T any](path string, defaultConfig T) *HotConfig[T] {
	h := &HotConfig[T]{
		path:   path,
		config: defaultConfig,
	}
	h.load()
	return h
}

func (h *HotConfig[T]) Get() T {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

func (h *HotConfig[T]) Set(config T) {
	h.mu.Lock()
	h.config = config
	h.mu.Unlock()
	h.Save()
}

func (h *HotConfig[T]) OnChange(fn func(T)) *HotConfig[T] {
	h.onChange = fn
	return h
}

func (h *HotConfig[T]) DynamicConfig() *HotConfig[T] {
	if h.stopCh != nil {
		return h
	}
	h.stopCh = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				h.checkReload()
			case <-h.stopCh:
				return
			}
		}
	}()
	return h
}

func (h *HotConfig[T]) StopWatching() {
	if h.stopCh != nil {
		close(h.stopCh)
		h.stopCh = nil
	}
}

func (h *HotConfig[T]) checkReload() {
	info, err := os.Stat(h.path)
	if err != nil {
		return
	}
	if info.ModTime().After(h.lastModified) {
		h.load()
		log.Printf("[HotConfig] Reloaded %s", h.path)
		if h.onChange != nil {
			h.mu.RLock()
			cfg := h.config
			h.mu.RUnlock()
			h.onChange(cfg)
		}
	}
}

func (h *HotConfig[T]) load() {
	data, err := os.ReadFile(h.path)
	if err != nil {
		h.Save()
		return
	}
	h.mu.Lock()
	yaml.Unmarshal(data, &h.config)
	h.mu.Unlock()
	if info, err := os.Stat(h.path); err == nil {
		h.lastModified = info.ModTime()
	}
}

func (h *HotConfig[T]) Save() {
	h.mu.RLock()
	data, _ := yaml.Marshal(h.config)
	h.mu.RUnlock()
	os.WriteFile(h.path, data, 0644)
	if info, err := os.Stat(h.path); err == nil {
		h.lastModified = info.ModTime()
	}
}
