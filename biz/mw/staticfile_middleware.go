package mw

import (
	"errors"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

// StaticFileMiddleware 是一个 Gin 中间件，用于提供静态文件服务，
// 支持忽略 API 路径，并为 SPA 应用 fallback 到 index.html。
// staticFS 必须是一个 fs.FS，比如使用 embed.FS 或 os.DirFS("./static")
func StaticFileMiddleware(staticFS fs.FS) gin.HandlerFunc {
	return func(c *gin.Context) {
		filePath := c.Request.URL.Path

		// 定义要跳过的 API 路径前缀
		skipPrefixes := []string{
			"/api",
			"/nacos/v1/auth/login",
			"/nacos/v1/cs/configs",
		}

		// 如果是 API 请求，则直接放行
		for _, prefix := range skipPrefixes {
			if strings.HasPrefix(filePath, prefix) {
				c.Next()
				return
			}
		}

		// 默认请求根路径时，返回 index.html
		if filePath == "" || filePath == "/" {
			filePath = "/index.html"
		}

		// 去掉开头的斜杠，构造相对于 static/ 的路径
		relPath := strings.TrimPrefix(filePath, "/")
		fullPath := "static/" + relPath  // 比如 static/index.html
		indexPath := "static/index.html" // fallback 文件

		// 尝试返回请求的文件
		if served := serveFileFromFS(c, staticFS, fullPath); served {
			c.Abort()
			return
		}

		// 如果找不到，尝试返回 index.html（用于 SPA 前端路由）
		if served := serveFileFromFS(c, staticFS, indexPath); served {
			slog.Debugf("[STATIC] 文件 '%s' 不存在，fallback 到 index.html", fullPath)
			c.Abort()
			return
		}

		// 如果都找不到，返回 404
		slog.Debugf("[STATIC] 文件 '%s' 和 index.html 均不存在，返回 404", fullPath)
		c.String(http.StatusNotFound, "404 Not Found")
		c.Abort()
	}
}

// serveFileFromFS 从 fs.FS 中读取文件内容并作为 HTTP 响应返回
// 返回值表示是否成功服务了文件
func serveFileFromFS(c *gin.Context, filesystem fs.FS, path string) bool {
	file, err := filesystem.Open(path)
	if err != nil {
		if !errorsIsNotFound(err) {
			slog.Debugf("[STATIC] 打开文件 '%s' 失败: %v", path, err)
		}
		return false
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Debugf("[STATIC] 关闭文件 '%s' 失败: %v", path, err)
		}
	}()

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Debugf("[STATIC] 读取文件 '%s' 失败: %v", path, err)
		return false
	}

	// 根据文件扩展名设置 Content-Type
	ext := filepath.Ext(path)
	contentType := mime.TypeByExtension(ext)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Data(http.StatusOK, contentType, data)
	return true
}

// errorsIsNotFound 判断错误是否为 "文件不存在"
// 适配 fs.FS 的错误类型，比如 embed.FS 或 os.DirFS 在文件不存在时返回 fs.ErrNotExist
func errorsIsNotFound(err error) bool {
	return errors.Is(err, fs.ErrNotExist)
}
