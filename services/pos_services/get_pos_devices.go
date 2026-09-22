package posservices

import (
	"log/slog"
	database "pos-master/config"
	"pos-master/models"
	posPb "pos-master/proto/posdevices"
	"pos-master/utils"
	"time"
)

func GetPosDevices(req *posPb.GetPosDevicesRequest) (*posPb.GetPosDevicesResponse, error) {

	var pos_devices []models.PosDevice

	tx := database.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	query := tx.Model(&models.PosDevice{})

	if req.BusinessId != "" {
		query = query.Where("business_id = ?", req.BusinessId)
	}

	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}

	if req.PhoneNumber != "" {
		query = query.Where("phone_number1 LIKE ? OR phone_number2 LIKE ?", "%"+req.PhoneNumber+"%", "%"+req.PhoneNumber+"%")
	}

	if req.SerialNumber != "" {
		query = query.Where("serial_number LIKE ?", "%"+req.SerialNumber+"%")
	}

	if req.AppVersion != "" {
		query = query.Where("current_app_version = ?", req.AppVersion)
	}

	if req.LocationLastUpdatedStart != "" {
		locationStartDate, err := time.Parse("2006-01-02", req.LocationLastUpdatedStart)
		if err == nil {
			query = query.Where("location_last_updated_at >= ?", locationStartDate)
		}
	}

	if req.LocationLastUpdatedEnd != "" {
		locationEndDate, err := time.Parse("2006-01-02", req.LocationLastUpdatedEnd)
		if err == nil {
			// Add 1 day to include the entire end date
			locationEndDate = locationEndDate.AddDate(0, 0, 1)
			query = query.Where("location_last_updated_at < ?", locationEndDate)
		}
	}

	if req.StartDate != "" {
		startDate, err := time.Parse(time.RFC3339, req.StartDate)
		if err != nil {
			return nil, utils.CapitalizeError("Invalid start date")
		}
		query = query.Where("created_at >= ?", startDate)
	}

	if req.EndDate != "" {
		endDate, err := time.Parse(time.RFC3339, req.EndDate)
		if err != nil {
			return nil, utils.CapitalizeError("Invalid end date")
		}
		query = query.Where("created_at <= ?", endDate)
	}

	var totalPosDevices int64
	err := query.Count(&totalPosDevices).Error

	if err != nil {
		return nil, utils.CapitalizeError("failed to count pos device pos devices")
	}

	totalPages := int32((totalPosDevices + int64(req.PageSize) - 1) / int64(req.PageSize))
	// Calculate offset for pagination
	offset := (req.Page - 1) * req.PageSize

	// Execute the final query with pagination and preloading
	err = query.Order("created_at DESC").Limit(int(req.PageSize)).
		Offset(int(offset)).
		Find(&pos_devices).Error

	if err != nil {
		return nil, utils.CapitalizeError("failed to retrieve pos device pos devices")
	}

	pbPosDevices := make([]*posPb.PosDevice, len(pos_devices))

	for i, history := range pos_devices {
		pbPosDevices[i] = &posPb.PosDevice{
			Id:                  history.ID.String(),
			SerialNumber:        history.SerialNumber,
			Name:                history.Name,
			CurrentAppVersion:   history.CurrentAppVersion,
			LastKnownLatitude:   history.LastKnownLatitude,
			LastKnownLongitude:  history.LastKnownLongitude,
			Status:              history.Status,
			DeviceModel:         history.DeviceModel,
			OperatingSystem:     history.OperatingSystem,
			Description:         history.Description,
			LocationLastUpdated: history.LocationLastUpdatedAt.Format(time.RFC3339),
			BusinessName:        history.Entity,
			PrimaryNumber:       history.PhoneNumber1,
			SecondaryNumber:     history.PhoneNumber2,
		}
	}

	return &posPb.GetPosDevicesResponse{
		Posdevice:   pbPosDevices,
		TotalPages:  totalPages,
		CurrentPage: req.Page,
		HasMore:     req.Page < totalPages,
		Count:       int32(totalPosDevices),
	}, nil
}

func GetPosById(posID string) (*posPb.PosDevice, error) {

	var posdevice models.PosDevice

	tx := database.DB.Begin()

	result := tx.First(&posdevice, "id = ?", posID)

	if result.Error != nil {
		utils.Log(slog.LevelError, "❌ error", result.Error)
		return nil, result.Error
	}

	return &posPb.PosDevice{
		Id:                 posdevice.ID.String(),
		SerialNumber:       posdevice.SerialNumber,
		Name:               posdevice.Name,
		Description:        posdevice.Description,
		CurrentAppVersion:  posdevice.CurrentAppVersion,
		LastKnownLatitude:  posdevice.LastKnownLatitude,
		LastKnownLongitude: posdevice.LastKnownLongitude,
		Status:             posdevice.Status,
		DeviceModel:        posdevice.DeviceModel,
		OperatingSystem:    posdevice.OperatingSystem,
	}, nil

}
