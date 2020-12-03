package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jinzhu/gorm"
	_ "github.com/lib/pq"
)

type database struct {
	db *gorm.DB
}

// DB is the object
var DB database

func initDB() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		rdsHost := os.Getenv("RDS_HOSTNAME")
		rdsPort := os.Getenv("RDS_PORT")
		rdsDbName := os.Getenv("RDS_DB_NAME")
		rdsUsername := os.Getenv("RDS_USERNAME")
		rdsPassword := os.Getenv("RDS_PASSWORD")
		dbURL = fmt.Sprintf("postgres://%s:%s@%s:%s/%s", rdsUsername, rdsPassword, rdsHost, rdsPort, rdsDbName)
	}
	if dbURL == "" {
		panic("DB env vars not found")
	}
	var err error
	db, err := gorm.Open("postgres", dbURL)
	db.LogMode(true)
	if err != nil {
		panic(err)
	}
	DB.db = db
	fmt.Println("db connected")
}

func (db database) getMyMedia(pubKey string) []Media {
	ms := []Media{}
	db.db.Where("owner_pub_key = ?", pubKey).Find(&ms)
	return ms
}

func (db database) getMediaWithDimensions() []Media {
	ms := []Media{}
	db.db.Where("width > 0 and height > 0").Find(&ms)
	return ms
}

func (db database) getAllMedia() []Media {
	ms := []Media{}
	db.db.Find(&ms)
	return ms
}

func (db database) getTemplates() []Media {
	ms := []Media{}
	db.db.Where("template = ?", true).Find(&ms)
	return ms
}

func (db database) getMediaWithDimensionsByMuid(muid string) Media {
	m := Media{}
	db.db.Where("width > 0 and height > 0 and id = ?", muid).First(&m)
	return m
}

func (db database) getTemplateByMuid(muid string) Media {
	m := Media{}
	db.db.Where("template = ? and id = ?", true, muid).First(&m)
	return m
}

func (db database) getMyMediaByMUID(pubKey, muid string) Media {
	m := Media{}
	db.db.Where("owner_pub_key = ? and id = ?", pubKey, muid).First(&m)
	return m
}

func (db database) getMediaByMUID(muid string) Media {
	m := Media{}
	db.db.Where("id = ?", muid).First(&m)
	return m
}

func (db database) mediaPurchase(pubKey, muid string) Media {
	if muid == "" {
		return Media{}
	}
	row := db.db.Raw(`UPDATE media SET 
	total_buys = total_buys + 1,
	total_sats = total_sats + price
	WHERE (id = ? and owner_pub_key = ?)
	RETURNING id,price,ttl`, muid, pubKey).Row()
	m := Media{}
	row.Scan(&m.ID, &m.Price, &m.TTL)
	return m // if unsuccessful will return empty Media
}

func (db database) updateMedia(muid string, u map[string]interface{}) bool {
	if muid == "" {
		return false
	}
	db.db.Model(&Media{ID: muid}).Where("muid = ?", muid).Updates(u)
	return true
}

func (db database) getOwner(muid string) string {
	m := Media{}
	db.db.Select("owner_pub_key").Where("id = ?", muid).First(&m)
	return m.OwnerPubKey
}

var updatables = []string{
	"name", "description", "price", "ttl", "tags", "nonce",
}

// check that update owner_pub_key does in fact throw error
func (db database) createMedia(m Media) (Media, error) {
	if m.OwnerPubKey == "" {
		return Media{}, errors.New("no pub key")
	}
	onConflict := "ON CONFLICT (id) DO UPDATE SET"
	for i, u := range updatables {
		onConflict = onConflict + fmt.Sprintf(" %s=EXCLUDED.%s", u, u)
		if i < len(updatables)-1 {
			onConflict = onConflict + ","
		}
	}
	if m.Name == "" {
		m.Name = "name"
	}
	if m.Description == "" {
		m.Description = "description"
	}
	if m.Tags == nil {
		m.Tags = []string{}
	}
	if err := db.db.Set("gorm:insert_option", onConflict).Create(&m).Error; err != nil {
		fmt.Println(err)
		return Media{}, err
	}
	// not working?
	db.db.Exec(`UPDATE media SET tsv =
  	setweight(to_tsvector(name), 'A') ||
	setweight(to_tsvector(description), 'B') ||
	setweight(array_to_tsvector(tags), 'C')
	WHERE id = '` + m.ID + "'")
	return m, nil
}

func (db database) searchMedia(s string) []Media {
	ms := []Media{}
	if s == "" {
		return ms
	}
	// set limit
	db.db.Raw(
		`SELECT id, owner_pub_key, name, description, price, ttl, filename, mime, size, ts_rank(tsv, q) as rank
		FROM media, to_tsquery('` + s + `') q
		WHERE tsv @@ q
		ORDER BY rank DESC LIMIT 12;`).Find(&ms)
	return ms
}
