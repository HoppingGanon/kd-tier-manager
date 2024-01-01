package auth

import (
	"crypto/rand"
	"errors"
	"fmt"
	common "listener/common"
	"log"
	"math/big"
	"time"

	"github.com/emirpasic/gods/lists/arraylist"
)

type TempToken struct {
	Token       string    `json:"token"`
	OneTime     string    `json:"oneTime"`
	ExpiredTime time.Time `json:"expiredTime"`
}

type AuthorizedToken struct {
	Token       string    `json:"token"`
	ExpiredTime time.Time `json:"creationTime"`
}

type RequestCount struct {
	RequestIp    string    `json:"requestIp"`
	CreationTime time.Time `json:"creationTime"`
}

// 乱数生成の挑戦回数
const randomChallenge = 5

var tempTokens = map[string]TempToken{}
var tokens = map[string]AuthorizedToken{}
var requestCounter = arraylist.New()

var UserId = "root"
var HashedPassword = "5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8"

// 重複IPの最大カウント
var DuplicateIpMax = 20

// 一時トークン・ワンタイムパスワードの寿命(秒)
var TempTokenLifespan = 300

//トークンの寿命(秒)
var TokenLifespan = 600

// アクセスがあった場合に、トークンの寿命を回復するかどうかのフラグ
var IsTokenContinue = true

// 認証リクエストIPの回数を集計する期間
var RequestCountSpan = 300

func RemoveTempToken(token string) {
	delete(tempTokens, token)
}

func RemoveToken(token string) {
	delete(tokens, token)
}

func Start() {
	go func() {
		for {
			Expire()
			time.Sleep(time.Second * time.Duration(60))
		}
	}()
}

func Expire() {
	// 一時トークンの整理
	for k, v := range tempTokens {
		if v.ExpiredTime.Before(time.Now()) {
			RemoveTempToken(k)
		}
	}
	// トークンの整理
	for k, v := range tokens {
		if v.ExpiredTime.Before(time.Now()) {
			RemoveToken(k)
		}
	}

	// 削除対象のインデックスをサーチする
	var rc RequestCount
	indexes := make([]bool, requestCounter.Size())
	requestCounter.Each(func(index int, value interface{}) {
		rc = value.(RequestCount)
		if rc.CreationTime.Add(time.Second * time.Duration(RequestCountSpan)).Before(time.Now()) {
			indexes[index] = true
		} else {
			indexes[index] = false
		}
	})

	//逆順削除
	for i := len(indexes) - 1; i >= 0; i-- {
		if indexes[i] {
			requestCounter.Remove(i)
		}
	}
}

// 認証情報をクリア
func Clear() {
	for k := range tempTokens {
		delete(tempTokens, k)
	}
	for k := range tokens {
		delete(tokens, k)
	}
	requestCounter.Clear()
}

// 連続アクセスしているIPかどうかチェック
func ValidIp(ip string) bool {
	var rc RequestCount

	cnt := 0
	requestCounter.Each(func(index int, value interface{}) {
		rc = value.(RequestCount)
		if ip == rc.RequestIp {
			cnt++
		}
	})
	if cnt < DuplicateIpMax {
		requestCounter.Add(RequestCount{
			RequestIp:    ip,
			CreationTime: time.Now(),
		})
		return true
	} else {
		common.WriteLog("PostTempToken", "error", ip, "上限を超過した認証リクエストを検知しました")
		return false
	}
}

func CheckUser(userId string, hashedPassword string) bool {
	if userId == UserId && common.GetSHA256(hashedPassword) == HashedPassword {
		return true
	} else {
		return false
	}
}

func CreateTempToken() (TempToken, error) {
	var token string
	var err error
	var isCreated = false

L1:
	for i := 0; i < randomChallenge; i++ {
		token, err = common.GetRandomBase64()
		if err == nil {
			_, ok := tempTokens[token]
			if !ok {
				isCreated = true
				break L1
			}
		}
	}

	if !isCreated {
		return TempToken{}, errors.New("一時トークンの生成に失敗しました")
	}

	n, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return TempToken{}, errors.New("ワンタイムパスワードの生成に失敗しました")
	}
	onetime := fmt.Sprintf("%06d", n)

	tempTokens[token] = TempToken{
		Token:       token,
		OneTime:     onetime,
		ExpiredTime: time.Now().Add(time.Second * time.Duration(TempTokenLifespan)),
	}

	return tempTokens[token], nil
}

func CheckTempToken(token string, oneTime string) bool {
	v, ok := tempTokens[token]
	if ok && v.OneTime == oneTime && v.ExpiredTime.After(time.Now()) {
		return true
	}
	return false
}

func CreateToken() (AuthorizedToken, bool) {
	var token string
	var err error
	var isCreated = false

L1:
	for i := 0; i < randomChallenge; i++ {
		token, err = common.GetRandomBase64()
		if err == nil {
			_, ok := tempTokens[token]
			if !ok {
				isCreated = true
				break L1
			}
		}
	}

	if !isCreated {
		log.Println("トークンの生成に失敗しました")
		return AuthorizedToken{}, false
	}

	tokens[token] = AuthorizedToken{
		Token:       token,
		ExpiredTime: time.Now().Add(time.Second * time.Duration(TokenLifespan)),
	}

	return tokens[token], true
}

func CheckToken(token string) (bool, AuthorizedToken) {
	v, ok := tokens[token]
	if ok && v.ExpiredTime.After(time.Now()) {
		if IsTokenContinue {
			// トークン期限の更新
			v.ExpiredTime = time.Now().Add(time.Second * time.Duration(TokenLifespan))
			tokens[token] = v
		}
		return true, v
	}

	return false, v
}
