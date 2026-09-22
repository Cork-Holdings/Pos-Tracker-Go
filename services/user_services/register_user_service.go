package userservices

import (
	"errors"
	"fmt"
	"strings"

	database "pos-master/config"
	"pos-master/models"
	pb "pos-master/proto/auth"
	emailservices "pos-master/services/emailservices"
	eventservices "pos-master/services/event_services"
	"pos-master/utils"

	"github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(req *pb.RegisterRequest) *pb.AuthResponse {
	fullname := strings.TrimSpace(req.Fullname)
	email := normalizeEmail(req.Email)
	roleName := strings.TrimSpace(req.Role)

	if fullname == "" {
		utils.Warn("create user failed", "reason", "fullname is required")
		return failureResponse("fullname is required")
	}
	if email == "" {
		utils.Warn("create user failed", "reason", "email is required")
		return failureResponse("email is required")
	}
	if strings.TrimSpace(req.Password) == "" {
		utils.Warn("create user failed", "reason", "password is required", "email", email)
		return failureResponse("password is required")
	}
	if roleName == "" {
		utils.Warn("create user failed", "reason", "role is required", "email", email)
		return failureResponse("role is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.Error("failed to hash password", "email", email, "error", err.Error())
		return failureResponse("Failed to hash password")
	}

	var role models.Role
	result := database.DB.Where("name = ?", roleName).First(&role)
	if result.Error != nil {
		utils.Error("role not found for user create", "role", roleName, "email", email, "error", result.Error.Error())
		return failureResponse("Role not found")
	}

	userID := uuid.New()
	user := models.User{
		ID:       userID,
		FullName: fullname,
		Email:    email,
		Password: string(hashedPassword),
		RoleID:   role.ID,
		Status:   models.UserStatusActive,
		Phone:    "",
		OtpCode:  "",
	}

	tx := database.DB.Begin()
	if tx.Error != nil {
		utils.Error("unable to start user create transaction", "error", tx.Error.Error())
		return failureResponse(fmt.Sprintf("Unable to start transaction: %v", tx.Error))
	}

	result = tx.Create(&user)
	if result.Error != nil {
		tx.Rollback()
		if isDuplicateKeyError(result.Error) {
			utils.Warn("create user failed", "reason", "duplicate email", "email", email)
			return failureResponse("A user with this email already exists")
		}
		utils.Error("unable to register user", "email", email, "error", result.Error.Error())
		return failureResponse(fmt.Sprintf("Unable to register user: %v", result.Error))
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		utils.Error("failed to commit user create transaction", "email", email, "error", err.Error())
		return failureResponse(fmt.Sprintf("Failed to commit transaction: %v", err))
	}

	go func() {
		if err := emailservices.SendWelcomeEmail(email, fullname); err != nil {
			utils.Error("failed to send welcome email", "email", email, "error", err.Error())
		}
	}()

	utils.Info("user registered",
		"user_id", userID.String(),
		"email", email,
		"role", roleName,
	)

	eventservices.RegisterEvent("User registered successfully", map[string]interface{}{
		"User ID":  userID,
		"Fullname": fullname,
		"Email":    email,
		"Role":     roleName,
	})

	return &pb.AuthResponse{
		Success: true,
		Message: "Registration successful",
		Status:  "success",
	}
}

func failureResponse(message string) *pb.AuthResponse {
	return &pb.AuthResponse{
		Success: false,
		Message: message,
		Status:  "failure",
	}
}

func isDuplicateKeyError(err error) bool {
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "duplicate")
}

// normalizeEmail keeps stored addresses in one canonical form, since login
// matches on an exact string comparison.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
