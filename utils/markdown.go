package utils

import (
	"regexp"
	"strings"
)

// MediaItem struct holds the information for an extracted media link.
type MediaItem struct {
	URL string
	Alt string
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

func ExtractContentBlocks(sourceText string) []ContentBlock {
	var orderedBlocks []ContentBlock
	
	// FindAllStringSubmatchIndex 返回 [][]int，包含整个匹配和捕获组的索引
	// 索引顺序: [MatchStart MatchEnd Submatch1Start Submatch1End]
	allMatches := mediaRegex.FindAllStringSubmatchIndex(sourceText, -1)
	
	lastIndex := 0
	
	// 刷新文本块的辅助函数（保留所有原始字符）
	flushText := func(text string) {
		if text != "" {
			orderedBlocks = append(orderedBlocks, ContentBlock{
				Type:    "text",
				Content: text, // 保持所有原始字符 (\n, \t, <, > 等)
			})
		}
	}
	
	for _, match := range allMatches {
		// 整个匹配的起始和结束索引 (用于从纯文本中切除，即舍弃描述文字部分)
		matchStart := match[0]
		matchEnd := match[1]
		
		// 1. **提取前面的文本块**
		textBefore := sourceText[lastIndex:matchStart]
		flushText(textBefore)
		
		// 2. **提取干净的媒体 URL**
		// 捕获组 1 的起始和结束索引
		urlStart := match[2]
		urlEnd := match[3]
		
		var url string
		if urlStart != -1 && urlEnd != -1 {
			url = sourceText[urlStart:urlEnd]
		}
		
		// 3. **创建媒体块**
		if url != "" {
			mediaType := "image"
			if isVideoURL(url) {
				mediaType = "video"
			}
			
			orderedBlocks = append(orderedBlocks, ContentBlock{
				Type: mediaType,
				Media: MediaItem{
					URL: url,
					Alt: "", // 明确舍弃 Alt/描述文本
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
