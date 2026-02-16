package ginstatic

import (
	"bytes"
	"compress/gzip"
	"io"
	"sync"

	"github.com/klauspost/compress/zstd"
)

// gzWriterPool Gzip 写入器池
var gzWriterPool = sync.Pool{
	New: func() interface{} {
		w, _ := gzip.NewWriterLevel(nil, gzip.BestSpeed)
		return w
	},
}

// gzReaderPool Gzip 读取器池
var gzReaderPool = sync.Pool{
	New: func() interface{} {
		return new(gzip.Reader)
	},
}

// zstdEncoderPool Zstd 编码器池
var zstdEncoderPool = sync.Pool{
	New: func() interface{} {
		enc, _ := zstd.NewWriter(nil, zstd.WithEncoderLevel(zstd.SpeedFastest))
		return enc
	},
}

// GzipCompress 使用 Gzip 压缩数据
// level: 压缩级别 1-9
func GzipCompress(data []byte, level int) ([]byte, error) {
	if level < 1 {
		level = gzip.BestSpeed
	}
	if level > 9 {
		level = gzip.BestCompression
	}

	var buf bytes.Buffer
	writer, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	_, err = writer.Write(data)
	if err != nil {
		writer.Close()
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GzipDecompress 解压 Gzip 数据
func GzipDecompress(data []byte) ([]byte, error) {
	reader := gzReaderPool.Get().(*gzip.Reader)
	defer gzReaderPool.Put(reader)

	err := reader.Reset(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	return io.ReadAll(reader)
}

// ZstdCompress 使用 Zstd 压缩数据
func ZstdCompress(data []byte) ([]byte, error) {
	encoder := zstdEncoderPool.Get().(*zstd.Encoder)
	defer zstdEncoderPool.Put(encoder)

	encoder.Reset(bytes.NewBuffer(nil))
	defer encoder.Close()

	return encoder.EncodeAll(data, nil), nil
}

// ZstdDecompress 解压 Zstd 数据
func ZstdDecompress(data []byte) ([]byte, error) {
	dec, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer dec.Close()

	return io.ReadAll(dec)
}

// GetCompressedData 获取压缩后的数据
// 根据客户端 Accept-Encoding 返回最佳压缩版本
func GetCompressedData(data, gzData, zstdData []byte, acceptEncoding string) ([]byte, string, bool) {
	// 优先使用 Zstd（性能更好）
	if zstdData != nil && containsEncoding(acceptEncoding, "zstd") {
		return zstdData, "zstd", true
	}

	// 其次使用 Gzip
	if gzData != nil && containsEncoding(acceptEncoding, "gzip") {
		return gzData, "gzip", true
	}

	// 返回原始数据
	return data, "", false
}

// containsEncoding 检查是否包含指定的编码
func containsEncoding(acceptEncoding, encoding string) bool {
	if acceptEncoding == "" {
		return false
	}

	// 简单检查，实际应该更严格
	encoding = encoding + ","
	encodingQ := encoding + ";q="

	for _, part := range splitComma(acceptEncoding) {
		part = stripWhitespace(part)
		if part == encoding {
			return true
		}
		if len(part) > len(encodingQ) && part[:len(encodingQ)] == encodingQ {
			// 检查 q 值
			q := part[len(encodingQ):]
			if len(q) > 0 && q[0] > '0' {
				return true
			}
		}
	}

	return false
}

// splitComma 按逗号分割
func splitComma(s string) []string {
	result := make([]string, 0)
	var current []byte

	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if len(current) > 0 {
				result = append(result, string(current))
				current = nil
			}
		} else {
			current = append(current, s[i])
		}
	}

	if len(current) > 0 {
		result = append(result, string(current))
	}

	return result
}

// stripWhitespace 去除空白字符
func stripWhitespace(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		if s[i] > ' ' {
			result = append(result, s[i])
		}
	}
	return string(result)
}

// GetGzipLevel 获取压缩级别
func GetGzipLevel(level int) int {
	if level < 1 || level > 9 {
		return gzip.BestSpeed
	}
	return level
}

// NewZstdWriter 创建 Zstd 写入器
func NewZstdWriter(w io.Writer) *zstd.Encoder {
	enc, _ := zstd.NewWriter(w, zstd.WithEncoderLevel(zstd.SpeedFastest))
	return enc
}
