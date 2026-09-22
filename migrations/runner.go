package migrations

import (
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID  string
	Run func(tx *gorm.DB) error
}

type schemaMigration struct {
	ID        string    `gorm:"primaryKey;size:128"`
	AppliedAt time.Time `gorm:"not null"`
}

func (schemaMigration) TableName() string {
	return "schema_migrations"
}

func Run(db *gorm.DB) error {
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	for _, m := range all {
		var existing schemaMigration
		err := db.Where("id = ?", m.ID).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("check migration %s: %w", m.ID, err)
		}

		log.Printf("applying migration %s", m.ID)

		tx := db.Begin()
		if tx.Error != nil {
			return fmt.Errorf("begin migration %s: %w", m.ID, tx.Error)
		}

		if err := m.Run(tx); err != nil {
			tx.Rollback()
			return fmt.Errorf("migration %s: %w", m.ID, err)
		}

		record := schemaMigration{ID: m.ID, AppliedAt: time.Now()}
		if err := tx.Create(&record).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("record migration %s: %w", m.ID, err)
		}

		if err := tx.Commit().Error; err != nil {
			return fmt.Errorf("commit migration %s: %w", m.ID, err)
		}

		log.Printf("applied migration %s", m.ID)
	}

	return nil
}

func columnExists(tx *gorm.DB, table, column string) (bool, error) {
	var n int64
	err := tx.Raw(`
		SELECT COUNT(*) FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?
	`, table, column).Scan(&n).Error
	return n > 0, err
}
