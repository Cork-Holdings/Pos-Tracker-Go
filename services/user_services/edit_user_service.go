package userservices

import (
	"errors"
	"strings"

	database "pos-master/config"
	"pos-master/models"
	pb "pos-master/proto/auth"
	eventservices "pos-master/services/event_services"
	"pos-master/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func EditUser(req *pb.EditUserRequest) error {
	userID, err := uuid.Parse(strings.TrimSpace(req.Id))
	if err != nil {
		return utils.CapitalizeError("invalid ID format")
	}

	var currentUser models.User
	if err := database.DB.Where("id = ?", userID).First(&currentUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.CapitalizeError("unable to find user with that ID")
		}
		return utils.CapitalizeError("failed to look up user")
	}

	updates := map[string]interface{}{}

	if fullname := strings.TrimSpace(req.Fullname); fullname != "" && fullname != currentUser.FullName {
		updates["fullname"] = fullname
	}

	if email := normalizeEmail(req.Email); email != "" && email != currentUser.Email {
		updates["email"] = email
	}

	// An absent status must not overwrite the stored one: a blank status would
	// lock the account out, since login requires an exactly "active" status.
	if req.Status != "" {
		status := strings.ToLower(strings.TrimSpace(req.Status))
		if !models.IsValidUserStatus(status) {
			return utils.CapitalizeError("status must be either active or inactive")
		}
		if status != currentUser.Status {
			updates["status"] = status
		}
	}

	if roleName := strings.TrimSpace(req.Role); roleName != "" {
		var role models.Role
		if err := database.DB.Where("name = ?", roleName).First(&role).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return utils.CapitalizeError("role not found")
			}
			return utils.CapitalizeError("failed to look up role")
		}
		if role.ID != currentUser.RoleID {
			updates["role_id"] = role.ID
		}
	}

	if len(updates) == 0 {
		return nil
	}

	if err := database.DB.Model(&models.User{}).
		Where("id = ?", userID).
		Updates(updates).Error; err != nil {
		if isDuplicateKeyError(err) {
			return utils.CapitalizeError("a user with this email already exists")
		}
		return utils.CapitalizeError("failed to update user")
	}

	eventservices.RegisterEvent("User edited successfully", map[string]interface{}{
		"user id":   req.Id,
		"Full name": req.Fullname,
		"Email":     req.Email,
		"Role":      req.Role,
		"Status":    req.Status,
	})

	return nil
}
