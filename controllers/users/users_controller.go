package users

import (
	"fmt"
	"log/slog"
	"strconv"

	"pos-master/models"
	"pos-master/proto/auth"
	userservices "pos-master/services/user_services"
	"pos-master/utils"

	"github.com/gin-gonic/gin"
)

func GetUsersHandler(c *gin.Context) {

	// Read pagination and search parameters from query string
	page := c.DefaultQuery("page", "1")
	pageSize := c.DefaultQuery("pageSize", "10")
	searchQuery := c.DefaultQuery("search", "")

	// Parse page and pageSize to integers
	pageNum, err := strconv.Atoi(page)
	if err != nil {
		pageNum = 1
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil {
		pageSizeNum = 10
	}

	getRequest := models.SearchRequest{
		GetRequest: models.GetRequest{
			Page:     pageNum,
			PageSize: pageSizeNum,
		},
		SearchQuery: searchQuery,
	}

	// Guards against a zero or negative pageSize, which would otherwise divide
	// by zero while computing the page count.
	getRequest.SetDefaults()

	req := &auth.GetUsersRequest{
		Page:        int32(getRequest.Page),
		PageSize:    int32(getRequest.PageSize),
		SearchQuery: getRequest.SearchQuery,
	}

	users, err := userservices.GetUsers(req)

	if err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to retrieve users", "details", string(fmt.Sprintf("error: %v", err)))
		utils.RespondWithError(c, 400, utils.FailedToRetrieve("Users"), fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, utils.SuccessfullyRetrieve("Users"), gin.H{
		"users": users,
	})

}

func EditUserHandler(c *gin.Context) {
	var req auth.EditUserRequest
	if err := utils.BindProtoJSON(c, &req); err != nil {
		utils.RespondWithError(c, 400, "Invalid request format", fmt.Sprintf("error: %v", err))
		return
	}

	err := userservices.EditUser(&req)
	if err != nil {
		utils.RespondWithError(c, 400, utils.FormatError("unable to update user", err))
		return
	}

	utils.RespondWithSuccess(c, "User updated successfully")

}

func GetUserHandler(c *gin.Context) {

	userid := c.Param("user_id")

	user, err := userservices.GetUser(userid)

	if err != nil {
		utils.RespondWithError(c, 400, utils.FormatError("unable to get user info", err))
		return
	}

	utils.RespondWithSuccess(c, "✅ user info retrieved", gin.H{
		"data": user,
	})

}

func DeleteUserHandler(c *gin.Context) {

	userID := c.Param("user_id")

	if userID == "" {
		utils.RespondWithError(c, 400, "user ID is required")
		return
	}

	err := userservices.DeleteUser(userID, c.GetString("userID"))

	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "User deleted successfully")

}

func ChangePasswordHandler(c *gin.Context) {

	var req auth.ChangeEmailOrPasswordRequest
	if err := utils.BindProtoJSON(c, &req); err != nil {
		utils.RespondWithError(c, 400, "Invalid request format", fmt.Sprintf("error: %v", err))
		return
	}

	// The account being changed is always the caller's own, taken from the
	// token rather than the body.
	req.UserId = c.GetString("userID")
	req.IsPasswordRequest = true
	req.IsEmailRequest = false

	if err := userservices.ChangeEmailorPassword(&req); err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "Password updated successfully")

}

func ChangeEmailHandler(c *gin.Context) {

	var req auth.ChangeEmailOrPasswordRequest
	if err := utils.BindProtoJSON(c, &req); err != nil {
		utils.RespondWithError(c, 400, "Invalid request format", fmt.Sprintf("error: %v", err))
		return
	}

	req.UserId = c.GetString("userID")
	req.IsEmailRequest = true
	req.IsPasswordRequest = false

	if err := userservices.ChangeEmailorPassword(&req); err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "Email updated successfully")

}
