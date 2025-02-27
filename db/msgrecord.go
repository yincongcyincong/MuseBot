package db

import (
	"fmt"
	"github.com/cohesion-org/deepseek-go"
	"log"
	"sort"
	"sync"
	"time"
)

const MaxUserLength = 1000
const MaxQAPair = 10

type MsgRecordInfo struct {
	AQs        []*AQ
	updateTime int64
}

type AQ struct {
	Question string
	Answer   string
}

type Record struct {
	ID       int
	Name     string
	Question string
	Answer   string
}

var MsgRecord = sync.Map{}

func InsertMsgRecord(username string, aq *AQ, insertDB bool) {
	var msgRecord *MsgRecordInfo
	msgRecordInter, ok := MsgRecord.Load(username)
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
	MsgRecord.Store(username, msgRecord)

	if insertDB {
		go insertRecord(&Record{
			Name:     username,
			Question: aq.Question,
			Answer:   aq.Answer,
		})
	}
}

func GetMsgRecord(username string) *MsgRecordInfo {
	msgRecord, ok := MsgRecord.Load(username)
	if !ok {
		return nil
	}
	return msgRecord.(*MsgRecordInfo)
}

func DeleteMsgRecord(username string) {
	MsgRecord.Delete(username)
	err := DeleteRecord(username)
	if err != nil {
		log.Printf("Error deleting record: %v \n", err)
	}
}

func StarCheckUserLen() {
	InsertRecord()

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("StarCheckUserLen panic err:%v\n", err)
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

	log.Printf("StarCheckUserLen totalNum:%d\n", totalNum)
	if totalNum < MaxUserLength {
		return
	}

	log.Println("start cleaning...")
	times := make([]int64, 0)
	for t := range timeUserPair {
		times = append(times, t)
	}
	sort.Slice(times, func(i, j int) bool {
		return times[i] > times[j]
	})

	for _, t := range times {
		for _, user := range timeUserPair[t] {
			MsgRecord.Delete(user)
			totalNum--
			if totalNum <= MaxUserLength {
				continue
			}
		}
	}
}

func UpdateUserInfo(username string, updateTime int64) {
	user, err := GetUserByName(username)
	if err != nil {
		log.Printf("Error get user by name: %v \n", err)
	}

	if user == nil {
		_, err = InsertUser(username, deepseek.DeepSeekChat)
		if err != nil {
			log.Printf("Error get user by name: %v \n", err)
		}
	}

	err = UpdateUserUpdateTime(username, updateTime)
	if err != nil {
		log.Printf("StarCheckUserLen UpdateUserUpdateTime err:%v\n", err)
	}
}

func InsertRecord() {
	users, err := GetUsers()
	if err != nil {
		log.Printf("InsertRecord GetUsers err:%v\n", err)
	}

	for _, user := range users {
		records, err := getRecordsByName(user.Name)
		if err != nil {
			log.Printf("InsertRecord GetUsers err:%v\n", err)
		}
		for _, record := range records {
			InsertMsgRecord(user.Name, &AQ{
				Question: record.Question,
				Answer:   record.Answer,
			}, false)
		}
	}

}

// getRecordsByName get latest 10 records by name
func getRecordsByName(name string) ([]Record, error) {
	// construct SQL statements
	query := fmt.Sprintf("SELECT id, name, question, answer FROM records WHERE name =  ? limit 10")

	// execute query
	rows, err := DB.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []Record
	for rows.Next() {
		var record Record
		err := rows.Scan(&record.ID, &record.Name, &record.Question, &record.Answer)
		if err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}

// insertRecord insert record
func insertRecord(record *Record) {
	query := `INSERT INTO records (name, question, answer) VALUES (?, ?, ?)`
	_, err := DB.Exec(query, record.Name, record.Question, record.Answer)
	if err != nil {
		log.Printf("insertRecord err:%v\n", err)
	}
}

// delete record
func DeleteRecord(name string) error {
	query := `DELETE FROM records WHERE name = ?`
	_, err := DB.Exec(query, name)
	return err
}
