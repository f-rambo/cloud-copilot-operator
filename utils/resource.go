package utils

import (
	"os"
)

// 判断文件是否存在
func CheckFileIsExist(filename string) bool {
	var exist = true
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		exist = false
	}
	return exist
}

// 创建一个可读可写的文件
func CreateFile(filename string) error {
	var file, err = os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return nil
}

// 创建一个嵌套的目录
func CreateDir(path string) error {
	var err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}
	return nil
}
