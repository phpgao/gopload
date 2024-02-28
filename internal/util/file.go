package util

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// CleanupOldFilesAndEmptyDirs 清理指定目录下的旧文件和空目录
func CleanupOldFilesAndEmptyDirs(dir string, age time.Duration) error {
	// 计算过期时间
	expiration := time.Now().Add(-age)

	// 删除过期文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if info.ModTime().Before(expiration) {
				fmt.Printf("Deleting old file: %s\n", path)
				return os.Remove(path)
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("error walking through files: %w", err)
	}

	// 删除空的子目录
	err = removeEmptyDirs(dir)
	if err != nil {
		return fmt.Errorf("error removing empty directories: %w", err)
	}

	return nil
}

// removeEmptyDirs 递归删除空的子目录
func removeEmptyDirs(dir string) error {
	dirsToDelete := []string{}
	// 清理空目录的非递归遍历
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != dir { // 跳过顶层目录
			empty, err := isDirEmpty(path)
			if err != nil {
				return err
			}
			if empty {
				dirsToDelete = append(dirsToDelete, path)
			}
		}
		return nil
	})

	// 对目录进行逆序操作, 确保子目录在父目录之前处理
	for i := len(dirsToDelete) - 1; i >= 0; i-- {
		fmt.Printf("Removing empty directory: %s\n", dirsToDelete[i])
		if err := os.Remove(dirsToDelete[i]); err != nil {
			return fmt.Errorf("error removing directory %s: %w", dirsToDelete[i], err)
		}
	}

	return err
}

// isDirEmpty 检查指定的目录是否为空
func isDirEmpty(dir string) (bool, error) {
	f, err := os.Open(dir)
	if err != nil {
		return false, err
	}
	defer f.Close()

	_, err = f.Readdirnames(1) // 尝试读取至少一个文件
	if err == io.EOF {
		return true, nil // 如果是EOF，则目录为空
	}
	return false, err
}

func RandStringBytes(n int) string {
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)
	var sb strings.Builder
	sb.Grow(n)
	for i := 0; i < n; i++ {
		sb.WriteByte(letterBytes[r.Int63()%int64(len(letterBytes))])
	}
	return sb.String()
}
