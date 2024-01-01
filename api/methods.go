package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"listener/auth"
	common "listener/common"
	"listener/db"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"
)

// デプロイ実行中はtrueになり、ほとんどの操作を受け付けない
var deployFlug = false

// 最初のログイン画面から送信された認証情報を処理する
// 認証に成功したら、一段階目の認証トークンを送付する
func PostTempToken(c echo.Context) error {
	requestIp := net.ParseIP(c.RealIP()).String()

	if !auth.ValidIp(requestIp) {
		return c.String(400, "連続アクセスの上限に達しました しばらく時間を空けてからアクセスしてください")
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 認証情報のパース
	var authData AuthData
	err = json.Unmarshal(b, &authData)
	if err != nil {
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 受け取ったユーザーIDとハッシュ化したパスワードが一致しない場合はエラー
	if !auth.CheckUser(authData.UserId, authData.Password) {
		return c.String(403, "ユーザー名またはパスワードが違います")
	}

	// 一段階目の認証トークンをランダムな文字列(Token68形式と互換のあるBase64形式)で生成
	tempToken, err := auth.CreateTempToken()
	if err != nil {
		common.WriteLog("PostTempToken", "error", requestIp, "一段階目の認証トークンの生成に失敗しました")
		common.WriteLog("PostTempToken", "error", requestIp, err.Error())
		return c.String(400, "一段階目の認証トークンの生成に失敗しました")
	}

	// メールでワンタイムパスワードを送信
	err = SendMail(
		SMTP_HOST,
		SMTP_PORT,
		SMTP_USER,
		SMTP_PASSWORD,
		SMTP_FROM_ADDRESS,
		[]string{SMTP_TO_ADDRESS},
		"二段階認証テスト ワンタイムパスワード通知",
		fmt.Sprintf("二段階認証テストのワンタイムパスワードを通知します。\n\n"+
			"\tワンタイムパスワード\n\t%s\n\n"+
			"期限:%s\n\n"+
			"----\n"+
			"このメールに覚えがない場合は削除してください\n",
			tempToken.OneTime, common.DateToStringFormated(tempToken.ExpiredTime)),
	)
	if err != nil {
		common.WriteLog("PostTempToken", "error", requestIp, "ワンタイムパスワードのメール送信ができませんでした")
		common.WriteLog("PostTempToken", "error", requestIp, err.Error())
		return c.String(400, "ワンタイムパスワードのメール送信ができませんでした")
	}

	return c.String(201, tempToken.Token)
}

func PostToken(c echo.Context) error {
	requestIp := net.ParseIP(c.RealIP()).String()

	if !auth.ValidIp(requestIp) {
		return c.String(400, "連続アクセスの上限に達しました しばらく時間を空けてからアクセスしてください")
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// 二段階認証情報のパース
	var tfd TowFactorData
	err = json.Unmarshal(b, &tfd)
	if err != nil {
		return c.String(403, "認証情報の送信方式が間違っています")
	}

	// トークンとワンタイムパスワードが一致しなければエラー
	if !auth.CheckTempToken(tfd.TempToken, tfd.OneTime) {
		return c.String(403, "ワンタイムパスワードが一致しません")
	}

	// 二段階認証済みトークンをランダムな文字列(Token68形式と互換のあるBase64形式)で生成
	token, f := auth.CreateToken()
	if !f {
		common.WriteLog("PostToken", "error", requestIp, "トークンの生成に失敗しました")
		common.WriteLog("PostToken", "error", requestIp, err.Error())
		return c.String(400, "トークンの生成に失敗しました")
	}

	// 一時トークンを削除
	auth.RemoveTempToken(tfd.TempToken)

	return c.String(201, token.Token)
}

func DeleteToken(c echo.Context) error {
	f, t := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(404, ResponseMessage{
			Message: "有効なトークンが存在しません",
		})
	}
	auth.RemoveToken(t.Token)
	return c.NoContent(200)
}

func GetCheckToken(c echo.Context) error {
	f, t := checkAuth(c, c.Path(), false)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "有効なトークンがありません",
		})
	}
	return c.JSON(200, t)
}

func execCommand(title string, isImportant bool, requestIp string, command string, args ...string) (string, error) {
	output, err := exec.Command(command, args...).Output()
	text := ""
	if err != nil {
		if isImportant {
			text = fmt.Sprintf("%sに失敗しました : %s : %s", title, err.Error(), string(output))
		}
		if isImportant {
			common.WriteLog(title, "error", requestIp, text)
		}
		return text, err
	}
	if isImportant {
		common.WriteLog(title, "info", requestIp, "コマンドの実行に成功しました")
	}
	return string(output), nil
}

// 指定されたコマンドを実行する
func getCommandResponse(c echo.Context, title string, isImportant bool, command string, args ...string) error {

	return c.JSON(200, ResponseMessage{
		Message: "Render移行後はこのコマンドを実行できません",
	})
	// Render移行後はこのコマンドを実行できない
	/*
		requestIp := net.ParseIP(c.RealIP()).String()
		f, _ := checkAuth(c, c.Path(), isImportant)
		if !f {
			return c.JSON(403, ResponseMessage{
				Message: "権限がありません",
			})
		}

		if isImportant && CheckBuilding() {
			return c.JSON(409, ResponseMessage{
				Message: "サーバーが処理中なため、この処理を受け付けることができません",
			})
		}

		text, err := execCommand(title, isImportant, requestIp, command, args...)
		if err != nil {
			return c.JSON(400, ResponseMessage{
				Message: text,
			})
		}

		if isImportant {
			common.WriteLog(title, "error", requestIp, text)
		}
		return c.JSON(200, ResponseMessage{
			Message: text,
		})
	*/
}

// Gitからプルする
func UpdateRep(c echo.Context) error {
	return getCommandResponse(c, "git pull", true, "bash", "shell/git-pull.sh")
}

// サーバーの状態を取得する
func GetInfo(c echo.Context) error {
	logging := false
	f, _ := checkAuth(c, c.Path(), logging)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}

	info := ServerInfo{}

	info.Building = CheckBuilding()

	output, _ := exec.Command("bash", "shell/get-storage-info.sh").Output()
	outputStr := string(output)
	outputStr = strings.Trim(outputStr, " ")
	outputStr = strings.Trim(outputStr, "\t")
	info.Storage.Used, _ = strconv.ParseInt(GetWord(strings.Split(outputStr, " "), 2), 10, 64)
	info.Storage.Total, _ = strconv.ParseInt(GetWord(strings.Split(outputStr, " "), 1), 10, 64)

	output, _ = exec.Command("bash", "shell/get-memory-info.sh").Output()
	outputStr = string(output)
	outputStr = strings.Trim(outputStr, " ")
	outputStr = strings.Trim(outputStr, "\t")
	info.Memory.Used, _ = strconv.ParseInt(GetWord(strings.Split(outputStr, " "), 2), 10, 64)
	info.Memory.Total, _ = strconv.ParseInt(GetWord(strings.Split(outputStr, " "), 1), 10, 64)

	return c.JSON(200, info)
}

func GetLog(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}
	name := c.Param("name")
	if strings.Contains(name, "/") || strings.Contains(name, "*") || strings.Contains(name, "\\") || strings.Contains(name, ".") {
		return c.JSON(400, ResponseMessage{
			Message: "不正な文字列が含まれています",
		})
	}
	return c.File("logs/" + c.Param("name") + ".log")
}

func GetWord(strs []string, index int) string {
	i := 0
	for _, str := range strs {
		if str != "" {
			if index == i {
				return str
			}
			i++
		}
	}
	return ""
}

func CheckBuilding() bool {
	if deployFlug {
		return true
	}
	output, _ := exec.Command("bash", "shell/check-build.sh").Output()
	if string(output) == "false\n" {
		return false
	}
	return true
}

func PostBuildFront(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsフロントエンドのビルド", true, "shell/build-front.sh")
}

func PostStartFront(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsフロントエンドの起動", true, "shell/start-front.sh")
}

func DeleteDownFront(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsフロントエンドの終了", true, "shell/stop-front.sh")
}

func GetStatusFront(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsフロントエンドの状態確認", false, "shell/get-status-front.sh")
}

func PostBuildBack(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsバックエンドのビルド", true, "shell/build-back.sh", os.Getenv("MANAGER_GOOS"), os.Getenv("MANAGER_GOARCH"))
}

func PostStartBack(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsバックエンドの起動", true, "shell/start-back.sh")
}

func DeleteDownBack(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsバックエンドの終了", true, "shell/stop-back.sh")
}

func GetStatusBack(c echo.Context) error {
	return getCommandResponse(c, "TierReviewsバックエンドの状態確認", false, "shell/get-status-back.sh")
}

func GetLogs(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}
	list, err := db.GetOperationLogs()
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "ログの読み込みに失敗しました",
		})
	}
	return c.JSON(200, list)
}

func GetErrorLogs(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}
	list, err := db.GetErrorLogs()
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "ログの読み込みに失敗しました",
		})
	}
	return c.JSON(200, list)
}

func GetNotifications(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}
	list, err := db.GetNotifications()
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "ログの読み込みに失敗しました",
		})
	}
	return c.JSON(200, list)
}

func PostNotification(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(400, "BODYの読み取りに失敗しました : "+err.Error())
	}
	var notification db.NotificationData
	err = json.Unmarshal(b, &notification)
	if err != nil {
		return c.String(400, "JSONパースに失敗しました : "+err.Error())
	}

	createAt := time.Now()
	if notification.CreatedAt != "" {
		createAt, err = time.Parse("2006-01-02T15:04:05Z", notification.CreatedAt)
	}
	if err != nil {
		return c.String(400, "createAtのパースに失敗しました : "+err.Error())
	}

	data, err := db.CreateNotification(notification.Content, notification.IsImportant, notification.Url, createAt.UTC())
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "通知の作成に失敗しました",
		})
	}
	requestIp := net.ParseIP(c.RealIP()).String()
	common.WriteLog(c.Path(), "info", requestIp, fmt.Sprintf("通知(ID:%d)を生成", data.Id))
	return c.NoContent(200)
}

func DeleteNotification(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "IDが不正です",
		})
	}

	err = db.DeleteNotification(id)
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "ログの削除に失敗しました",
		})
	}

	// IPアドレス取得
	requestIp := net.ParseIP(c.RealIP()).String()
	common.WriteLog(c.Path(), "info", requestIp, fmt.Sprintf("通知(ID:%d)を削除", id))
	return c.NoContent(200)
}

func UpdateNotification(c echo.Context) error {
	f, _ := checkAuth(c, c.Path(), true)
	if !f {
		return c.JSON(403, ResponseMessage{
			Message: "権限がありません",
		})
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "IDが不正です",
		})
	}

	// Bodyの読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(400, "BODYの読み取りに失敗しました : "+err.Error())
	}
	var notification EditingNotificationData
	err = json.Unmarshal(b, &notification)
	if err != nil {
		return c.String(400, "JSONパースに失敗しました : "+err.Error())
	}

	err = db.UpdateNotification(id, notification.IsDeleted)
	if err != nil {
		return c.JSON(400, ResponseMessage{
			Message: "ログの更新に失敗しました : " + err.Error(),
		})
	}

	ft := ""
	if notification.IsDeleted {
		ft = "true"
	} else {
		ft = "false"
	}

	// IPアドレス取得
	requestIp := net.ParseIP(c.RealIP()).String()
	common.WriteLog(c.Path(), "info", requestIp, fmt.Sprintf("通知(ID:%d)を'is_delete=%s'に更新", id, ft))
	return c.NoContent(200)
}

// 参考: GolangでHMAC-SHA256署名する
// https://cipepser.hatenablog.com/entry/2017/05/27/100516
func MakeHMAC(msg string, key string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

func deploy(requestIp string) {
	deployFlug = true
	_, e1 := execCommand("TierReviewsフロントエンドの終了", true, requestIp, "shell/stop-front.sh")
	_, e2 := execCommand("TierReviewsバックエンドの終了", true, requestIp, "shell/stop-back.sh")
	_, e3 := execCommand("git pull", true, requestIp, "bash", "shell/git-pull.sh")
	_, e4 := execCommand("TierReviewsフロントエンドのビルド", true, requestIp, "shell/build-front.sh")
	_, e5 := execCommand("TierReviewsバックエンドのビルド", true, requestIp, "shell/build-back.sh", os.Getenv("MANAGER_GOOS"), os.Getenv("MANAGER_GOARCH"))
	_, e6 := execCommand("TierReviewsバックエンドの起動", true, requestIp, "shell/start-back.sh")
	_, e7 := execCommand("TierReviewsフロントエンドの起動", true, requestIp, "shell/start-front.sh")

	if e1 == nil && e2 == nil && e3 == nil && e4 == nil && e5 == nil && e6 == nil && e7 == nil {
		common.WriteLog("deploy", "success", requestIp, "デプロイは正常に完了しました")
	} else if e6 == nil && e7 == nil {
		common.WriteLog("deploy", "success", requestIp, "デプロイには失敗しましたが、サーバーは正常に起動したようです")
	} else {
		common.WriteLog("deploy", "success", requestIp, "デプロイに失敗したうえ、サーバーも起動しません！！！ヤバ=スンギ")
	}

	deployFlug = false
}

func PostDeploy(c echo.Context) error {
	// IPアドレス取得
	requestIp := net.ParseIP(c.RealIP()).String()
	// イベントタイプ取得
	event := c.Request().Header.Get("X-GitHub-Event")

	// Body（ペイロード）の読み取り
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(400, "ペイロードの読み取りに失敗しました : "+err.Error())
	}

	// 署名をチェック
	err = checkWebhookSign(c, b)
	if err != nil {
		return c.String(403, "WebHookの署名が一致しません")
	}

	common.WriteLog(c.Path(), "info", requestIp, fmt.Sprintf("WebHookをイベント'%s'で受け取りました", event))

	// ペイロードをチェック
	err = checkWebhookPayload(event, b)
	if err != nil {
		common.WriteLog(c.Path(), "info", requestIp, "受理されましたが、対象外のイベントだったためデプロイは実行されません : "+err.Error())
		return c.String(202, "受理されましたが、対象外のイベントだったためデプロイは実行されません : "+err.Error())
	}

	go deploy(requestIp)

	return c.JSON(202, "デプロイの開始に成功しました")
}
