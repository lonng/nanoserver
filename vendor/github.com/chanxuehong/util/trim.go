package util

import (
	"bytes"
	"strings"
)

var newlineBytes = []byte{'\n'}

// 去掉 src 开头和结尾的空白, 如果 src 包括换行, 去掉换行和这个换行符两边的空白
//  NOTE: 根据 '\n' 来分行的, 某些系统或软件用 '\r' 来分行, 则不能正常工作.
func TrimSpace(src []byte) []byte {
	bytesArr := bytes.Split(src, newlineBytes)
	for i := 0; i < len(bytesArr); i++ {
		bytesArr[i] = bytes.TrimSpace(bytesArr[i])
	}
	return bytes.Join(bytesArr, nil)
}

// 去掉 src 开头和结尾的空白, 如果 src 包括换行, 去掉换行和这个换行符两边的空白
//  NOTE: 根据 '\n' 来分行的, 某些系统或软件用 '\r' 来分行, 则不能正常工作.
func TrimSpaceString(src string) string {
	strs := strings.Split(src, "\n")
	for i := 0; i < len(strs); i++ {
		strs[i] = strings.TrimSpace(strs[i])
	}
	return strings.Join(strs, "")
}
