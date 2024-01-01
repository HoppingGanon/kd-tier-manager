package main

import (
	"listener/api"
	"listener/auth"
	"listener/common"
	"listener/db"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
)

func main() {
	api.Echo = echo.New()

	common.WriteLog("main", "info", "local server", "サーバーを起動しています")

	// 環境変数の読み込み
	api.LoadEnv()

	// データベースへ接続
	db.ConnectDB()

	// これを設定しないと、同オリジンからのアクセスが拒否される
	api.Echo.Use(middleware.CORS())

	auth.UserId = api.MANAGER_ROOT_USERID
	auth.HashedPassword = api.MANAGER_ROOT_HASED_PASSWORD

	api.Echo.GET("/view/login.html", api.ShowLogin)
	api.Echo.GET("/view/onetime.html", api.ShowOneTime)
	api.Echo.GET("/view/menu.html", api.ShowMenu)

	api.Echo.POST("/auth/temp-token", api.PostTempToken)
	api.Echo.POST("/auth/token", api.PostToken)
	api.Echo.DELETE("/auth/token", api.DeleteToken)
	api.Echo.GET("/auth/check-token", api.GetCheckToken)

	api.Echo.GET("/server/info", api.GetInfo)
	api.Echo.GET("/server/log/:name", api.GetLog)
	api.Echo.PATCH("/git/pull", api.UpdateRep)

	api.Echo.POST("/front/build", api.PostBuildFront)
	api.Echo.POST("/front/start", api.PostStartFront)
	api.Echo.GET("/front/status", api.GetStatusFront)
	api.Echo.DELETE("/front/stop", api.DeleteDownFront)

	api.Echo.POST("/back/build", api.PostBuildBack)
	api.Echo.POST("/back/start", api.PostStartBack)
	api.Echo.GET("/back/status", api.GetStatusBack)
	api.Echo.DELETE("/back/stop", api.DeleteDownBack)

	api.Echo.GET("/db/logs", api.GetLogs)
	api.Echo.GET("/db/error-logs", api.GetErrorLogs)
	api.Echo.GET("/db/notifications", api.GetNotifications)
	api.Echo.POST("/db/notification", api.PostNotification)
	api.Echo.DELETE("/db/notification/:id", api.DeleteNotification)
	api.Echo.PATCH("/db/notification/:id", api.UpdateNotification)

	api.Echo.POST("/hook/deploy", api.PostDeploy)

	// 定刻処理の開始
	common.WriteLog("main", "info", "local server", "定刻処理を開始")
	auth.Start()

	// リスナーポート番号
	common.WriteLog("main", "info", "local server", "%s番ポートでアクセスを受け付けます", api.MANAGER_LISTENER_PORT)
	api.Echo.Logger.Fatal(api.Echo.Start(":" + api.MANAGER_LISTENER_PORT))
}
