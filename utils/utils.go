package utils

// ConvertToInterfaceSlice 辅助函数，将 []string 转换为 []interface{}
func ConvertToInterfaceSlice(strs []string) []interface{} {
	result := make([]interface{}, len(strs))
	for i, v := range strs {
		result[i] = v
	}
	return result
}
