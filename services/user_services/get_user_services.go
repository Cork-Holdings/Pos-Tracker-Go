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
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func GetUsers(req *pb.GetUsersRequest) (*pb.GetUsersResponse, error) {

	var users []models.User

	query := database.DB.Preload("Role").Model(&models.User{})

	if search := strings.TrimSpace(req.SearchQuery); search != "" {
		pattern := "%" + search + "%"
		query = query.Where("fullname LIKE ? OR email LIKE ?", pattern, pattern)
	}

	var totalUsers int64
	err := query.Count(&totalUsers).Error

	if err != nil {
		return nil, utils.CapitalizeError("failed to count users")
	}

	totalPages := int32((totalUsers + int64(req.PageSize) - 1) / int64(req.PageSize))
	// Calculate offset for pagination
	offset := (req.Page - 1) * req.PageSize

	// Execute the final query with pagination and preloading
	err = query.Order("created_at DESC").
		Limit(int(req.PageSize)).
		Offset(int(offset)).
		Find(&users).Error

	if err != nil {
		return nil, utils.CapitalizeError("failed to retrieve users")
	}

	pbUsers := make([]*pb.User, len(users))

	for i, user := range users {

		pbUsers[i] = &pb.User{
			Id:       user.ID.String(),
			Fullname: user.FullName,
			Email:    user.Email,
			Role:     user.Role.Name,
			Status:   user.Status,
		}
	}

	return &pb.GetUsersResponse{
		User:        pbUsers,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
		HasMore:     req.Page < totalPages,
		Count:       int32(totalUsers),
	}, nil
}

func GetUser(userID string) (*pb.User, error) {

	var user models.User

	userid, err := uuid.Parse(strings.TrimSpace(userID))

	if err != nil {
		return nil, utils.CapitalizeError("invalid ID format")
	}

	err = database.DB.Preload("Role").Where("id = ?", userid).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, utils.CapitalizeError("unable to find user with that ID")
		}
		return nil, utils.CapitalizeError("failed to look up user")
	}

	return &pb.User{
		Email:    user.Email,
		Id:       user.ID.String(),
		Fullname: user.FullName,
		Role:     user.Role.Name,
		Status:   user.Status,
	}, nil
}

// ChangeEmailorPassword updates the credentials of a single account. The caller
// must pass the ID of the authenticated user; it is never read from the request
// body, and the current password is always verified before anything changes.
func ChangeEmailorPassword(req *pb.ChangeEmailOrPasswordRequest) error {

	userID, err := uuid.Parse(strings.TrimSpace(req.UserId))
	if err != nil {
		return utils.CapitalizeError("invalid uuid. Check ID being sent")
	}

	if !req.IsEmailRequest && !req.IsPasswordRequest {
		return utils.CapitalizeError("nothing to update")
	}

	var user models.User

	err = database.DB.Where("id = ?", userID).First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.CapitalizeError("unable to find user")
		}
		return utils.CapitalizeError("unable to find user")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		return utils.CapitalizeError("your current password is incorrect")
	}

	updates := map[string]interface{}{}

	if req.IsEmailRequest {
		if normalizeEmail(req.OldEmail) != user.Email {
			return utils.CapitalizeError("Old email does not match our records")
		}

		newEmail := normalizeEmail(req.NewEmail)
		if newEmail == "" {
			return utils.CapitalizeError("new email is required")
		}
		if newEmail == user.Email {
			return utils.CapitalizeError("the new email matches your current one")
		}

		updates["email"] = newEmail
	}

	if req.IsPasswordRequest {
		if len(req.NewPassword) < 8 {
			return utils.CapitalizeError("your new password must be at least 8 characters")
		}

		if req.ConfirmPassword != req.NewPassword {
			return utils.CapitalizeError("your passwords do not match")
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return utils.CapitalizeError("unable to hash password")
		}

		updates["password"] = string(hashedPassword)
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		return utils.CapitalizeError("unable to start transaction")
	}

	if err := tx.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		tx.Rollback()
		if isDuplicateKeyError(err) {
			return utils.CapitalizeError("a user with this email already exists")
		}
		return utils.CapitalizeError("unable to update your account")
	}

	if err := tx.Commit().Error; err != nil {
		return utils.CapitalizeError("failed to commit changes")
	}

	if req.IsEmailRequest {
		eventservices.RegisterEvent("User changed their email", map[string]interface{}{
			"User ID": user.ID.String(),
		})
	}
	if req.IsPasswordRequest {
		eventservices.RegisterEvent("User changed their password", map[string]interface{}{
			"User ID": user.ID.String(),
		})
	}

	return nil
}
