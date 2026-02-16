package ginstatic

import (
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

// cacheEntry 缓存条目
type cacheEntry struct {
	Data       []byte      // 原始文件内容
	Gzipped    []byte      // Gzip 压缩后的内容
	ModTime    time.Time   // 文件修改时间
	Size       int64       // 文件大小
	ETag       string      // ETag 值
	LastAccess int64       // 最后访问时间（Unix 时间戳）
	Path       string      // 文件路径
	once       sync.Once   // 确保只加载一次
	loadErr    error       // 加载错误
}

// Cache 内存缓存
type Cache struct {
	mu            sync.RWMutex
	entries       sync.Map // map[string]*cacheEntry
	totalSize    int64     // 当前缓存总大小
	maxSize      int64     // 最大缓存大小
	maxFiles     int32     // 最大缓存文件数
	fileCount    int32     // 当前缓存文件数
	evictCounter uint64    // 淘汰计数器
	onEvict      func(string) // 淘汰回调
}

// NewCache 创建新的缓存实例
func NewCache(maxSize int64, maxFiles int, onEvict func(string)) *Cache {
	if maxSize <= 0 {
		maxSize = 100 * 1024 * 1024 // 默认 100MB
	}
	if maxFiles <= 0 {
		maxFiles = 500
	}
	return &Cache{
		maxSize:   maxSize,
		maxFiles:  int32(maxFiles),
		onEvict:   onEvict,
		entries:   sync.Map{},
		totalSize: 0,
		fileCount: 0,
	}
}

// Get 获取缓存条目
func (c *Cache) Get(key string) (*cacheEntry, bool) {
	entry, ok := c.entries.Load(key)
	if !ok {
		return nil, false
	}
	
	ce := entry.(*cacheEntry)
	// 更新最后访问时间
	atomic.StoreInt64(&ce.LastAccess, time.Now().UnixNano())
	return ce, true
}

// Set 设置缓存条目
func (c *Cache) Set(key string, entry *cacheEntry) {
	// 检查是否需要淘汰
	c.evictIfNeeded(entry.Size)

	// 检查缓存大小限制
	if c.totalSize+entry.Size > c.maxSize {
		// 尝试淘汰更多
		c.evictIfNeeded(entry.Size)
		if c.totalSize+entry.Size > c.maxSize {
			// 仍然超出限制，跳过缓存
			return
		}
	}

	// 存储条目
	oldEntry, loaded := c.entries.LoadOrStore(key, entry)
	if loaded {
		// 已存在，更新
		oldCe := oldEntry.(*cacheEntry)
		atomic.StoreInt64(&c.totalSize, atomic.LoadInt64(&c.totalSize)-oldCe.Size)
	}

	atomic.AddInt64(&c.totalSize, entry.Size)
	atomic.AddInt32(&c.fileCount, 1)
}

// Delete 删除缓存条目
func (c *Cache) Delete(key string) {
	entry, ok := c.entries.Load(key)
	if !ok {
		return
	}
	
	ce := entry.(*cacheEntry)
	c.entries.Delete(key)
	atomic.AddInt64(&c.totalSize, -ce.Size)
	atomic.AddInt32(&c.fileCount, -1)
	
	if c.onEvict != nil {
		c.onEvict(key)
	}
}

// evictIfNeeded 检查并执行淘汰
func (c *Cache) evictIfNeeded(newEntrySize int64) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查文件数限制
	for c.fileCount >= c.maxFiles && c.fileCount > 0 {
		c.evictOldest()
	}

	// 检查大小限制
	for c.totalSize+newEntrySize > c.maxSize && c.totalSize > 0 {
		if !c.evictOldest() {
			break
		}
	}
}

// evictOldest 淘汰最旧的条目
// 返回是否成功淘汰
func (c *Cache) evictOldest() bool {
	var oldestKey string
	var oldestEntry *cacheEntry
	var oldestTime int64 = ^int64(0)

	// 遍历所有条目找到最旧的
	c.entries.Range(func(key, value interface{}) bool {
		ce := value.(*cacheEntry)
		lastAccess := atomic.LoadInt64(&ce.LastAccess)
		if lastAccess < oldestTime {
			oldestTime = lastAccess
			oldestKey = key.(string)
			oldestEntry = ce
		}
		return true
	})

	if oldestKey == "" {
		return false
	}

	c.entries.Delete(oldestKey)
	c.totalSize -= oldestEntry.Size
	c.fileCount--
	c.evictCounter++

	if c.onEvict != nil {
		c.onEvict(oldestKey)
	}

	return true
}

// Clear 清空缓存
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries.Range(func(key, _ interface{}) bool {
		c.entries.Delete(key.(string))
		return true
	})
	c.totalSize = 0
	c.fileCount = 0
}

// Size 返回当前缓存大小
func (c *Cache) Size() int64 {
	return atomic.LoadInt64(&c.totalSize)
}

// FileCount 返回当前缓存文件数
func (c *Cache) FileCount() int {
	return int(atomic.LoadInt32(&c.fileCount))
}

// EvictCount 返回淘汰次数
func (c *Cache) EvictCount() uint64 {
	return atomic.LoadUint64(&c.evictCounter)
}

// LoadFile 加载文件到缓存
func (c *Cache) LoadFile(root, relPath string, enableGzip bool, gzipLevel int) (*cacheEntry, error) {
	// 检查缓存
	if entry, ok := c.Get(relPath); ok {
		return entry, nil
	}

	// 读取文件
	absPath := filepath.Join(root, relPath)
	data, err := os.ReadFile(absPath)
	if err != nil {
		return nil, err
	}

	// 获取文件信息
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	// 创建缓存条目
	entry := &cacheEntry{
		Data:    data,
		ModTime: info.ModTime(),
		Size:    info.Size(),
		ETag:    generateETag(absPath, info),
		Path:    absPath,
	}

	// 压缩（如启用）
	if enableGzip && len(data) >= 1024 {
		gzData, err := GzipCompress(data, gzipLevel)
		if err == nil {
			entry.Gzipped = gzData
		}
	}

	// 存储到缓存
	c.Set(relPath, entry)

	return entry, nil
}

// generateETag 生成 ETag
func generateETag(path string, info os.FileInfo) string {
	// 使用 inode、size 和 mtime 生成 ETag
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("%d-%d-%s", info.Size(), info.ModTime().UnixNano(), path)))
	return fmt.Sprintf(`"%x"`, h.Sum(nil))
}

// preloadDirectory 预加载目录下所有文件
func (c *Cache) preloadDirectory(root string, enableGzip bool, gzipLevel int) error {
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relPath, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		_, err = c.LoadFile(root, relPath, enableGzip, gzipLevel)
		return err
	})
}

// containsDotFile 检查路径是否包含点文件
func containsDotFile(path string) bool {
	parts := filepath.SplitList(path)
	for _, part := range parts {
		if len(part) > 0 && part[0] == '.' {
			return true
		}
	}
	return false
}
