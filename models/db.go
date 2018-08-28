package models

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type DB struct {
	*gorm.DB
}

// OpenDB opens a connection to a Postgres database, returning the db instance
func OpenDB(dataSourceName string) (db *DB, err error) {
	gormDB, err := gorm.Open("postgres", dataSourceName)
	if err != nil {
		return
	}

	db = &DB{gormDB}

	if err = db.Migrate(); err != nil {
		return
	}

	return db, nil
}

// DropDB drops the databases. This is for TESTING only.
// In production code, this would not even be here as an option
func DropDB(dataSourceName string) (*DB, error) {
	gormDB, err := gorm.Open("postgres", dataSourceName)
	if err != nil {
		return nil, err
	}
	defer gormDB.Close()

	db := &DB{gormDB}

	db.Exec("DROP TABLE metrics CASCADE;")
	db.Exec("DROP TABLE tags CASCADE;")
	db.Exec("DROP TABLE metric_tags;")

	return db, nil
}

// Migrate sets up all the database tables and constraints if they aren't set up yet
func (db *DB) Migrate() (err error) {
	// Create tables
	if err = db.AutoMigrate(&Metric{}, &Tag{}, &MetricTags{}).Error; err != nil {
		return
	}
	// Add foreign key constaints
	if err = db.Model(MetricTags{}).AddForeignKey("metric_id", "metrics(id)", "CASCADE", "CASCADE").Error; err != nil {
		return
	}
	if err = db.Model(MetricTags{}).AddForeignKey("tag_id", "tags(id)", "RESTRICT", "RESTRICT").Error; err != nil {
		return
	}
	// this index is the opposite order from the primary key
	if err = db.Model(MetricTags{}).AddIndex("idx_metric_tags_tag_id_metric_id", "tag_id", "metric_id").Error; err != nil {
		return
	}
	return
}
