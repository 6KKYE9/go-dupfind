package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestFindDuplicates(t *testing.T) {
	dir := t.TempDir()
	// 两个内容相同、一个不同
	a := filepath.Join(dir, "a.txt")
	b := filepath.Join(dir, "b.txt")
	c := filepath.Join(dir, "c.txt")
	os.WriteFile(a, []byte("hello"), 0644)
	os.WriteFile(b, []byte("hello"), 0644)
	os.WriteFile(c, []byte("world"), 0644)

	// 直接验证 sha256File 一致性
	ha, _ := sha256File(a)
	hb, _ := sha256File(b)
	hc, _ := sha256File(c)
	if ha != hb {
		t.Fatal("相同内容应得到相同哈希")
	}
	if ha == hc {
		t.Fatal("不同内容应得到不同哈希")
	}

	// 全量分组
	groups := groupByHash([]string{a, b, c})
	// 应有两条：hello 组含 a,b；world 组仅 c
	if len(groups[ha]) != 2 {
		t.Fatalf("hello 组应有 2 个文件, got %d", len(groups[ha]))
	}
	if len(groups[hc]) != 1 {
		t.Fatalf("world 组应有 1 个文件, got %d", len(groups[hc]))
	}
	sort.Strings(groups[ha])
	if groups[ha][0] != a || groups[ha][1] != b {
		t.Fatalf("hello 组成员应为 a,b, got %v", groups[ha])
	}
}
