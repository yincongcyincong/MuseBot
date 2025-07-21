package db

import (
	"database/sql"
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
	"time"
	
	"github.com/cohesion-org/deepseek-go"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
	"github.com/yincongcyincong/telegram-deepseek-bot/param"
)

const MaxQAPair = 10

type MsgRecordInfo struct {
	AQs        []*AQ
	updateTime int64
}

type AQ struct {
	Question string
	Answer   string
	Content  string
	Token    int
}

type Record struct {
	ID         int    `json:"id"`
	UserId     string `json:"user_id"`
	Question   string `json:"question"`
	Answer     string `json:"answer"`
	Content    string `json:"content"`
	Token      int    `json:"token"`
	IsDeleted  int    `json:"is_deleted"`
	CreateTime int64  `json:"create_time"`
	RecordType int    `json:"record_type"`
}

var MsgRecord = sync.Map{}

func InsertMsgRecord(userId string, aq *AQ, insertDB bool) {
	var msgRecord *MsgRecordInfo
	msgRecordInter, ok := MsgRecord.Load(userId)
	if !ok {
		msgRecord = &MsgRecordInfo{
			AQs:        []*AQ{aq},
			updateTime: time.Now().Unix(),
		}
	} else {
		msgRecord = msgRecordInter.(*MsgRecordInfo)
		msgRecord.AQs = append(msgRecord.AQs, aq)
		if len(msgRecord.AQs) > MaxQAPair {
			msgRecord.AQs = msgRecord.AQs[1:]
		}
		msgRecord.updateTime = time.Now().Unix()
	}
	MsgRecord.Store(userId, msgRecord)
	
	if insertDB {
		go InsertRecordInfo(&Record{
			UserId:     userId,
			Question:   aq.Question,
			Answer:     aq.Answer,
			Content:    aq.Content,
			Token:      aq.Token,
			RecordType: param.TextRecordType,
		})
	}
}

func GetMsgRecord(userId string) *MsgRecordInfo {
	msgRecord, ok := MsgRecord.Load(userId)
	if !ok {
		return nil
	}
	return msgRecord.(*MsgRecordInfo)
}

func DeleteMsgRecord(userId string) {
	MsgRecord.Delete(userId)
	err := DeleteRecord(userId)
	if err != nil {
		logger.Error("Error deleting record", "err", err)
	}
}

func UpdateUserTime() {
	InsertRecord()
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("StarCheckUserLen panic err", "err", err, "stack", string(debug.Stack()))
			}
		}()
		timer := time.NewTicker(time.Minute)
		for range timer.C {
			UpdateDBData()
		}
		
	}()
}

func UpdateDBData() {
	totalNum := 0
	timeUserPair := make(map[int64][]string)
	MsgRecord.Range(func(k, v interface{}) bool {
		msgRecord := v.(*MsgRecordInfo)
		if _, ok := timeUserPair[msgRecord.updateTime]; !ok {
			timeUserPair[msgRecord.updateTime] = make([]string, 0)
		}
		timeUserPair[msgRecord.updateTime] = append(timeUserPair[msgRecord.updateTime], k.(string))
		UpdateUserInfo(k.(string), msgRecord.updateTime)
		totalNum++
		return true
	})
}

func UpdateUserInfo(userId string, updateTime int64) {
	err := UpdateUserUpdateTime(userId, updateTime)
	if err != nil {
		logger.Error("StarCheckUserLen UpdateUserUpdateTime err", "err", err)
	}
}

func InsertRecord() {
	users, err := GetUsers()
	if err != nil {
		logger.Error("InsertRecord GetUsers err", "err", err)
	}
	
	for _, user := range users {
		records, err := getRecordsByUserId(user.UserId)
		if err != nil {
			logger.Error("InsertRecord GetUsers err", "err", err)
		}
		for i := len(records) - 1; i >= 0; i-- {
			record := records[i]
			InsertMsgRecord(user.UserId, &AQ{
				Question: record.Question,
				Answer:   record.Answer,
				Content:  record.Content,
			}, false)
			metrics.TotalRecords.Inc()
		}
	}
	
	metrics.TotalUsers.Add(float64(len(users)))
	
}

// getRecordsByUserId get latest 10 records by user_id
func getRecordsByUserId(userId string) ([]Record, error) {
	// construct SQL statements
	query := fmt.Sprintf("SELECT id, user_id, question, answer, content FROM records WHERE user_id =  ? " +
		"and is_deleted = 0 and record_type = 0 order by create_time desc limit 10")
	
	// execute query
	rows, err := DB.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.UserId, &record.Question, &record.Answer, &record.Content)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	
	return records, nil
}

// InsertRecordInfo insert record
func InsertRecordInfo(record *Record) {
	query := `INSERT INTO records (user_id, question, answer, content, token, create_time, is_deleted, record_type) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, record.UserId, record.Question, record.Answer, record.Content, record.Token, time.Now().Unix(), record.IsDeleted, record.RecordType)
	metrics.TotalRecords.Inc()
	if err != nil {
		logger.Error("insertRecord err", "err", err)
	}
	
	user, err := GetUserByID(record.UserId)
	if err != nil {
		logger.Error("Error get user by userid", "err", err)
	}
	
	if user == nil {
		_, err = InsertUser(record.UserId, deepseek.DeepSeekChat)
		if err != nil {
			logger.Error("Error insert user by userid", "err", err)
		}
	}
	
	err = UpdateUserToken(record.UserId, record.Token)
	if err != nil {
		logger.Error("Error update token by user", "err", err)
	}
}

// DeleteRecord delete record
func DeleteRecord(userId string) error {
	query := `UPDATE records set is_deleted = 1 WHERE user_id = ?`
	_, err := DB.Exec(query, userId)
	return err
}

func GetTokenByUserIdAndTime(userId string, start, end int64) (int, error) {
	querySQL := `SELECT sum(token) FROM records WHERE user_id = ? and create_time >= ? and create_time <= ?`
	row := DB.QueryRow(querySQL, userId, start, end)
	
	// scan row get result
	var user User
	err := row.Scan(&user.Token)
	if err != nil {
		if err == sql.ErrNoRows {
			// 如果没有找到数据，返回 nil
			return 0, nil
		}
		return 0, err
	}
	return user.Token, nil
}

func GetLastImageRecord(userId string, recordType int) (*Record, error) {
	query := fmt.Sprintf("SELECT id, user_id, question, answer, content FROM records WHERE user_id =  ? and record_type = ? and is_deleted = 0 order by id desc")
	
	// execute query
	rows, err := DB.Query(query, userId, recordType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.UserId, &record.Question, &record.Answer, &record.Content)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}
	
	if len(records) == 0 {
		return nil, nil
	}
	
	return &records[0], nil
}

func GetRecordCount(userId int64, isDeleted int) (int, error) {
	var count int
	query := "SELECT COUNT(*) FROM records"
	var args []interface{}
	var conditions []string
	
	if userId != 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userId)
	}
	
	if isDeleted >= 0 {
		conditions = append(conditions, "is_deleted = ?")
		args = append(args, isDeleted)
	}
	
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	
	err := DB.QueryRow(query, args...).Scan(&count)
	return count, err
}

func GetRecordList(userId int64, page int, pageSize int, isDeleted int) ([]Record, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize
	
	query := `
		SELECT id, user_id, question, answer, content, token, is_deleted, create_time
		FROM records`
	var args []interface{}
	var conditions []string
	
	if userId != 0 {
		conditions = append(conditions, "user_id = ?")
		args = append(args, userId)
	}
	if isDeleted >= 0 {
		conditions = append(conditions, "is_deleted = ?")
		args = append(args, isDeleted)
	}
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	
	query += " ORDER BY id DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)
	
	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var records []Record
	for rows.Next() {
		var r Record
		if err := rows.Scan(&r.ID, &r.UserId, &r.Question, &r.Answer, &r.Content, &r.Token, &r.IsDeleted, &r.CreateTime); err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}
