package utils

import (
	"context"
	"encoding/base64"
	"regexp"
	"strings"
	
	"github.com/yincongcyincong/MuseBot/logger"
)

// MediaItem struct holds the information for an extracted media link.
type MediaItem struct {
	URL     string
	Content []byte
}

type ContentBlock struct {
	Type    string    // "text" or "image" or "video"
	Content string    // For Type="text"
	Media   MediaItem // For Type="media"
}

// 假设 isVideoURL 是一个外部函数
func isVideoURL(url string) bool {
	// 简单的判断逻辑，可根据实际需求调整
	return strings.HasSuffix(url, ".mp4") || strings.HasSuffix(url, ".webm") || strings.HasSuffix(url, ".mov")
}

// mediaRegex 仅匹配 Markdown 图像格式：![描述文字](URL)
// 捕获组 1 (.+?) 用于提取干净的 URL。
var mediaRegex = regexp.MustCompile(
	`!\[.*?\]\((.+?)\)`,
)

func ExtractContentBlocks(ctx context.Context, sourceText string) []ContentBlock {
	var orderedBlocks []ContentBlock
	allMatches := mediaRegex.FindAllStringSubmatchIndex(sourceText, -1)
	
	lastIndex := 0
	
	// 刷新文本块的辅助函数（保留所有原始字符）
	flushText := func(text string) {
		if text != "" {
			orderedBlocks = append(orderedBlocks, ContentBlock{
				Type:    "text",
				Content: text,
			})
		}
	}
	
	for _, match := range allMatches {
		matchStart := match[0]
		matchEnd := match[1]
		
		textBefore := sourceText[lastIndex:matchStart]
		flushText(textBefore)
		
		mediaStart := match[2]
		mediaEnd := match[3]
		
		var media string
		if mediaStart != -1 && mediaEnd != -1 {
			media = sourceText[mediaStart:mediaEnd]
		}
		
		if strings.HasPrefix(media, "http") {
			mediaType := "image"
			if isVideoURL(media) {
				mediaType = "video"
			}
			
			content, err := DownloadFile(media)
			if err != nil {
				logger.ErrorCtx(ctx, "download file error", "err", err)
				continue
			}
			
			orderedBlocks = append(orderedBlocks, ContentBlock{
				Type: mediaType,
				Media: MediaItem{
					URL:     media,
					Content: content,
				},
			})
		} else {
			tmp := strings.Split(media, "base64,")
			if len(tmp) > 1 {
				media = tmp[1]
			}
			
			mediaContent, err := base64.StdEncoding.DecodeString(media)
			if err != nil {
				logger.ErrorCtx(ctx, "decode base64 error", "err", err)
				continue
			}
			
			mediaType := "image"
			if DetectVideoMimeType(mediaContent) != "unknown" || strings.HasSuffix(media, "data:video") {
				mediaType = "video"
			}
			
			orderedBlocks = append(orderedBlocks, ContentBlock{
				Type: mediaType,
				Media: MediaItem{
					URL:     media,
					Content: mediaContent,
				},
			})
		}
		
		// 更新起始位置到整个匹配的末尾，从而跳过描述文字和 Markdown 语法
		lastIndex = matchEnd
	}
	
	// 4. **提取末尾剩余的文本块**
	textAfter := sourceText[lastIndex:]
	flushText(textAfter)
	
	return orderedBlocks
}
