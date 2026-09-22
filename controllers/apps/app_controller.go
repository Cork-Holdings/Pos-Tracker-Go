package apps

import (
	"fmt"
	"log/slog"
	"strconv"

	"pos-master/models"
	appPb "pos-master/proto/app"
	"pos-master/proto/posdevices"
	appservices "pos-master/services/app_services"
	"pos-master/utils"
	"pos-master/utils/sentry"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

func RegisterAppHandler(c *gin.Context) {

	var req appPb.RegisterAppRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	err := appservices.RegisterApp(&req)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "App Registered successfully")

}

func RegisterNewAppVersionHandler(c *gin.Context) {

	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		utils.Log(slog.LevelError, "❌error", "failed to parse form")
		utils.RespondWithError(c, 400, "Failed to parse form", fmt.Sprintf("error: %v", err))
		return
	}

	versionData := c.Request.FormValue("version")

	if versionData == "" {
		utils.Log(slog.LevelError, "❌error", "invalid app version data")
		utils.RespondWithError(c, 400, "App version Data is missing")
		return
	}

	var req appPb.RegisterAppVersionRequest

	if err := protojson.Unmarshal([]byte(versionData), &req); err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to marshal app data")
		utils.RespondWithError(c, 400, "Unable to marshal app data", fmt.Sprintf("error: %v", err))
		return
	}

	err := appservices.RegisterAppVersion(c, &req)

	if err != nil {
		utils.Log(slog.LevelError, "❌error", fmt.Sprintf("error: %v", err))
		utils.RespondWithError(c, 400, "Unable to upload new app to version", fmt.Sprintf("error: %v", err))
		return

	}

	utils.RespondWithSuccess(c, "Added new app version data")

}

func GetAppsHandler(c *gin.Context) {

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

	req := &appPb.GetAppsRequest{
		Page:        int32(getRequest.Page),
		PageSize:    int32(getRequest.PageSize),
		SearchQuery: getRequest.SearchQuery,
	}

	apps, err := appservices.GetApps(req)

	if err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to retrieve apps", "details", string(fmt.Sprintf("error: %v", err)))
		utils.RespondWithError(c, 400, utils.FailedToRetrieve("Apps"), fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, utils.SuccessfullyRetrieve("Apps"), gin.H{
		"apps": apps,
	})

}

func GetAppVersionsHandler(c *gin.Context) {
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

	req := &appPb.GetAppVersionsRequest{
		Page:        int32(getRequest.Page),
		PageSize:    int32(getRequest.PageSize),
		SearchQuery: getRequest.SearchQuery,
	}

	apps, err := appservices.GetAppVersions(req)

	if err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to retrieve app versions ", "details", string(fmt.Sprintf("error: %v", err)))
		utils.RespondWithError(c, 400, utils.FailedToRetrieve("Apps versions"), fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, utils.SuccessfullyRetrieve("Apps versions"), gin.H{
		"apps": apps,
	})

}

func CheckAppUpdate(c *gin.Context) {

	var req posdevices.CheckUpdateRequest

	utils.Log(slog.LevelInfo, "✅info", "check app update", "details", string(fmt.Sprintf("request: %v", c.Request)))

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to bind request", "details", string(fmt.Sprintf("error: %v", err)))
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	updateData, err := appservices.CheckAppUpdate(&req)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "Checked for update successfully", gin.H{
		"data": updateData,
	})
}

func EditAppHandler(c *gin.Context) {
	var req appPb.EditAppRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	err := appservices.EditApp(&req)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "App updated successfully")

}

func DeleteAppVersionHandler(c *gin.Context) {

	versionId := c.Param("id")

	if versionId == "" {
		utils.RespondWithError(c, 400, "version ID is required")
		return
	}

	err := appservices.DeleteAppVersion(versionId)

	if err != nil {
		sentry.SentryLogger(c, err)
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "successfully deleted app version")

}

func EditAppVersionHandler(c *gin.Context) {
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		utils.Log(slog.LevelError, "❌error", "failed to parse form")
		utils.RespondWithError(c, 400, "Failed to parse form", fmt.Sprintf("error: %v", err))
		return
	}

	versionData := c.Request.FormValue("version")

	if versionData == "" {
		utils.Log(slog.LevelError, "❌error", "invalid app version data")
		utils.RespondWithError(c, 400, "App version Data is missing")
		return
	}

	var req appPb.EditAppVersionRequest

	if err := protojson.Unmarshal([]byte(versionData), &req); err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to marshal app data")
		utils.RespondWithError(c, 400, "Unable to marshal app data", fmt.Sprintf("error: %v", err))
		return
	}

	err := appservices.EditAppVersion(c, &req)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "App updated successfully")

}

func DeleteAppHandler(c *gin.Context) {
	appId := c.Param("id")

	if appId == "" {
		utils.RespondWithError(c, 400, "app ID is required")
		return
	}

	err := appservices.DeleteApp(appId)

	if err != nil {
		sentry.SentryLogger(c, err)
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}
	utils.RespondWithSuccess(c, "successfully deleted app")

}
