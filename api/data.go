package api

import (
	"github.com/labstack/echo"
)

var Echo *echo.Echo

// 定数
var GIT_TARGET_PATH = "/manager/repositories/tier-reviews/"

// 環境変数
var GIT_REMOTE_BRANCH = "develop"
var GIT_LOCAL_BRANCH = "develop"
var MANAGER_LISTENER_PORT = "8290"
var MANAGER_BASE_URI = "http://localhost:8290"
var MANAGER_HOOK_SECRET = "token"
var MANAGER_GOOS = "linux"
var MANAGER_GOARCH = "amd64"
var MANAGER_BUILD = "RELEASE"
var MANAGER_BUILD_BACK = "RELEASE"
var MANAGER_BUILD_FRONT = "RELEASE"

var BACK_DB_HOST = ""
var BACK_DB_USER = ""
var BACK_DB_PASSWORD = ""
var BACK_DB_NAME = ""
var BACK_DB_PORT = ""
var BACK_DB_TIMEZONE = ""

// WEBフックのヘッダコミット時間の許容誤差(分)
var MANAGER_WEBHOOK_TIME = 240

// SMTPサーバーのホスト 環境変数'SMTP_HOST'に対応する
var SMTP_HOST = "smtp.gmail.com"

// SMTPの宛先ポート 環境変数'SMTP_PORT'に対応する
var SMTP_PORT = "587"

// SMTPに含めるユーザー名 環境変数'SMTP_USER'に対応する
var SMTP_USER = "<from>@gmail.com"

// SMTPに含めるパスワード 環境変数'SMTP_PASSWORD'に対応する
var SMTP_PASSWORD = "<application password>"

// 送信元のメールアドレス 環境変数'SMTP_FROM_ADDRESS'に対応する
var SMTP_FROM_ADDRESS = "<from>@gmail.com"

// 送信先のメールアドレス 環境変数'SMTP_TO_ADDRESS'に対応する
var SMTP_TO_ADDRESS = "<to>@gmail.com"

// ユーザーID
var MANAGER_ROOT_USERID = "root"

// パスワード「password」をSHA256でハッシュ化したもの
var MANAGER_ROOT_HASED_PASSWORD = "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"

type ResponseMessage struct {
	Message string `json:"message"`
}

// ユーザーから送付される認証情報
type AuthData struct {
	UserId   string `json:"userId"`
	Password string `json:"password"`
}

// ユーザーから送付される二段階認証情報
type TowFactorData struct {
	TempToken string `json:"tempToken"`
	OneTime   string `json:"oneTime"`
}

type ServerInfo struct {
	Storage struct {
		Used  int64 `json:"used"`
		Total int64 `json:"total"`
	} `json:"storage"`
	Memory struct {
		Used  int64 `json:"used"`
		Total int64 `json:"total"`
	} `json:"memory"`
	Building bool `json:"building"`
}

type EditingNotificationData struct {
	IsDeleted bool `json:"isDeleted"`
}

type WebhookPush struct {
	Ref        string     `json:"ref"`
	Repository Repository `json:"repository"`
	HeadCommit Commit     `json:"head_commit"`
}

type Repository struct {
	FullName  string `json:"full_name"`
	UpdatedAt string `json:"updated_at"`
}

type Commit struct {
	Timestamp string `json:"timestamp"`
}
