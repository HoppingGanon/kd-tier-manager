package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo"

	"listener/auth"
	common "listener/common"
)

// envLoad 環境変数のロード
func LoadEnv() {
	fmt.Println("---------------------")
	fmt.Println("環境変数の読み込み")
	fmt.Println("")

	// 環境変数を読み込む
	GIT_REMOTE_BRANCH = getEnv("GIT_REMOTE_BRANCH", GIT_REMOTE_BRANCH)
	GIT_LOCAL_BRANCH = getEnv("GIT_LOCAL_BRANCH", GIT_LOCAL_BRANCH)
	MANAGER_LISTENER_PORT = getEnv("MANAGER_LISTENER_PORT", MANAGER_LISTENER_PORT)
	MANAGER_BASE_URI = getEnv("MANAGER_BASE_URI", MANAGER_BASE_URI)
	MANAGER_HOOK_SECRET = getEnv("MANAGER_HOOK_SECRET", MANAGER_HOOK_SECRET)
	MANAGER_GOOS = getEnv("MANAGER_GOOS", MANAGER_GOOS)
	MANAGER_GOARCH = getEnv("MANAGER_GOARCH", MANAGER_GOARCH)

	MANAGER_BUILD = getEnv("MANAGER_BUILD", MANAGER_BUILD)
	MANAGER_BUILD_FRONT = getEnv("MANAGER_BUILD_FRONT", MANAGER_BUILD_FRONT)
	MANAGER_BUILD_BACK = getEnv("MANAGER_BUILD_BACK", MANAGER_BUILD_BACK)

	SMTP_HOST = getEnv("SMTP_HOST", SMTP_HOST)
	SMTP_PORT = getEnv("SMTP_PORT", SMTP_PORT)
	SMTP_USER = getEnv("SMTP_USER", SMTP_USER)
	SMTP_PASSWORD = getEnv("SMTP_PASSWORD", SMTP_PASSWORD)
	SMTP_FROM_ADDRESS = getEnv("SMTP_FROM_ADDRESS", SMTP_FROM_ADDRESS)
	SMTP_TO_ADDRESS = getEnv("SMTP_TO_ADDRESS", SMTP_TO_ADDRESS)
	MANAGER_ROOT_USERID = getEnv("MANAGER_ROOT_USERID", MANAGER_ROOT_USERID)
	MANAGER_ROOT_HASED_PASSWORD = getEnv("MANAGER_ROOT_HASED_PASSWORD", MANAGER_ROOT_HASED_PASSWORD)
	MANAGER_WEBHOOK_TIME = getEnvNumber("MANAGER_WEBHOOK_TIME", MANAGER_WEBHOOK_TIME)

	fmt.Println("")
	fmt.Println("---------------------")
}

func getEnv(envName string, defaultValue string) string {
	val := os.Getenv(envName)
	if val == "" {
		fmt.Printf("%s=%s\n", envName, defaultValue)
		return defaultValue
	}
	fmt.Printf("%s=%s\n", envName, val)
	return val
}

func getEnvNumber(envName string, defaultValue int) int {
	text := os.Getenv(envName)
	if text == "" {
		fmt.Printf("%s=%d\n", envName, defaultValue)
		return defaultValue
	}
	i, err := strconv.Atoi(text)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%s=%d\n", envName, i)
	return i
}

// ヘッダからBearerトークンを抜き出す関数
func getBearer(c echo.Context) (string, bool) {
	auth := c.Request().Header.Get("Authorization")
	typeStr := common.Substring(auth, 0, 7)

	if typeStr != "Bearer " {
		return "", false
	}
	return common.Substring(auth, 7, len(auth)-7), true
}

// 5分間有効なワンタイムトークンまたは二段階認証済みトークンをチェックする
func checkAuth(c echo.Context, uri string, logging bool) (bool, auth.AuthorizedToken) {
	Echo.Logger.Warn()
	requestIp := net.ParseIP(c.RealIP()).String()

	header := c.Request().Header.Get("Authorization")
	typeStr := common.Substring(header, 0, 7)

	if typeStr != "Bearer " {
		if logging {
			common.WriteLog("checkAuth", "error", requestIp, "発信元[%s]よりURI[%s]へのアクセスを拒否しました", requestIp, uri)
		}
		return false, auth.AuthorizedToken{}
	}
	receivedToken := common.Substring(header, 7, len(header)-7)

	if receivedToken == "" {
		if logging {
			common.WriteLog("checkAuth", "error", requestIp, "発信元[%s]よりURI[%s]へのアクセスを拒否しました", requestIp, uri)
		}
		return false, auth.AuthorizedToken{}
	}

	var token string

	token = common.GetSHA256(receivedToken)
	if token == MANAGER_HOOK_SECRET {
		if logging {
			common.WriteLog("checkAuth", "error", requestIp, "発信元[%s]よりURI[%s]へのアクセスをWebHook専用トークンで受理しました", requestIp, uri)
		}
		return true, auth.AuthorizedToken{
			Token:       token,
			ExpiredTime: time.Now(),
		}
	}

	f, checkedToken := auth.CheckToken(receivedToken)
	if f {
		if logging {
			common.WriteLog("checkAuth", "error", requestIp, "発信元[%s]よりURI[%s]へのアクセスを二段階認証済みトークンで受理しました", requestIp, uri)
		}
		return true, checkedToken
	}

	if logging {
		common.WriteLog("checkAuth", "error", requestIp, "発信元[%s]よりURI[%s]へのアクセスを拒否しました", requestIp, uri)
	}
	return false, auth.AuthorizedToken{}
}

func SendMail(
	// SMTPサーバーのホスト名
	hostname string,
	// SMTPサーバーのポート番号
	port string,
	// ユーザー名(送信元Gmailアドレス)
	username string,
	// API キー
	password string,
	// 送信元アドレス
	from string,
	// 宛先アドレス
	to []string,
	// 件名
	subject string,
	// 本文
	body string,
) error {
	auth := smtp.PlainAuth("", username, password, hostname)
	msg := []byte(strings.ReplaceAll(fmt.Sprintf("To: %s\nSubject: %s\n\n%s", strings.Join(to, ","), subject, body), "\n", "\r\n"))
	return smtp.SendMail(fmt.Sprintf("%s:%s", hostname, port), auth, from, to, msg)
}

// GitHubWebhookの署名を検証
func checkWebhookSign(c echo.Context, b []byte) error {
	// IPアドレス取得
	requestIp := net.ParseIP(c.RealIP()).String()

	// 署名の検証
	signature := fmt.Sprintf("sha256=%s", MakeHMAC(string(b), MANAGER_HOOK_SECRET))
	if c.Request().Header.Get("x-hub-signature-256") != signature {
		common.WriteLog("checkHookSecret", "error", requestIp, "WebHookの署名が一致しません")
		return errors.New("署名が一致しません")
	}
	return nil
}

// GitHubWebhookのペイロードを検証
func checkWebhookPayload(event string, body []byte) error {

	// 現在時刻を取得
	now := time.Now()

	if event == "push" {
		var p WebhookPush
		err := json.Unmarshal(body, &p)
		if err != nil {
			return err
		}

		if p.Ref != ("refs/heads/" + GIT_REMOTE_BRANCH) {
			return errors.New("デプロイを実行できるWebHookは" + GIT_REMOTE_BRANCH + "ブランチのみです")
		}

		if p.Repository.FullName != ("HoppingGanon/tier-reviews") {
			return errors.New("デプロイを実行できるリポジトリはHoppingGanon/tier-reviewsのみです")
		}

		// 最終更新が近いかどうか
		t1, _ := time.Parse("2006-01-02T15:04:05Z07:00", p.Repository.UpdatedAt)
		f1 := t1.After(now.Add(time.Duration(-1*MANAGER_WEBHOOK_TIME)*time.Minute)) && t1.Before(now.Add(time.Duration(MANAGER_WEBHOOK_TIME)*time.Minute))

		t2, _ := time.Parse("2006-01-02T15:04:05Z07:00", p.HeadCommit.Timestamp)
		f2 := t2.Before(now.Add(time.Duration(MANAGER_WEBHOOK_TIME)*time.Minute)) && t2.After(now.Add(time.Duration(-1*MANAGER_WEBHOOK_TIME)*time.Minute))

		// リポジトリの最終更新が近いかどうか
		if f1 == false && f2 == false {
			return errors.New(fmt.Sprintf("デプロイを実行できるWebHookは現在時刻から前後%d分のものに限ります", MANAGER_WEBHOOK_TIME))
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("イベント'%s'はデプロイの対象外です", event))
	}

}
