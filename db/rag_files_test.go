package db

import (
	"testing"
	"time"
	
	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestInsertAndGetRagFile(t *testing.T) {
	
	fileName := "test.txt"
	fileMd5 := "abc123"
	
	// 插入
	id, err := InsertRagFile(fileName, fileMd5)
	assert.NoError(t, err)
	assert.NotZero(t, id)
	
	// 按 md5 查询
	files, err := GetRagFileByFileMd5(fileMd5)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	assert.Equal(t, fileName, files[0].FileName)
	
	// 按 fileName 查询
	files2, err := GetRagFileByFileName(fileName)
	assert.NoError(t, err)
	assert.Len(t, files2, 1)
	assert.Equal(t, fileMd5, files2[0].FileMd5)
}

func TestUpdateAndDeleteRagFile(t *testing.T) {
	fileName := "test.txt"
	fileMd5 := "abc123"
	
	// 插入
	_, err := InsertRagFile(fileName, fileMd5)
	assert.NoError(t, err)
	
	// 更新 vector_id
	newVector := "vec-1"
	err = UpdateVectorIdByFileMd5(fileMd5, newVector)
	assert.NoError(t, err)
	
	files, err := GetRagFileByFileName(fileName)
	assert.NoError(t, err)
	assert.Equal(t, newVector, files[0].VectorId)
	
	// 删除 by fileName
	err = DeleteRagFileByFileName(fileName)
	assert.NoError(t, err)
	
	files, err = GetRagFileByFileName(fileName)
	assert.NoError(t, err)
	assert.Len(t, files, 0)
	
	// 再插入一个
	_, err = InsertRagFile("b.txt", "def456")
	assert.NoError(t, err)
	
	// 删除 by vectorId
	err = UpdateVectorIdByFileMd5("def456", "vec-2")
	assert.NoError(t, err)
	
	err = DeleteRagFileByVectorId("vec-2")
	assert.NoError(t, err)
	
	files, err = GetRagFileByFileName("b.txt")
	assert.NoError(t, err)
	assert.Len(t, files, 0)
}

func TestInsertTimeStamps(t *testing.T) {
	fileName := "time.txt"
	fileMd5 := "time123"
	
	// 插入
	_, err := InsertRagFile(fileName, fileMd5)
	assert.NoError(t, err)
	
	files, err := GetRagFileByFileMd5(fileMd5)
	assert.NoError(t, err)
	assert.Len(t, files, 1)
	
	now := time.Now().Unix()
	assert.LessOrEqual(t, files[0].UpdateTime, now)
}
