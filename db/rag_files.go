package db

import (
	"time"
	
	"github.com/yincongcyincong/MuseBot/conf"
	"github.com/yincongcyincong/MuseBot/metrics"
)

type RagFiles struct {
	ID         int64  `json:"id"`
	FileName   string `json:"file_name"`
	FileMd5    string `json:"file_md5"`
	VectorId   string `json:"vector_id"`
	UpdateTime int64  `json:"update_time"`
	CreateTime int    `json:"create_time"`
	IsDeleted  int    `json:"is_deleted"`
}

func InsertRagFile(fileName, fileMd5 string) (int64, error) {
	// insert data
	insertSQL := `INSERT INTO rag_files (file_name, file_md5, create_time, update_time, vector_id, from_bot) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := DB.Exec(insertSQL, fileName, fileMd5, time.Now().Unix(), time.Now().Unix(), "", *conf.BaseConfInfo.BotName)
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
	querySQL := `SELECT id, file_name, file_md5, update_time, create_time FROM rag_files WHERE file_md5 = ? and is_deleted = 0 and from_bot = ?`
	rows, err := DB.Query(querySQL, fileMd5, *conf.BaseConfInfo.BotName)
	
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

func GetRagFileByFileName(fileName string) ([]*RagFiles, error) {
	querySQL := `SELECT id, file_name, file_md5, update_time, create_time, vector_id FROM rag_files WHERE file_name = ? and is_deleted = 0 and from_bot = ?`
	rows, err := DB.Query(querySQL, fileName, *conf.BaseConfInfo.BotName)
	
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var ragFiles []*RagFiles
	for rows.Next() {
		var ragFile RagFiles
		if err := rows.Scan(&ragFile.ID, &ragFile.FileName, &ragFile.FileMd5, &ragFile.UpdateTime, &ragFile.CreateTime, &ragFile.VectorId); err != nil {
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

func DeleteRagFileByFileName(fileName string) error {
	query := `UPDATE rag_files set is_deleted = 1 WHERE file_name = ? and from_bot = ?`
	_, err := DB.Exec(query, fileName, *conf.BaseConfInfo.BotName)
	return err
}

func DeleteRagFileByVectorId(fileName string) error {
	query := `UPDATE rag_files set is_deleted = 1 WHERE vector_id = ? and from_bot = ?`
	_, err := DB.Exec(query, fileName, *conf.BaseConfInfo.BotName)
	return err
}

func UpdateVectorIdByFileMd5(fileMd5, vectorId string) error {
	query := `UPDATE rag_files set vector_id = ? WHERE file_md5 = ? and from_bot = ?`
	_, err := DB.Exec(query, vectorId, fileMd5, *conf.BaseConfInfo.BotName)
	return err
}
