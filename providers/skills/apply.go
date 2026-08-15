package skills

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/bruaguspons/devsync/core"
)

// Apply performs one confirmed skill DiffItem's install/update/remove
// action, touching only the single named subdirectory under
// ~/.claude/skills — never the parent, and never .atl/skill-registry.md
// or skills-lock.json.
func (p *Provider) Apply(item core.DiffItem) error {
	dst := filepath.Join(p.installedRoot, item.Name)

	switch item.Kind {
	case core.KindNew, core.KindUpdate:
		src := filepath.Join(p.desiredRoot, item.Name)
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("remove stale install of %q: %w", item.Name, err)
		}
		if err := copyDir(src, dst); err != nil {
			return fmt.Errorf("install skill %q: %w", item.Name, err)
		}
		return nil
	case core.KindRemoved:
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("remove skill %q: %w", item.Name, err)
		}
		return nil
	default:
		return fmt.Errorf("apply: unknown diff kind %q for skill %q", item.Kind, item.Name)
	}
}

// copyDir recursively copies src to dst, creating directories as
// needed and preserving file modes.
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}

		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target, info)
	})
}

// copyFile copies one file's contents from src to dst, creating dst
// with the source's mode, then setting dst's mtime to match src's
// mtime — preserving the source's mtime instead of leaving the
// copy-time timestamp, so a future mtime-based lock optimization is
// not misled by copy-time timestamps.
func copyFile(src, dst string, info os.FileInfo) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	return os.Chtimes(dst, info.ModTime(), info.ModTime())
}
