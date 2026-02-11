package cleanup8

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCleanup8(t *testing.T) {
	cleanup := NewCleanup8()
	assert.NotNil(t, cleanup)
}

func TestCleanup8_CleanupDirectory(t *testing.T) {
	t.Run("removes old files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create an old file
		oldFile := filepath.Join(tmpDir, "old_file.txt")
		err := os.WriteFile(oldFile, []byte("old content"), 0644)
		require.NoError(t, err)

		// Set modification time to 2 hours ago
		oldTime := time.Now().Add(-2 * time.Hour)
		err = os.Chtimes(oldFile, oldTime, oldTime)
		require.NoError(t, err)

		// Create a new file
		newFile := filepath.Join(tmpDir, "new_file.txt")
		err = os.WriteFile(newFile, []byte("new content"), 0644)
		require.NoError(t, err)

		cleanup := NewCleanup8()
		err = cleanup.CleanupDirectory(tmpDir, 1*time.Hour)
		require.NoError(t, err)

		// Old file should be removed
		_, err = os.Stat(oldFile)
		assert.True(t, os.IsNotExist(err), "old file should be removed")

		// New file should still exist
		_, err = os.Stat(newFile)
		assert.NoError(t, err, "new file should still exist")
	})

	t.Run("keeps files newer than maxAge", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a new file
		newFile := filepath.Join(tmpDir, "new_file.txt")
		err := os.WriteFile(newFile, []byte("content"), 0644)
		require.NoError(t, err)

		cleanup := NewCleanup8()
		err = cleanup.CleanupDirectory(tmpDir, 24*time.Hour)
		require.NoError(t, err)

		// File should still exist
		_, err = os.Stat(newFile)
		assert.NoError(t, err)
	})

	t.Run("handles empty directory", func(t *testing.T) {
		tmpDir := t.TempDir()

		cleanup := NewCleanup8()
		err := cleanup.CleanupDirectory(tmpDir, 1*time.Hour)
		assert.NoError(t, err)
	})

	t.Run("handles non-existent directory", func(t *testing.T) {
		cleanup := NewCleanup8()
		err := cleanup.CleanupDirectory("/non/existent/path", 1*time.Hour)
		assert.Error(t, err)
	})

	t.Run("skips subdirectories", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create a subdirectory
		subDir := filepath.Join(tmpDir, "subdir")
		err := os.Mkdir(subDir, 0755)
		require.NoError(t, err)

		// Create an old file in subdirectory
		oldFile := filepath.Join(subDir, "old_file.txt")
		err = os.WriteFile(oldFile, []byte("old content"), 0644)
		require.NoError(t, err)

		// Set modification time to 2 hours ago
		oldTime := time.Now().Add(-2 * time.Hour)
		err = os.Chtimes(oldFile, oldTime, oldTime)
		require.NoError(t, err)

		cleanup := NewCleanup8()
		err = cleanup.CleanupDirectory(tmpDir, 1*time.Hour)
		require.NoError(t, err)

		// Subdirectory should still exist
		_, err = os.Stat(subDir)
		assert.NoError(t, err, "subdirectory should still exist")

		// Old file in subdirectory should be removed (Walk is recursive)
		_, err = os.Stat(oldFile)
		assert.True(t, os.IsNotExist(err), "old file in subdirectory should be removed")
	})

	t.Run("removes multiple old files", func(t *testing.T) {
		tmpDir := t.TempDir()

		// Create multiple old files
		oldTime := time.Now().Add(-2 * time.Hour)
		for i := 0; i < 3; i++ {
			oldFile := filepath.Join(tmpDir, "old_file_"+string(rune('a'+i))+".txt")
			err := os.WriteFile(oldFile, []byte("old content"), 0644)
			require.NoError(t, err)
			err = os.Chtimes(oldFile, oldTime, oldTime)
			require.NoError(t, err)
		}

		cleanup := NewCleanup8()
		err := cleanup.CleanupDirectory(tmpDir, 1*time.Hour)
		require.NoError(t, err)

		// All old files should be removed
		entries, err := os.ReadDir(tmpDir)
		require.NoError(t, err)
		assert.Empty(t, entries)
	})
}
