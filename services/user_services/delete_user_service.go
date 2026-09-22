package userservices

import (
	"errors"

	database "pos-master/config"
	"pos-master/models"
	eventservices "pos-master/services/event_services"
	"pos-master/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DeleteUser removes the user identified by userId. requesterId is the ID of the
// authenticated caller, taken from the JWT rather than the request body.
func DeleteUser(userId string, requesterId string) error {
	parsedID, err := uuid.Parse(userId)
	if err != nil {
		return utils.CapitalizeError("invalid ID format")
	}

	if requesterID, err := uuid.Parse(requesterId); err == nil && requesterID == parsedID {
		return utils.CapitalizeError("you cannot delete your own account")
	}

	var user models.User
	if err := database.DB.Preload("Role").Where("id = ?", parsedID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.CapitalizeError("unable to find user with that ID")
		}
		return utils.CapitalizeError("failed to look up user")
	}

	if user.Role.Name == models.RoleAdmin {
		var admins int64
		if err := database.DB.Model(&models.User{}).
			Where("role_id = ?", user.RoleID).
			Count(&admins).Error; err != nil {
			return utils.CapitalizeError("failed to count administrators")
		}

		if admins <= 1 {
			return utils.CapitalizeError("cannot delete the last administrator")
		}
	}

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Unscoped().Delete(&models.User{}, "id = ?", parsedID).Error; err != nil {
		tx.Rollback()
		return utils.CapitalizeError("failed to delete user")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.CapitalizeError("failed to commit changes")
	}

	eventservices.RegisterEvent("User deleted successfully", map[string]interface{}{
		"User ID": userId,
		"Email":   user.Email,
	})
	return nil
}
