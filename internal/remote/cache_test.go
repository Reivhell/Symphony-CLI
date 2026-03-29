package remote

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"github.com/user/repo", "github.com_user_repo"},
		{"github.com/user/repo@v1.2.0", "github.com_user_repo_v1.2.0"},
		{"https://example.com/template.zip", "https_example.com_template.zip"},
		{"c:/Users/someone/Desktop/prj", "c_Users_someone_Desktop_prj"}, // safe win path
	}

	for _, tc := range tests {
		actual := CacheKey(tc.input)
		assert.Equal(t, tc.expected, actual)
	}
}

func TestCacheFlow(t *testing.T) {
	cacheDir := t.TempDir()

	entry := CacheEntry{
		Source:    "github.com/user/myrepo",
		LocalPath: filepath.Join(cacheDir, "github.com_user_myrepo"),
		CachedAt:  time.Now(),
		SizeBytes: 1024,
		IsTagged:  false,
	}

	// 1. Initially Empty
	list, err := List(cacheDir)
	assert.NoError(t, err)
	assert.Empty(t, list)
	assert.False(t, IsCached(entry.Source, cacheDir))

	// 2. Write Metadata
	err = WriteMeta(entry.Source, cacheDir, entry)
	require.NoError(t, err)

	// 3. Cached (within TTL)
	assert.True(t, IsCached(entry.Source, cacheDir))

	// 4. List Cache
	list, err = List(cacheDir)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, entry.Source, list[0].Source)

	// 5. Invalidate
	err = Invalidate(entry.Source, cacheDir)
	require.NoError(t, err)

	assert.False(t, IsCached(entry.Source, cacheDir))
}

func TestCacheTTL(t *testing.T) {
	cacheDir := t.TempDir()

	entry := CacheEntry{
		Source:    "expired_repo",
		CachedAt:  time.Now().Add(-48 * time.Hour), // 2 hari lalu
		IsTagged:  false, // Regular branch => expired setelah 24h
	}
	require.NoError(t, WriteMeta(entry.Source, cacheDir, entry))
	assert.False(t, IsCached(entry.Source, cacheDir)) // Harus false karena TTL habis

	taggedEntry := CacheEntry{
		Source:    "tagged_repo@v1",
		CachedAt:  time.Now().Add(-48 * time.Hour),
		IsTagged:  true, // Tagged release => immutable, unexpired
	}
	require.NoError(t, WriteMeta(taggedEntry.Source, cacheDir, taggedEntry))
	assert.True(t, IsCached(taggedEntry.Source, cacheDir)) // Harus True
}
