package db

import (
	"time"
)

type Bot struct {
	ID         int    `json:"id"`
	Address    string `json:"address"`
	CrtFile    string `json:"crt_file"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
	IsDeleted  int    `json:"is_deleted"`
	Status     string `json:"status" db:"-"`
}

func CreateBot(address, crtFile string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`INSERT INTO bot (address, crt_file, create_time, update_time, is_deleted) VALUES (?, ?, ?, ?, 0)`,
		address, crtFile, now, now)
	return err
}

func GetBotByID(id int) (*Bot, error) {
	row := DB.QueryRow(`SELECT id, address, crt_file, create_time, update_time, is_deleted FROM bot WHERE id = ? AND is_deleted = 0`, id)
	b := &Bot{}
	err := row.Scan(&b.ID, &b.Address, &b.CrtFile, &b.CreateTime, &b.UpdateTime, &b.IsDeleted)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func UpdateBotAddress(id int, newAddress, crtFile string) error {
	now := time.Now().Unix()
	_, err := DB.Exec(`UPDATE bot SET address = ?, crt_file = ?, update_time = ? WHERE id = ?`, newAddress, crtFile, now, id)
	return err
}

func SoftDeleteBot(id int) error {
	_, err := DB.Exec(`UPDATE bot SET is_deleted = 1 WHERE id = ?`, id)
	return err
}

func ListBots(offset, limit int) ([]*Bot, int, error) {
	rows, err := DB.Query(`SELECT id, address, crt_file, create_time, update_time, is_deleted FROM bot WHERE is_deleted = 0 LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	
	bots := make([]*Bot, 0)
	for rows.Next() {
		var b Bot
		err := rows.Scan(&b.ID, &b.Address, &b.CrtFile, &b.CreateTime, &b.UpdateTime, &b.IsDeleted)
		if err != nil {
			return nil, 0, err
		}
		bots = append(bots, &b)
	}
	
	var total int
	err = DB.QueryRow(`SELECT COUNT(*) FROM bot WHERE is_deleted = 0`).Scan(&total)
	if err != nil {
		return nil, 0, err
	}
	
	return bots, total, nil
}
