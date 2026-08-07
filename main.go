// go-dupfind：按内容（SHA-256）查找重复文件（仅用标准库）。
// 用法：go run . <目录> [-min N] [-del]
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func main() {
	minSize := flag.Int64("min", 1, "只比较大于等于该字节数的文件")
	del := flag.Bool("del", false, "交互式删除重复项（保留每组第一个）")
	flag.Parse()

	roots := flag.Args()
	if len(roots) == 0 {
		roots = []string{"."}
	}

	// 先按大小分组，快速排除不同大小的文件
	sizeGroups := make(map[int64][]string)
	err := walkAll(roots, *minSize, func(path string, info os.FileInfo) {
		sizeGroups[info.Size()] = append(sizeGroups[info.Size()], path)
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "遍历出错:", err)
		os.Exit(1)
	}

	hashGroups := make(map[string][]string)
	for _, paths := range sizeGroups {
		if len(paths) < 2 {
			continue
		}
		mergeGroups(hashGroups, groupByHash(paths))
	}

	found := 0
	for _, group := range hashGroups {
		if len(group) < 2 {
			continue
		}
		found++
		fmt.Printf("重复组 #%d（%d 个文件）:\n", found, len(group))
		keep := group[0]
		for _, p := range group {
			mark := " "
			if p == keep {
				mark = "*"
			}
			fmt.Printf("  [%s] %s\n", mark, p)
		}
		if *del {
			for _, p := range group[1:] {
				if askDelete(p) {
					if err := os.Remove(p); err != nil {
						fmt.Fprintf(os.Stderr, "删除失败 %s: %v\n", p, err)
					} else {
						fmt.Printf("    已删除: %s\n", p)
					}
				}
			}
		}
	}
	if found == 0 {
		fmt.Println("未发现重复文件。")
	}
}

func walkAll(roots []string, minSize int64, fn func(path string, info os.FileInfo)) error {
	for _, root := range roots {
		err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil // 跳过无法访问的条目
			}
			if info.IsDir() || info.Size() < minSize {
				return nil
			}
			fn(path, info)
			return nil
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func sha256File(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// groupByHash 计算一组文件的哈希，按哈希分组返回路径列表。
func groupByHash(paths []string) map[string][]string {
	out := make(map[string][]string)
	for _, p := range paths {
		h, err := sha256File(p)
		if err != nil {
			fmt.Fprintf(os.Stderr, "计算哈希失败 %s: %v\n", p, err)
			continue
		}
		out[h] = append(out[h], p)
	}
	return out
}

// mergeGroups 把一个哈希分组合并进目标 map。
func mergeGroups(dst, src map[string][]string) {
	for k, v := range src {
		dst[k] = append(dst[k], v...)
	}
}

func askDelete(path string) bool {
	fmt.Printf("    删除 %s ? [y/N] ", path)
	var ans string
	fmt.Scanln(&ans)
	return ans == "y" || ans == "Y"
}
