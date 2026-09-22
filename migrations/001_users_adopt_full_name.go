package migrations

import "gorm.io/gorm"

var all = []Migration{
	{
		ID:  "001_users_adopt_full_name",
		Run: adoptUsersFullName,
	},
}

func adoptUsersFullName(tx *gorm.DB) error {
	hasLegacy, err := columnExists(tx, "users", "fullname")
	if err != nil {
		return err
	}
	if !hasLegacy {
		return nil
	}

	hasShared, err := columnExists(tx, "users", "full_name")
	if err != nil {
		return err
	}

	if hasShared {
		if err := tx.Exec(`
			UPDATE users SET full_name = fullname
			WHERE (full_name IS NULL OR full_name = '')
			  AND fullname IS NOT NULL AND fullname <> ''
		`).Error; err != nil {
			return err
		}
	}

	return tx.Exec(`ALTER TABLE users MODIFY IF EXISTS fullname LONGTEXT NULL`).Error
}
