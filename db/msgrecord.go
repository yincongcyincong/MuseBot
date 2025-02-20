package db

import (
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

var MsgRecord = sync.Map{}

func InsertMsgRecord(username string, aq *AQ) {
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
		if len(msgRecord.AQs) >= MaxQAPair {
			msgRecord.AQs = msgRecord.AQs[1:]
		}
		msgRecord.updateTime = time.Now().Unix()
	}
	MsgRecord.Store(username, msgRecord)
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
}

func StarCheckUserLen() {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("StarCheckUserLen panic err:%v\n", err)
			}
		}()
		timer := time.NewTicker(time.Minute)
		for range timer.C {
			totalNum := 0
			timeUserPair := make(map[int64][]string)
			MsgRecord.Range(func(k, v interface{}) bool {
				msgRecord := v.(*MsgRecordInfo)
				if _, ok := timeUserPair[msgRecord.updateTime]; !ok {
					timeUserPair[msgRecord.updateTime] = make([]string, 0)
				}
				timeUserPair[msgRecord.updateTime] = append(timeUserPair[msgRecord.updateTime], k.(string))
				totalNum++
				return true
			})

			log.Printf("StarCheckUserLen totalNum:%d\n", totalNum)
			if totalNum < MaxUserLength {
				continue
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

	}()
}
