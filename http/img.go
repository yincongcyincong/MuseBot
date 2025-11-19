package http

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
	"github.com/yincongcyincong/MuseBot/param"
	"github.com/yincongcyincong/MuseBot/utils"
)

var (
	oneLayerDirPaths = make(map[string][]string)
	twoLayerDirPaths = make(map[string][][]string)
	validExtensions  = map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
)

func initImg() {
	conf.BaseConfInfo.ImageDay = -1
	err := loadAllImagePaths()
	if err != nil {
		logger.Error("Failed to complete image path loading process", "error", err)
	}
}

func loadTwoLayerPaths(imageType string, dirPath string) error {
	// 读取第一层子目录 (例如 CategoryA, CategoryB)
	categoryDirs, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logger.Error("Failed to read 2-layer base directory", "directory", dirPath, "error", err)
		return err
	}
	
	var allPathsForThisType [][]string // 用于存储所有类别的图片路径
	
	for _, catDir := range categoryDirs {
		if !catDir.IsDir() {
			continue // 跳过非目录文件，只处理子目录
		}
		
		categoryName := catDir.Name()
		currentCategoryPath := filepath.Join(dirPath, categoryName)
		
		// 读取第二层子目录中的文件
		files, err := ioutil.ReadDir(currentCategoryPath)
		if err != nil {
			logger.Error("Failed to read 2-layer category directory", "directory", currentCategoryPath, "error", err)
			continue
		}
		
		var pathsInThisCategory []string
		for _, file := range files {
			if !file.IsDir() {
				ext := strings.ToLower(filepath.Ext(file.Name()))
				if validExtensions[ext] {
					fullPath := filepath.Join(currentCategoryPath, file.Name())
					pathsInThisCategory = append(pathsInThisCategory, fullPath)
				}
			}
		}
		
		if len(pathsInThisCategory) > 0 {
			// 将当前类别下的所有图片路径作为一个切片存入
			allPathsForThisType = append(allPathsForThisType, pathsInThisCategory)
			logger.Info("Loaded 2-layer category paths", "type", imageType, "category", categoryName, "count", len(pathsInThisCategory))
		}
	}
	
	if len(allPathsForThisType) > 0 {
		twoLayerDirPaths[strings.ToLower(imageType)] = allPathsForThisType
	}
	
	return nil
}

// loadAllImagePaths 是主加载函数，负责区分一层和两层结构
func loadAllImagePaths() error {
	oneLayerDirPaths = make(map[string][]string)
	twoLayerDirPaths = make(map[string][][]string)
	
	dirs, err := ioutil.ReadDir(*conf.BaseConfInfo.ImagePath)
	if err != nil {
		logger.Error("Failed to read image directory", "directory", *conf.BaseConfInfo.ImagePath, "error", err)
		return err
	}
	
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue // Skip non-directory files
		}
		
		imageType := dir.Name() // 文件夹名作为 'type'
		currentDirPath := filepath.Join(*conf.BaseConfInfo.ImagePath, imageType)
		
		// 1. 读取当前目录内容以判断结构
		contents, err := ioutil.ReadDir(currentDirPath)
		if err != nil {
			logger.Error("Failed to read directory contents", "directory", currentDirPath, "error", err)
			continue
		}
		
		// 2. 检查目录中是否存在子目录，以判断是否为 2-Layer 结构
		isTwoLayer := false
		for _, content := range contents {
			if content.IsDir() {
				isTwoLayer = true
				break
			}
		}
		
		if isTwoLayer {
			err = loadTwoLayerPaths(imageType, currentDirPath)
			if err != nil {
				logger.Error("Failed to load 2-layer paths", "type", imageType, "error", err)
			}
			continue
		}
		
		// --- 1-Layer Structure (如果没有子目录，则按一层结构处理) ---
		var paths []string
		for _, file := range contents {
			if !file.IsDir() {
				ext := strings.ToLower(filepath.Ext(file.Name()))
				if validExtensions[ext] {
					fullPath := filepath.Join(currentDirPath, file.Name())
					paths = append(paths, fullPath)
				}
			}
		}
		
		if len(paths) > 0 {
			oneLayerDirPaths[strings.ToLower(imageType)] = paths
			logger.Info("Loaded 1-layer image paths", "type", imageType, "count", len(paths))
		}
	}
	
	// 最终总结日志
	logger.Info("Image Loading Completed",
		"one_layer_types_count", len(oneLayerDirPaths),
		"two_layer_types_count", len(twoLayerDirPaths))
	
	lestTwoLayerNum := -1
	for _, paths := range twoLayerDirPaths {
		if lestTwoLayerNum == -1 {
			lestTwoLayerNum = len(paths)
		}
		if len(paths) < lestTwoLayerNum {
			lestTwoLayerNum = len(paths)
		}
	}
	
	if lestTwoLayerNum > 0 {
		conf.BaseConfInfo.ImageDay = time.Now().YearDay() % lestTwoLayerNum
	}
	
	return nil
}

func imageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query()
	imageType := strings.ToLower(query.Get("type"))
	if imageType == "" {
		logger.Error("Missing 'type' query parameter in image request")
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, "")
		return
	}
	
	rStr := query.Get("rand")
	rInt := utils.ParseInt(rStr)
	
	var imagePath string
	if _, ok := twoLayerDirPaths[imageType]; ok {
		dayIdx := time.Now().YearDay() % len(twoLayerDirPaths[imageType])
		imgNum := len(twoLayerDirPaths[imageType][dayIdx])
		imagePath = twoLayerDirPaths[imageType][dayIdx][int(rInt)%imgNum]
	} else if _, ok := oneLayerDirPaths[imageType]; ok {
		imgNum := len(oneLayerDirPaths[imageType])
		imagePath = oneLayerDirPaths[imageType][int(rInt)%imgNum]
	}
	
	selectedFileName := filepath.Base(imagePath)
	
	file, err := os.Open(imagePath)
	if err != nil {
		logger.Error("Failed to open image file", "file", imagePath, "error", err)
		utils.Failure(ctx, w, r, param.CodeParamError, param.MsgParamError, "")
		return
	}
	defer file.Close()
	
	mimeType := getMimeType(selectedFileName)
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("X-Selected-File", selectedFileName) // Debug info
	
	_, err = io.Copy(w, file)
	if err != nil {
		logger.Error("Failed to copy image file to response", "file", imagePath, "error", err)
	}
}

// getMimeType infers the MIME Type based on the file extension.
func getMimeType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		// Default for unknown or non-image types
		return "application/octet-stream"
	}
}
