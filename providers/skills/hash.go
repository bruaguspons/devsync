// Package skills implements core.Provider for repo-declared Claude
// Code skills: it syncs skill directories from a version-controlled
// skills/ folder at the repo root to ~/.claude/skills, using
// SHA-256 directory-tree content hashing (not semver) to detect
// changes. It never reads or writes .atl/skill-registry.md or the
// gitignored skills-lock.json used by gentle-ai's skill-registry
// mechanism.
package skills

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bruaguspons/devsync/core"
)

// hashDirTree computes a deterministic SHA-256 hash over a skill
// directory's full file tree:
//  1. filepath.Walk the directory, collecting relative paths.
//  2. Sort relative paths lexicographically (explicit sort makes the
//     algorithm platform-independent, not just relying on Walk's
//     natural order).
//  3. For each file in sorted order, feed the hash: the relative path
//     (forward-slash normalized, UTF-8 bytes), a NUL separator, the
//     file's raw content bytes, another NUL separator.
//
// Including the path (not just content) means a rename-only change is
// detected as a hash change. Returns the hex-encoded digest and the
// max file mtime observed (for lock-file reconciliation), or an error
// if the directory cannot be walked/read.
func hashDirTree(dir string) (hexDigest string, maxMTime time.Time, err error) {
	var relPaths []string
	fileMTimes := map[string]time.Time{}

	walkErr := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		relPaths = append(relPaths, rel)
		fileMTimes[rel] = info.ModTime()
		return nil
	})
	if walkErr != nil {
		return "", time.Time{}, fmt.Errorf("walk %q: %w", dir, walkErr)
	}

	sort.Strings(relPaths)

	h := sha256.New()
	var maxSeen time.Time
	for _, rel := range relPaths {
		if _, err := io.WriteString(h, rel); err != nil {
			return "", time.Time{}, err
		}
		if _, err := h.Write([]byte{0}); err != nil {
			return "", time.Time{}, err
		}

		content, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return "", time.Time{}, fmt.Errorf("read %q: %w", rel, err)
		}
		if _, err := h.Write(content); err != nil {
			return "", time.Time{}, err
		}
		if _, err := h.Write([]byte{0}); err != nil {
			return "", time.Time{}, err
		}

		if t := fileMTimes[rel]; t.After(maxSeen) {
			maxSeen = t
		}
	}

	return fmt.Sprintf("%x", h.Sum(nil)), maxSeen, nil
}

// listSkillDirs returns the sorted names of every immediate
// subdirectory of root (each subdirectory name is a skill's
// identity). A missing root is not an error — it returns an empty
// slice, since neither ~/.claude/skills nor a repo's skills/ folder
// are guaranteed to exist.
func listSkillDirs(root string) ([]string, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("read dir %q: %w", root, err)
	}

	var names []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		if strings.HasPrefix(e.Name(), ".") {
			continue
		}
		names = append(names, e.Name())
	}
	sort.Strings(names)
	return names, nil
}

// loadSkillResources walks root's immediate skill subdirectories and
// returns one content-hashed core.Resource per skill, shared by both
// LocalState() (root = installedRoot) and DesiredState() (root =
// desiredRoot). errContext labels wrapped errors with the caller's
// intent (e.g. "hash installed skill" vs "hash repo skill") so error
// message format stays unchanged for each call site.
func loadSkillResources(root, errContext string) ([]core.Resource, error) {
	names, err := listSkillDirs(root)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(names))
	for _, name := range names {
		digest, maxMTime, err := hashDirTree(filepath.Join(root, name))
		if err != nil {
			return nil, fmt.Errorf("%s %q: %w", errContext, name, err)
		}
		resources = append(resources, core.Resource{
			Kind:        core.ResourceKindSkill,
			Name:        name,
			Version:     digest,
			SourceMTime: &maxMTime,
		})
	}
	return resources, nil
}
