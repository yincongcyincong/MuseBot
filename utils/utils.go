package utils

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unicode/utf16"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/logger"
)

// Utf16len calculates the length of a string in UTF-16 code units.
func Utf16len(s string) int {
	utf16Str := utf16.Encode([]rune(s))
	return len(utf16Str)
}

func ParseInt(str string) int {
	num, _ := strconv.Atoi(str)
	return num
}

func ReplaceCommand(content string, command string, botName string) string {
	mention := "@" + botName
	
	content = strings.ReplaceAll(content, command, mention)
	content = strings.ReplaceAll(content, mention, "")
	prompt := strings.TrimSpace(content)
	
	return prompt
}

func MD5(input string) string {
	// 计算 MD5
	hash := md5.Sum([]byte(input))
	
	// 转换为 16 进制字符串
	md5Str := hex.EncodeToString(hash[:])
	return md5Str
}

func GetRobotProxyClient() *http.Client {
	transport := &http.Transport{}
	
	if *conf.BaseConfInfo.RobotProxy != "" {
		proxy, err := url.Parse(*conf.BaseConfInfo.RobotProxy)
		if err != nil {
			logger.Warn("parse proxy url fail", "err", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	
	return &http.Client{
		Transport: transport,
	}
}

func GetLLMProxyClient() *http.Client {
	transport := &http.Transport{}
	
	if *conf.BaseConfInfo.LLMProxy != "" {
		proxy, err := url.Parse(*conf.BaseConfInfo.LLMProxy)
		if err != nil {
			logger.Warn("parse proxy url fail", "err", err)
		}
		transport.Proxy = http.ProxyURL(proxy)
	}
	
	return &http.Client{
		Transport: transport,
		Timeout:   5 * time.Minute, // 设置超时
	}
}

func FileRecognize(audioContent []byte) (string, error) {
	
	client := BuildAsrClient()
	client.Appid = *conf.AudioConfInfo.VolAudioAppID
	client.Token = *conf.AudioConfInfo.VolAudioToken
	client.Cluster = *conf.AudioConfInfo.VolAudioRecCluster
	
	asrResponse, err := client.RequestAsr(audioContent)
	if err != nil {
		logger.Error("fail to request asr ", "err", err)
		return "", err
	}
	
	if len(asrResponse.Results) == 0 {
		logger.Error("fail to request asr", "results", asrResponse.Results)
		return "", errors.New("fail to request asr")
	}
	
	return asrResponse.Results[0].Text, nil
	
}

func FileToMd5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hash := md5.New()
	
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	
	md5sum := fmt.Sprintf("%x", hash.Sum(nil))
	return md5sum, nil
}

func SetStructFieldByJSONTag(s interface{}, key string, value interface{}) error {
	val := reflect.ValueOf(s)
	if val.Kind() != reflect.Ptr || val.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("want a struct，get a %T", s)
	}
	
	elem := val.Elem()
	typ := elem.Type()
	
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := elem.Field(i)
		
		jsonTag := field.Tag.Get("json")
		
		if jsonTag == key {
			if !fieldValue.CanSet() {
				return fmt.Errorf("field %s can't be set", field.Name)
			}
			
			setValue := reflect.ValueOf(value)
			if fieldValue.Kind() == reflect.Ptr {
				elemType := fieldValue.Type().Elem()
				
				// 检查传入的值是否可转换为元素类型
				if !setValue.Type().ConvertibleTo(elemType) {
					return fmt.Errorf("can't put %T change to field %s with %s type", value, field.Name, elemType)
				}
				
				newValue := reflect.New(elemType)
				newValue.Elem().Set(setValue.Convert(elemType))
				
				fieldValue.Set(newValue)
				return nil
			} else {
				if setValue.Type().ConvertibleTo(fieldValue.Type()) {
					fieldValue.Set(setValue.Convert(fieldValue.Type()))
					return nil
				} else {
					return fmt.Errorf("can't put %T change to field %s with %s type", value, field.Name, fieldValue.Type())
				}
			}
		}
	}
	
	return fmt.Errorf("can't find with key '%s' matched JSON field", key)
}

func ValueToString(value interface{}) string {
	if value == nil {
		return "nil" // Handle nil values
	}
	
	// Get the reflect.Value and reflect.Type of the input
	v := reflect.ValueOf(value)
	t := reflect.TypeOf(value)
	
	// Handle pointer types: if it's a pointer, get the element it points to
	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return "nil"
		}
		v = v.Elem() // Get the value the pointer points to
		t = v.Type() // Update type to the actual type
	}
	
	switch t.Kind() {
	case reflect.String:
		return v.String() // Return string directly
	
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10) // Convert integer to decimal string
	
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10) // Convert unsigned integer to decimal string
	
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), 'f', -1, 64) // Format float as string
		// 'f' means decimal format, -1 uses necessary precision, 64 for float64
	
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()) // Convert boolean to string
	
	case reflect.Array, reflect.Slice:
		// For arrays and slices, convert each element to a string and join with commas
		elements := make([]string, v.Len())
		for i := 0; i < v.Len(); i++ {
			elements[i] = ValueToString(v.Index(i).Interface()) // Recursive call for nested types
		}
		return strings.Join(elements, ",")
	
	case reflect.Map:
		// For maps, convert each key-value pair to string
		keys := v.MapKeys()
		parts := make([]string, len(keys))
		for i, key := range keys {
			val := v.MapIndex(key)
			parts[i] = fmt.Sprintf("%s: %s", ValueToString(key.Interface()), ValueToString(val.Interface()))
		}
		return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
	
	case reflect.Struct:
		// For structs, convert each field to a string
		parts := make([]string, v.NumField())
		for i := 0; i < v.NumField(); i++ {
			field := v.Field(i)
			fieldName := t.Field(i).Name
			parts[i] = fmt.Sprintf("%s: %s", fieldName, ValueToString(field.Interface()))
		}
		return fmt.Sprintf("{%s}", strings.Join(parts, ", "))
	
	case reflect.Invalid:
		return "<invalid value>" // Handle invalid values such as uninitialized reflect.Value
	
	default:
		// For other types like Channel, Func, Interface, Ptr (already handled), UnsafePointer, etc.
		// Use fmt.Sprintf to get the default string representation
		return fmt.Sprintf("%v", value)
	}
}

func MapKeysToString(input interface{}) string {
	val := reflect.ValueOf(input)
	
	// Check if it's a map
	if val.Kind() != reflect.Map {
		return ""
	}
	
	keys := val.MapKeys()
	keyStrs := make([]string, len(keys))
	
	for i, key := range keys {
		// Convert key to string using fmt.Sprint (handles int, string, etc.)
		keyStrs[i] = fmt.Sprint(key.Interface())
	}
	
	return strings.Join(keyStrs, ",")
}

func DownloadFile(url string) ([]byte, error) {
	if url == "" {
		return nil, errors.New("url is empty")
	}
	
	client := GetRobotProxyClient()
	
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	// 检查 HTTP 状态码
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to download file: " + resp.Status)
	}
	
	// 读取响应体内容
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	return data, nil
}

func DetectAudioFormat(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}
	
	switch {
	case bytes.HasPrefix(data, []byte("OggS")):
		return "ogg"
	case bytes.HasPrefix(data, []byte("ID3")) || (data[0] == 0xFF && (data[1]&0xE0) == 0xE0):
		return "mp3"
	case bytes.HasPrefix(data, []byte("RIFF")) && string(data[8:12]) == "WAVE":
		return "wav"
	case bytes.HasPrefix(data, []byte("fLaC")):
		return "flac"
	case bytes.HasPrefix(data[4:], []byte("ftyp")):
		return "mp4"
	case len(data) >= 4 && data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3:
		return "webm"
	case bytes.HasPrefix(data, []byte("#!AMR")):
		return "amr"
	case bytes.HasPrefix(data, []byte("#!AMR-WB")):
		return "amr-wb"
	default:
		return "unknown"
	}
}

func DetectImageFormat(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}
	
	switch {
	case bytes.HasPrefix(data, []byte{0xFF, 0xD8, 0xFF}):
		return "jpeg"
	case bytes.HasPrefix(data, []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}):
		return "png"
	case bytes.HasPrefix(data, []byte("GIF87a")) || bytes.HasPrefix(data, []byte("GIF89a")):
		return "gif"
	case bytes.HasPrefix(data, []byte{0x42, 0x4D}):
		return "bmp"
	case bytes.HasPrefix(data, []byte("RIFF")) && bytes.HasPrefix(data[8:], []byte("WEBP")):
		return "webp"
	default:
		return "unknown"
	}
}

func ConvertToPNGFile(imageContent []byte) (*os.File, error) {
	contentType := DetectImageFormat(imageContent)
	var img image.Image
	var err error
	
	if contentType == "png" {
		tmpFile, err := os.CreateTemp("", "image_*.png")
		if err != nil {
			return nil, fmt.Errorf("create tmp file fail: %v", err)
		}
		if _, err := tmpFile.Write(imageContent); err != nil {
			tmpFile.Close()
			return nil, fmt.Errorf("write tmp file fail: %v", err)
		}
		tmpFile.Seek(0, io.SeekStart)
		return tmpFile, nil
	}
	
	img, _, err = image.Decode(bytes.NewReader(imageContent))
	if err != nil {
		return nil, fmt.Errorf("image decode fail: %v", err)
	}
	
	tmpFile, err := os.CreateTemp("", "image_*.png")
	if err != nil {
		return nil, fmt.Errorf("create tmp file fail: %v", err)
	}
	
	if err := png.Encode(tmpFile, img); err != nil {
		tmpFile.Close()
		return nil, fmt.Errorf("png encode fail: %v", err)
	}
	
	tmpFile.Seek(0, io.SeekStart)
	return tmpFile, nil
}

func DetectVideoMimeType(data []byte) string {
	if len(data) < 12 {
		return "unknown"
	}
	
	if string(data[4:8]) == "ftyp" {
		return "mp4"
	}
	
	if data[0] == 0x1A && data[1] == 0x45 && data[2] == 0xDF && data[3] == 0xA3 {
		return "webm"
	}
	
	if string(data[:4]) == "OggS" {
		return "ogg"
	}
	
	// fallback
	return "unknown"
}

func RandomFilename(ext string) string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	
	return fmt.Sprintf("%d_%s.%s", time.Now().UnixNano(), hex.EncodeToString(b), ext)
}
