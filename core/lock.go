package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// LockFile is the on-disk cache of per-resource installed
// version/hash, written after every successful apply. It is never
// substituted for a live LocalState() query; see reconcileWithLock.
type LockFile struct {
	Version   int          `json:"version"`
	Resources []LockedItem `json:"resources"`
}

// LockedItem is one cached resource entry in the lock file.
type LockedItem struct {
	Kind        ResourceKind `json:"kind"`
	Name        string       `json:"name"`
	Version     string       `json:"version"`
	InstalledAt time.Time    `json:"installedAt"`
	SourceMTime *time.Time   `json:"sourceMTime,omitempty"` // skills only
}

const lockFileVersion = 1

// DefaultLockPath returns the on-disk location of the devsync lock
// file, mirroring the mise provider's XDG-style snapshot convention.
// Entirely separate from the gitignored skills-lock.json used by
// gentle-ai's skill-registry mechanism (different path, different
// schema, different owner).
func DefaultLockPath() string {
	return filepath.Join(UserHomeDir(), ".local", "share", "devsync", "devsync-lock.json")
}

// LoadLock reads and decodes the lock file at path. A missing file is
// not an error — it returns an empty LockFile, since the lock is a
// cache, not a trust source. A corrupt/unreadable file returns a
// non-nil error alongside an empty LockFile; callers (core.Run) treat
// this as non-fatal and proceed as if no lock existed.
func LoadLock(path string) (LockFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LockFile{Version: lockFileVersion}, nil
		}
		return LockFile{Version: lockFileVersion}, fmt.Errorf("read lock %q: %w", path, err)
	}

	var lock LockFile
	if err := json.Unmarshal(data, &lock); err != nil {
		return LockFile{Version: lockFileVersion}, fmt.Errorf("decode lock %q: %w", path, err)
	}
	return lock, nil
}

// SaveLock atomically writes lock to path (write to a .tmp file, then
// rename), creating any missing parent directories.
func SaveLock(path string, lock LockFile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create lock dir: %w", err)
	}

	data, err := json.MarshalIndent(lock, "", "  ")
	if err != nil {
		return fmt.Errorf("encode lock: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o644); err != nil {
		return fmt.Errorf("write lock tmp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("rename lock tmp file: %w", err)
	}
	return nil
}

// reconcileWithLock is the hashing-cost optimization hook: LocalState()
// is always called and its freshly computed Resources are always what
// wins the diff — this function never substitutes live state for stale
// cached state. It only short-circuits by reusing the lock's cached
// Version for a resource whose on-disk SourceMTime (populated only by
// the skills provider's cheap mtime walk) exactly matches the lock's
// recorded SourceMTime for that resource, since an unchanged mtime
// means the resource's content is unchanged since the last successful
// apply. Any mismatch, missing lock entry, or missing SourceMTime
// leaves the freshly computed Resource as-is — this is a no-op for
// tools since core.Resource.SourceMTime is always nil for
// ResourceKindTool (mise's `mise ls --json` is already one batched
// subprocess call returning every tool at once, so there is nothing to
// skip). The lock lookup is keyed by (Kind, Name) via lockKey, with no
// separate branch on ResourceKind.
func reconcileWithLock(local []Resource, lock LockFile) []Resource {
	byKey := make(map[string]LockedItem, len(lock.Resources))
	for _, item := range lock.Resources {
		byKey[lockKey(item.Kind, item.Name)] = item
	}

	out := make([]Resource, len(local))
	for i, r := range local {
		out[i] = r
		cached, ok := byKey[lockKey(r.Kind, r.Name)]
		if !ok || cached.SourceMTime == nil || r.SourceMTime == nil {
			continue
		}
		if cached.SourceMTime.Equal(*r.SourceMTime) {
			out[i].Version = cached.Version
		}
	}
	return out
}

// updateLock upserts one LockedItem per successful ApplyResult
// (failed applies are not recorded — the item stays "out of sync" and
// reappears next run), keyed by (Kind, Name). KindRemoved successes
// delete the entry instead of upserting.
func updateLock(lock *LockFile, results []ApplyResult) {
	if lock.Version == 0 {
		lock.Version = lockFileVersion
	}

	index := make(map[string]int, len(lock.Resources))
	for i, item := range lock.Resources {
		index[lockKey(item.Kind, item.Name)] = i
	}

	now := time.Now().UTC()
	for _, r := range results {
		if r.Err != nil {
			continue
		}
		key := lockKey(r.Item.ResourceKind, r.Item.Name)

		if r.Item.Kind == KindRemoved {
			if i, ok := index[key]; ok {
				lock.Resources = append(lock.Resources[:i], lock.Resources[i+1:]...)
				delete(index, key)
				for k, v := range index {
					if v > i {
						index[k] = v - 1
					}
				}
			}
			continue
		}

		newItem := LockedItem{
			Kind:        r.Item.ResourceKind,
			Name:        r.Item.Name,
			Version:     r.Item.DesiredVersion,
			InstalledAt: now,
			SourceMTime: r.Item.DesiredMTime,
		}
		if i, ok := index[key]; ok {
			lock.Resources[i] = newItem
		} else {
			lock.Resources = append(lock.Resources, newItem)
			index[key] = len(lock.Resources) - 1
		}
	}

	sort.Slice(lock.Resources, func(i, j int) bool {
		if lock.Resources[i].Kind != lock.Resources[j].Kind {
			return lock.Resources[i].Kind < lock.Resources[j].Kind
		}
		return lock.Resources[i].Name < lock.Resources[j].Name
	})
}

func lockKey(kind ResourceKind, name string) string {
	return string(kind) + "\x00" + name
}
