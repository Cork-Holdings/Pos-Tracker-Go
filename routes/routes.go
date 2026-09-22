package routes

import (
	"pos-master/controllers/apps"
	"pos-master/controllers/auth"
	"pos-master/controllers/business"
	"pos-master/controllers/dashboard"
	"pos-master/controllers/events"
	locationhistory "pos-master/controllers/location_history"
	posdevices "pos-master/controllers/pos_devices"
	terminaltype "pos-master/controllers/terminal_types"
	"pos-master/controllers/users"
	"pos-master/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {

	r.POST("/v1/login", auth.LoginHandler)

	protected := r.Group("/v1")
	protected.Use(middleware.AuthMiddleware())

	//users
	protected.POST("/create-user", auth.RegisterHandler)
	protected.GET("/users/get", users.GetUsersHandler)
	protected.POST("/user/update", users.EditUserHandler)
	protected.GET("/user/get/:user_id", users.GetUserHandler)
	protected.DELETE("/user/delete/:user_id", users.DeleteUserHandler)

	//account self-service
	protected.POST("/user/password/change", users.ChangePasswordHandler)
	protected.POST("/user/email/change", users.ChangeEmailHandler)

	//dashboard
	protected.GET("/dashboard/tiles/get", dashboard.GetTileInfoHandler)
	protected.GET("/dashboard/pie/get", dashboard.GetPieChartDataHandler)
	protected.GET("/dashboard/bar/get", dashboard.GetLineChartHandler)
	protected.GET("/dashboard/events/get", events.GetEventsHandler)

	//Pos devices
	r.POST("/v1/pos/register", posdevices.RegisterPosDeviceHandler)
	protected.GET("/pos/devices/get", posdevices.GetPosDevicesHandler)
	protected.POST("/pos/device/update", posdevices.EditDeviceHandler)
	protected.DELETE("/pos/device/:id", posdevices.DeleteDeviceHandler)
	r.POST("/v1/pos/device/heartbeat", posdevices.HeartBeatHandler)

	//Apps
	protected.POST("/app/register", apps.RegisterAppHandler)
	protected.GET("/apps/get", apps.GetAppsHandler)
	r.POST("/v1/app/update", apps.CheckAppUpdate)
	protected.POST("/app/info/update", apps.EditAppHandler)
	protected.DELETE("/app/:id", apps.DeleteAppHandler)

	//App versions
	protected.POST("/app/version/register", apps.RegisterNewAppVersionHandler)
	protected.GET("/app/versions/get", apps.GetAppVersionsHandler)
	protected.POST("/app/version/update", apps.EditAppVersionHandler)
	protected.DELETE("/app/version/:id", apps.DeleteAppVersionHandler)

	//location history
	r.POST("/v1/location/register", locationhistory.RegisterNewLocationHandler)
	protected.GET("/locations/get", locationhistory.GetLocationsHandler)
	protected.DELETE("/location/:id", locationhistory.DeleteLocationHandler)

	//business
	protected.POST("/business/create", business.CreateBusinessHandler)
	protected.GET("/businesses/get", business.GetBusinessesHandler)
	protected.GET("/business/get/:id", business.GetBusinessById)
	protected.POST("/business/update", business.EditBusinessHandler)
	protected.DELETE("/business/delete/:id", business.DeleteBusinessHandler)
	protected.DELETE("/business/:id", business.DeleteBusinessHandler)

	//Terminal Types
	protected.POST("/terminal-type/register", terminaltype.CreateTerminalTypeHandler)
	r.GET("/v1/terminal-types/get", terminaltype.GetTerminalTypesHandler)
	protected.POST("/terminal-type/update", terminaltype.EditTerminalTypeHandler)
	protected.DELETE("/terminal-type/:id", terminaltype.DeleteTerminalTypeHandler)

}
