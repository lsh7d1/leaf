package dao

import (
	"context"
	"leaf/dal"
	"testing"
)

const (
	mysqldsn = "root:root1234@tcp(127.0.0.1:13306)/db2?charset=utf8mb4&parseTime=True"
)

var _ = mysqldsn

func TestSegmentRepoImpl(t *testing.T) {
	repo := NewSegmentRepoImpl(dal.ConnectSQLite("../test.db").Debug())

	// 创建一个leaf
	//leaf, err := repo.CreateAndGetLeaf(context.Background(), "身份证号", 10000)
	//if err != nil {
	//	t.Logf("CreateAndGetLeaf failed, err: %v\n", err)
	//}
	//t.Logf("create a leaf: %#v\n", leaf)
	//
	//// 通过tag得到一个leaf
	//leaf, err = repo.GetLeafByTag(context.Background(), "身份证号")
	//if err != nil {
	//	t.Logf("GetLeafByTag failed, err: %v\n", err)
	//}
	//t.Logf("get a leaf: %#v\n", leaf)
	//
	//// 得到所有leaf
	//leafs, err := repo.GetAllLeafs(context.Background())
	//if err != nil {
	//	t.Logf("GetAllLeafs failed, err: %v\n", err)
	//}
	//t.Logf("get all leafs: %#v\n", leafs)
	//
	//// 得到所有tag
	//tags, err := repo.ListAllTags(context.Background())
	//if err != nil {
	//	t.Logf("ListAllTags failed, err: %v\n", err)
	//}
	//t.Logf("list all leafs: %#v\n", tags)
	//
	//// 用tag更新leaf，【预期失败】
	//_, err = repo.UpdateAndGetLeaf(context.Background(), "学生证号")
	//if err != nil {
	//	t.Logf("UpdateAndGetLeaf failed, err: %v\n", err)
	//}
	//
	//// 用tag更新leaf
	//leaf, err = repo.UpdateAndGetLeaf(context.Background(), "aaaa")
	//if err != nil {
	//	t.Logf("UpdateAndGetLeaf failed, err: %v\n", err)
	//}
	//t.Logf("updated maxid, leaf: %#v\n", leaf)

	// 用tag + step更新leaf
	leaf, err := repo.UpdateAndGetLeafWithStep(context.Background(), "aaaa", 999)
	if err != nil {
		t.Logf("UpdateAndGetLeafWithStep failed, err: %v\n", err)
	}
	t.Logf("updated maxid with step, leaf: %#v\n", leaf)
}
