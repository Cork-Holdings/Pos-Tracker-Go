package seeders

import (
	"errors"
	"fmt"
	"pos-master/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedAdminUser(db *gorm.DB) error {

	tx := db.Begin()

	var adminRole models.Role

	if err := tx.Where("name = ?", "admin").
		First(&adminRole).Error; err != nil {

		tx.Rollback()
		return errors.New("admin role not found, run role seeder first")
	}

	var existing models.User
	err := tx.Where("email = ?", "luswepo17@gmail.com").
		First(&existing).Error

	if err == nil {
		fmt.Println("Admin user already exists")
		tx.Rollback()
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte("Admin@123"), bcrypt.DefaultCost,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	admin := models.User{
		ID:       uuid.New(),
		FullName: "System",
		Email:    "luswepo17@gmail.com",
		Password: string(hashedPassword),
		RoleID:   adminRole.ID,
		Status:   models.UserStatusActive,
		Phone:    "",
		OtpCode:  "",
	}

	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	fmt.Println("Seeded admin user + wallet")

	return nil
}
