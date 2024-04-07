package dal

import (
	"fmt"
	"leaf/dal/model"
	"leaf/dal/query"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var _ = sqlite.Open("../test.db")

func ConnectSQLite(table string) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(table), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("sqlite.Open(table) failed, err: %v", err))
	}
	query.SetDefault(db)
	db.AutoMigrate(&model.LeafAlloc{})
	return db.Debug()
}

func ConnectMySQL(dsn string) *gorm.DB {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Errorf("mysql.Open(dsn) failed, err: %v", err))
	}
	query.SetDefault(db)
	db.AutoMigrate(&model.LeafAlloc{})
	return db
}
