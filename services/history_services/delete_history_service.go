package historyservices

import (
	"errors"

	database "pos-master/config"
	"pos-master/models"
	eventservices "pos-master/services/event_services"
	"pos-master/utils"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func DeleteLocationHistory(locationID string) error {
	parsedID, err := uuid.Parse(locationID)
	if err != nil {
		return utils.CapitalizeError("invalid ID format")
	}

	var location models.LocationHistory
	if err := database.DB.Where("id = ?", parsedID).First(&location).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return utils.CapitalizeError("unable to find location with that ID")
		}
		return utils.CapitalizeError("failed to look up location")
	}

	if err := database.DB.Delete(&models.LocationHistory{}, "id = ?", parsedID).Error; err != nil {
		return utils.CapitalizeError("failed to delete location")
	}

	eventservices.RegisterEvent("Location deleted successfully", map[string]interface{}{
		"Location ID": locationID,
	})
	return nil
}
