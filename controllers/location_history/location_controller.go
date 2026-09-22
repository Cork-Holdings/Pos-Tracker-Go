package locationhistory

import (
	"fmt"
	"log/slog"
	"strconv"

	"pos-master/models"
	"pos-master/proto/posdevices"
	historyservices "pos-master/services/history_services"
	"pos-master/utils"

	"github.com/gin-gonic/gin"
)

func RegisterNewLocationHandler(c *gin.Context) {

	var req posdevices.RegisterLocationHistoryRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	err := historyservices.RegisterNewLocationHistory(&req)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "Location Registered successfully")

}

func GetLocationsHandler(c *gin.Context) {
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

	req := &posdevices.GetLocationHistorysRequest{
		Page:        int32(getRequest.Page),
		PageSize:    int32(getRequest.PageSize),
		SearchQuery: getRequest.SearchQuery,
	}

	history, err := historyservices.GetLocationHistory(req)

	if err != nil {
		utils.Log(slog.LevelError, "❌error", "unable to retrieve location history ", "details", string(fmt.Sprintf("error: %v", err)))
		utils.RespondWithError(c, 400, utils.FailedToRetrieve("location history"), fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, utils.SuccessfullyRetrieve("location history"), gin.H{
		"history": history,
	})

}

func DeleteLocationHandler(c *gin.Context) {
	locationID := c.Param("id")

	if locationID == "" {
		utils.RespondWithError(c, 400, "location ID is required")
		return
	}

	err := historyservices.DeleteLocationHistory(locationID)
	if err != nil {
		utils.RespondWithError(c, 400, fmt.Sprintf("error: %v", err))
		return
	}

	utils.RespondWithSuccess(c, "successfully deleted location")
}
