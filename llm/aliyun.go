package llm

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/param"
)

func GenerateAliyunImg(prompt string, imageContent []byte) ([]byte, int, error) {

	var (
		url     string
		payload map[string]interface{}
	)

	if imageContent != nil {
		url = "https://dashscope.aliyuncs.com/api/v1/services/aigc/multimodal-generation/generation"

		base64Img := "data:image/png;base64," + base64.StdEncoding.EncodeToString(imageContent)

		payload = map[string]interface{}{
			"model": "qwen-image-edit",
			"input": map[string]interface{}{
				"messages": []map[string]interface{}{
					{
						"role": "user",
						"content": []map[string]interface{}{
							{"image": base64Img}, // 本地图片 base64 形式
							{"text": prompt},     // 修改描述
						},
					},
				},
			},
			"parameters": map[string]interface{}{
				"negative_prompt": "",
				"watermark":       false,
			},
		}
	} else {
		// 没有图片 → 调用文生图接口
		url = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text2image/image-synthesis"

		payload = map[string]interface{}{
			"model": "qwen-image-plus",
			"input": map[string]interface{}{
				"prompt": prompt,
			},
			"parameters": map[string]interface{}{
				"size":          "1328*1328",
				"n":             1,
				"prompt_extend": true,
				"watermark":     true,
			},
		}
	}

	// 转 JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, 0, fmt.Errorf("序列化 JSON 失败: %w", err)
	}

	// 构造请求
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, 0, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+*conf.BaseConfInfo.AliyunToken)

	// 文生图需要异步 header
	if imageContent == nil {
		req.Header.Set("X-DashScope-Async", "enable")
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("读取响应失败: %w", err)
	}

	return body, param.ImageTokenUsage, nil
}
