package common

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// SHA256のハッシュをバイナリで返す
func GetBinSHA256(s string) []byte {
	r := sha256.Sum256([]byte(s))
	return r[:]
}

// SHA256の文字列(hex)をバイナリで返す
func GetSHA256(s string) string {
	return hex.EncodeToString(GetBinSHA256(s))
}

func Substring(s string, start int, count int) string {
	if len(s) < start {
		return ""
	} else if len(s) < start+count {
		return s[start:]
	} else {
		return s[start : start+count]
	}
}

// ランダムなハッシュをBase64形式で生成する
func GetRandomBase64() (string, error) {
	// 1億通りの乱数を生成
	n, err := rand.Int(rand.Reader, big.NewInt(100000000))
	if err != nil {
		return "", nil
	}
	// ハッシュをBase64にして返す
	hashbyte := GetBinSHA256(fmt.Sprintf("%s", n))
	return base64.StdEncoding.EncodeToString(hashbyte), nil
}

func DateToStringFormated(t time.Time) string {
	return t.UTC().Add(time.Hour * time.Duration(9)).Format("01月02日15時04分05秒")
}

func DateToString(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05Z")
}

func DateToStringJp(t time.Time) string {
	return t.UTC().Add(time.Hour * time.Duration(9)).Format("2006-01-02T15:04:05Z")
}

func ExistsPath(path string) bool {
	return true
	// Render移行後、この機能は使用しない
	/*
		_, err := os.Stat(path)
		return !os.IsNotExist(err)
	*/
}

func WriteLog(title string, status string, ip string, message string, args ...any) {
	// Render移行後、Logは使用しない
	/*
		if !ExistsPath("logs") {
			if err := os.Mkdir("logs", 0777); err != nil {
				fmt.Println(err.Error())
			}
		}
		logpath := fmt.Sprintf("logs/%s.log", time.Now().UTC().Add(time.Hour*time.Duration(9)).Format("2006-01-02"))
		if ExistsPath(logpath) {
			// 追記
			if f, err := os.OpenFile(logpath, os.O_APPEND|os.O_WRONLY, 0600); err != nil {
				fmt.Println(err)
			} else {
				defer f.Close()
				fmt.Println(fmt.Sprintf(("%s\t%s\t%s\t"), title, status, ip) + fmt.Sprintf(message+"\n", args...))
				f.WriteString(fmt.Sprintf(("%s\t%s\t%s\t"), title, status, ip) + fmt.Sprintf(message+"\n", args...))
			}
		} else {
			// 作成
			if f, err := os.OpenFile(logpath, os.O_CREATE|os.O_RDWR, 0600); err != nil {
				fmt.Println(err)
			} else {
				defer f.Close()
				fmt.Println(fmt.Sprintf(("%s\t%s\t%s\t"), title, status, ip) + fmt.Sprintf(message+"\n", args...))
				f.WriteString(fmt.Sprintf(("%s\t%s\t%s\t"), title, status, ip) + fmt.Sprintf(message+"\n", args...))
			}
		}
	*/
}
