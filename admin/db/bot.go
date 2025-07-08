package db

import (
	"database/sql"
	"time"
)

type Bot struct {
	ID         int    `json:"id"`
	Address    string `json:"address"`
	KeyFile    string `json:"key_file"`
	CaFile     string `json:"ca_file"`
	CrtFile    string `json:"crt_file"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	IsDeleted  int    `json:"is_deleted"`
	Status     string `json:"status" db:"-"`
}

func CreateBot(address, crtFile, secretFile, caFile string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`INSERT INTO bot (address, key_file, crt_file, ca_file, create_time, update_time, is_deleted) VALUES (?, ?, ?, ?, ?, ?, 0)`,
		address, crtFile, secretFile, caFile, now, now)
	return err
}

func GetBotByID(id int) (*Bot, error) {
	row := DB.QueryRow(`SELECT id, address, key_file, crt_file, ca_file, create_time, update_time, is_deleted FROM bot WHERE id = ? AND is_deleted = 0`, id)
	b := &Bot{}
	err := row.Scan(&b.ID, &b.Address, &b.KeyFile, &b.CrtFile, &b.CaFile, &b.CreateTime, &b.UpdateTime, &b.IsDeleted)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func UpdateBotAddress(id int, newAddress, crtFile, secretFile, caFile string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`UPDATE bot SET address = ?, crt_file = ?, key_file = ?, ca_file = ?, update_time = ? WHERE id = ?`, newAddress, crtFile, secretFile, caFile, now, id)
	return err
}

func SoftDeleteBot(id int) error {
	_, err := DB.Exec(`UPDATE bot SET is_deleted = 1 WHERE id = ?`, id)
	return err
}

func ListBots(offset, limit int, address string) ([]*Bot, int, error) {
	var (
		rows  *sql.Rows
		err   error
		args  []interface{}
		query string
	)
	
	bots := make([]*Bot, 0)
	
	if address != "" {
		query = `SELECT id, address, crt_file, key_file, ca_file, create_time, update_time, is_deleted
		         FROM bot
		         WHERE is_deleted = 0 AND address LIKE ?
		         LIMIT ? OFFSET ?`
		args = append(args, "%"+address+"%", limit, offset)
	} else {
		query = `SELECT id, address, crt_file, key_file, ca_file, create_time, update_time, is_deleted
		         FROM bot
		         WHERE is_deleted = 0
		         LIMIT ? OFFSET ?`
		args = append(args, limit, offset)
	}
	
	rows, err = DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var b Bot
		if err := rows.Scan(&b.ID, &b.Address, &b.CrtFile, &b.KeyFile, &b.CaFile, &b.CreateTime, &b.UpdateTime, &b.IsDeleted); err != nil {
			return nil, 0, err
		}
		bots = append(bots, &b)
	}
	
	var total int
	if address != "" {
		err = DB.QueryRow(`SELECT COUNT(*) FROM bot WHERE is_deleted = 0 AND address LIKE ?`, "%"+address+"%").Scan(&total)
	} else {
		err = DB.QueryRow(`SELECT COUNT(*) FROM bot WHERE is_deleted = 0`).Scan(&total)
	}
	if err != nil {
		return nil, 0, err
	}
	
	return bots, total, nil
}
