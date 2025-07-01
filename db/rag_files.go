package db

import (
	"time"
	
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
)

type RagFiles struct {
	ID         int64  `json:"id"`
	FileName   string `json:"file_name"`
	FileMd5    string `json:"file_md5"`
	UpdateTime int64  `json:"update_time"`
	CreateTime int    `json:"create_time"`
	IsDeleted  int    `json:"is_deleted"`
}

func InsertRagFile(fileName, fileMd5 string) (int64, error) {
	// insert data
	insertSQL := `INSERT INTO rag_files (file_name, file_md5, create_time, update_time) VALUES (?, ?, ?, ?)`
	result, err := DB.Exec(insertSQL, fileName, fileMd5, time.Now().Unix(), time.Now().Unix())
	if err != nil {
		return 0, err
	}
	
	// get last insert id
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	metrics.TotalUsers.Inc()
	return id, nil
}

func GetRagFileByFileMd5(fileMd5 string) ([]*RagFiles, error) {
	querySQL := `SELECT id, file_name, file_md5, update_time, create_time FROM rag_files WHERE file_md5 = ? and is_deleted = 0`
	rows, err := DB.Query(querySQL, fileMd5)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ragFiles []*RagFiles
	for rows.Next() {
		var ragFile RagFiles
		if err := rows.Scan(&ragFile.ID, &ragFile.FileName, &ragFile.FileMd5, &ragFile.UpdateTime, &ragFile.CreateTime); err != nil {
			return nil, err
		}
		ragFiles = append(ragFiles, &ragFile)
	}
	
	// check error
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ragFiles, nil
}

func DeleteRagFileByFileName(FileName string) error {
	query := `UPDATE rag_files set is_deleted = 1 WHERE file_name = ?`
	_, err := DB.Exec(query, FileName)
	return err
}
