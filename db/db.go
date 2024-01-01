package db

import (
	"errors"
	"fmt"
	"listener/common"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// データベース
var Db *gorm.DB = nil

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("環境変数'%s'が指定されていません", key))
	} else {
		fmt.Printf("%s=%s\n", key, val)
	}
	return val
}

// データベースに接続する関数
func ConnectDB() *gorm.DB {
	if Db != nil {
		return Db
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=%s",
		getEnv("BACK_DB_HOST"),
		getEnv("BACK_DB_USER"),
		getEnv("BACK_DB_PASSWORD"),
		getEnv("BACK_DB_NAME"),
		getEnv("BACK_DB_PORT"),
		getEnv("BACK_DB_TIMEZONE"))

	var err error
	Db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		common.WriteLog("ConnectDB", "error", "local server", err.Error())
		fmt.Println("データベース接続エラー")
	}

	if Db == nil {
		fmt.Println("データベース接続エラー")
	} else {
		fmt.Println("データベース接続を確認")
	}

	return Db
}

func GetOperationLogs() ([]OperationLogData, error) {
	ConnectDB()
	if Db == nil {
		fmt.Println("データベース接続エラー")
		return []OperationLogData{}, errors.New("データベース接続エラー")
	}
	var list []OperationLogData
	db := Db.Model(&OperationLog{}).Order("created_at DESC").Select("operation_logs.*", "users.name").Joins("left join users on operation_logs.user_id = users.user_id").Limit(1000).Scan(&list)

	return list, db.Error
}

func GetErrorLogs() ([]ErrorLogData, error) {
	ConnectDB()
	if Db == nil {
		fmt.Println("データベース接続エラー")
		return []ErrorLogData{}, errors.New("データベース接続エラー")
	}
	var list []ErrorLogData
	db := Db.Model(&ErrorLog{}).Order("created_at DESC").Select("error_logs.*", "users.name").Joins("left join users on error_logs.user_id = users.user_id").Limit(1000).Scan(&list)

	return list, db.Error
}

func GetNotifications() ([]NotificationData, error) {
	ConnectDB()
	if Db == nil {
		fmt.Println("データベース接続エラー")
		return []NotificationData{}, errors.New("データベース接続エラー")
	}
	var list []NotificationData
	db := Db.Order("created_at DESC, id DESC").Limit(1000).Model(&Notification{}).Scan(&list)

	return list, db.Error
}

func CreateNotification(content string, isImpotant bool, url string, createdAt time.Time) (Notification, error) {
	ConnectDB()
	if Db == nil {
		fmt.Println("データベース接続エラー")
		return Notification{}, errors.New("データベース接続エラー")
	}
	notification := Notification{
		Content:     content,
		IsImportant: isImpotant,
		Url:         url,
		CreatedAt:   createdAt,
		IsDeleted:   false,
	}
	db := Db.Create(&notification)
	fmt.Println(notification)

	return notification, db.Error
}

func DeleteNotification(id int64) error {
	ConnectDB()
	if Db == nil {
		fmt.Println("データベース接続エラー")
		return errors.New("データベース接続エラー")
	}
	db := Db.Where("id = ?", id).Delete(&Notification{})

	return db.Error
}

func UpdateNotification(id int64, isDelete bool) error {
	ConnectDB()
	if Db == nil {
		fmt.Println("データベース接続エラー")
		return errors.New("データベース接続エラー")
	}
	db := Db.Model(&Notification{}).Where("id = ?", id).Update("is_deleted", isDelete)

	return db.Error
}

// ユーザーデータ
type User struct {
	UserId           string `gorm:"primaryKey;not null"`    // ランダムで決定するユーザー固有のID
	IconUrl          string `gorm:"not null"`               // TwitterのアイコンURL
	Name             string `gorm:"not null"`               // 登録名
	Profile          string `gorm:"not null"`               // 自己紹介文
	AllowTwitterLink bool   `gorm:"not null;default:false"` // Twitterへのリンク許可
	KeepSession      int    `gorm:"not null;default:3600"`  // セッション保持時間(秒)

	TwitterId       string `gorm:""` // TwitterID(自分自身でのログイン時およびTwitter連携を許可した時のみ開示)
	TwitterUserName string `gorm:""` // @名
	GoogleId        string `gorm:""` // Google 固有ID
	GoogleEmail     string `gorm:""` // Google Gmailアドレス

	CreatedAt time.Time `gorm:""` // 作成日
	UpdatedAt time.Time `gorm:""` // 更新日
}

// アクセスログ
// 条件: ログイン、ログアウト、ユーザー登録・変更・削除、Tier作成・編集・削除、レビュー作成・編集・削除
type OperationLog struct {
	UserId    string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	Operation string    `gorm:"not null"`                 // 操作対象(エラーコードに準じる)
	Content   string    `gorm:"not null"`                 // 操作内容
	CreatedAt time.Time `gorm:"not null;index"`           // 作成日
}

type OperationLogData struct {
	UserId    string    `json:"userId"`    // ユーザーデータの固有ID
	IpAddress string    `json:"ipAddress"` // セッション確立時のIPアドレス
	Operation string    `json:"operation"` // 操作対象(エラーコードに準じる)
	Content   string    `json:"content"`   // 操作内容
	CreatedAt time.Time `json:"createdAt"` // 作成日

	TwitterName string `json:"twitterName"`
	Name        string `json:"name"`
}

// エラーログ
// 条件: 致命的なエラーの場合
type ErrorLog struct {
	UserId       string    `gorm:"not null"`                 // ユーザーデータの固有ID
	IpAddress    string    `gorm:"not null;default:0.0.0.0"` // セッション確立時のIPアドレス
	ErrorId      string    `gorm:"not null"`                 // エラーID
	Operation    string    `gorm:"not null"`                 // 操作内容
	Descriptions string    `gorm:"not null"`                 // 操作内容(詳細)
	CreatedAt    time.Time `gorm:"not null;index"`           // 作成日
}

type ErrorLogData struct {
	UserId       string    `json:"userId"`       // ユーザーデータの固有ID
	IpAddress    string    `json:"ipAddress"`    // セッション確立時のIPアドレス
	ErrorId      string    `json:"errorId"`      // エラーID
	Operation    string    `json:"operation"`    // 操作内容
	Descriptions string    `json:"descriptions"` // 操作内容(詳細)
	CreatedAt    time.Time `json:"createdAt"`    // 作成日

	TwitterName string `json:"twitterName"`
	Name        string `json:"name"`
}

type Notification struct {
	Id          uint      `gorm:"primaryKey"`
	Content     string    `gorm:""`                       // 表示する文章
	IsImportant bool      `gorm:"default:false;not null"` // 重要情報フラグ
	Url         string    `gorm:""`                       // クリックした際に飛ぶURL
	CreatedAt   time.Time `gorm:"index"`                  // 発信日時
	IsDeleted   bool      `gorm:"default:false;index"`    // 削除フラグ
}

type NotificationData struct {
	Id          uint   `json:"id"`
	Content     string `json:"content"`
	IsImportant bool   `json:"isImportant"`
	Url         string `json:"url"`
	CreatedAt   string `json:"createdAt"`
	IsDeleted   bool   `json:"isDeleted"`
}
