package api

import (
	"bytes"
	"fmt"
	"html/template"
	common "listener/common"
	"net"

	"github.com/labstack/echo"
)

// テスト用のhtmlファイルに渡す環境変数の値
type ViewEnv struct {
	EnvAuthBaseUri  string
	EnvManagerBuild string
	EnvFrontBuild   string
	EnvBackBuild    string
}

const errorPage = `
<html>
	<head>
		<title>エラー</title>
	</head>
	<body style="text-align: center;">
		<table style="border: 1px solid gray;width: 480px;">
			<tr><th><h4>エラー</h4></th></tr>
			<tr><td style="text-align: center;">サーバー内でエラーが発生しています</td></tr>
			<tr><td style="text-align: center;">%s</td></tr>
		</table>
	</body>
</html>
`

func makeView(c echo.Context, name string, path string) error {
	requestIp := net.ParseIP(c.RealIP()).String()
	tp := template.New(name)

	t, err := tp.ParseFiles(path)
	if err != nil {
		common.WriteLog("makeView", "error", requestIp, "htmlテンプレートのパースに失敗しました")
		common.WriteLog("makeView", "error", requestIp, err.Error())
		return c.HTML(500, fmt.Sprintf(errorPage, "htmlテンプレートのパースに失敗しました"))
	}

	var buff bytes.Buffer
	if err = t.Execute(&buff, ViewEnv{
		EnvAuthBaseUri:  MANAGER_BASE_URI,
		EnvManagerBuild: MANAGER_BUILD,
		EnvFrontBuild:   MANAGER_BUILD_FRONT,
		EnvBackBuild:    MANAGER_BUILD_BACK,
	}); err != nil {
		common.WriteLog("makeView", "error", requestIp, "htmlデータの生成に失敗しました")
		common.WriteLog("makeView", "error", requestIp, err.Error())
		return c.HTML(500, fmt.Sprintf(errorPage, "htmlデータの生成に失敗しました"))
	}
	return c.HTML(200, buff.String())
}

func ShowLogin(c echo.Context) error {
	return makeView(c, "login", "views/login.html")
}

func ShowOneTime(c echo.Context) error {
	return makeView(c, "onetime", "views/onetime.html")
}

func ShowMenu(c echo.Context) error {
	return makeView(c, "menu", "views/menu.html")
}
