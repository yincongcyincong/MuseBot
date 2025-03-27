package db

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	"github.com/cohesion-org/deepseek-go"
	"github.com/yincongcyincong/telegram-deepseek-bot/logger"
	"github.com/yincongcyincong/telegram-deepseek-bot/metrics"
)

const MaxQAPair = 10

type MsgRecordInfo struct {
	AQs        []*AQ
	updateTime int64
}

type AQ struct {
	Question string
	Answer   string
	Token    int
}

type Record struct {
	ID        int
	UserId    int64
	Question  string
	Answer    string
	Token     int
	IsDeleted int
}

var MsgRecord = sync.Map{}

func InsertMsgRecord(userId int64, aq *AQ, insertDB bool) {
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
			UserId:   userId,
			Question: aq.Question,
			Answer:   aq.Answer,
			Token:    aq.Token,
		})
	}
}

func GetMsgRecord(userId int64) *MsgRecordInfo {
	msgRecord, ok := MsgRecord.Load(userId)
	if !ok {
		return nil
	}
	return msgRecord.(*MsgRecordInfo)
}

func DeleteMsgRecord(userId int64) {
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
				logger.Error("StarCheckUserLen panic err", "err", err)
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
	timeUserPair := make(map[int64][]int64)
	MsgRecord.Range(func(k, v interface{}) bool {
		msgRecord := v.(*MsgRecordInfo)
		if _, ok := timeUserPair[msgRecord.updateTime]; !ok {
			timeUserPair[msgRecord.updateTime] = make([]int64, 0)
		}
		timeUserPair[msgRecord.updateTime] = append(timeUserPair[msgRecord.updateTime], k.(int64))
		UpdateUserInfo(k.(int64), msgRecord.updateTime)
		totalNum++
		return true
	})
}

func UpdateUserInfo(userId int64, updateTime int64) {
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
			}, false)
			metrics.TotalRecords.Inc()
		}
	}

	metrics.TotalUsers.Add(float64(len(users)))

}

// getRecordsByUserId get latest 10 records by user_id
func getRecordsByUserId(userId int64) ([]Record, error) {
	// construct SQL statements
	query := fmt.Sprintf("SELECT id, user_id, question, answer FROM records WHERE user_id =  ? and is_deleted = 0 order by create_time desc limit 10")

	// execute query
	rows, err := DB.Query(query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.UserId, &record.Question, &record.Answer)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// InsertRecordInfo insert record
func InsertRecordInfo(record *Record) {
	query := `INSERT INTO records (user_id, question, answer, token, create_time, is_deleted) VALUES (?, ?, ?, ?, ?, ?)`
	_, err := DB.Exec(query, record.UserId, record.Question, record.Answer, record.Token, time.Now().Unix(), record.IsDeleted)
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
func DeleteRecord(userId int64) error {
	query := `UPDATE records set is_deleted = 1 WHERE user_id = ?`
	_, err := DB.Exec(query, userId)
	return err
}

func GetTokenByUserIdAndTime(userId int64, start, end int64) (int, error) {
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
