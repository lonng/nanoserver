package util

// 转换 bool 类型到 int 类型
// true 转换为 1, false 转换为 0
func Bool2Int(b bool) int {
	if b {
		return 1
	}
	return 0
}
