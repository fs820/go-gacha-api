package main // エントリーポイント

// ライブラリのインポート
import (
	"encoding/json" // JSONのエンコード/デコードに使用
	"math/rand/v2"  // 乱数 ガチャ用 (高速)
	"net/http"
)

// 定数の定義
const (
	// cookieの日数
	cookieDays = 30           // セッションIDを保存するCookieの有効期限（日数）
	oneDay     = 24 * 60 * 60 // 1日の秒数（CookieのMaxAgeに使用）

	// ガチャ1回あたりの石の消費量
	gachaCost = 300

	// ガチャの確率を定数として定義
	probBaseStar5 = 6  // 星5の当たる基本確率（6/1000 = 0.6%）
	probBaseStar4 = 51 // 星4の当たる基本確率（51/1000 = 5.1%）

	// 天井の回数を定数として定義
	star4Limit = 10 // 星4以上が出るまでの回数
	star5Limit = 90 // 星5が出るまでの回数

	// ソフトピティの設定
	pitySoftStart     = 74 // 確率が上がり始める回数
	softPityIncrement = 6  // 確率が上がる割合 (6%ずつ増加)
)

// ユーザデータ
type UserData struct {
	Stones                 int
	Star4LimitCounter      int
	Star5LimitCounter      int
	IsNextPickupGuaranteed bool
	GachaHistory           []GachaResult
}

// ガチャの結果を入れる構造体 変数名の先頭が大文字にすると外部からアクセスできる（JSONに変換するために必要）
type GachaResult struct {
	Rarity    string `json:"rarity"`    // レアリティ (`json:"rarity"`は、JSONに変換するときのキー名)
	Character string `json:"character"` // キャラクター名 (`json:"character"`は、JSONに変換するときのキー名)
}

// ブラウザへ返すレスポンス
type GachaResponse struct {
	Results   []GachaResult `json:"results"`   // 今回の結果リスト
	Pity5Star int           `json:"pity5Star"` // 星5天井まであと何回か
	Pity4Star int           `json:"pity4Star"` // 星4天井まであと何回か
	Stones    int           `json:"stones"`    // 所持石数
}

// 石を追加するハンドラー
func addStonesHandler(w http.ResponseWriter, r *http.Request) {
	// POSTリクエストのみ受け付ける
	if r.Method != http.MethodPost {
		http.Error(w, "不正なリクエストです", http.StatusMethodNotAllowed)
		return
	}

	// CookieからユーザーIDを取得
	uid, err := getSession(r)
	if err != nil {
		http.Error(w, "ログインしてください", http.StatusUnauthorized)
		return
	}

	// 石を追加する関数を呼び出す
	err = addStones(uid, 3000)
	if err != nil {
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// ユーザーデータを取得して、レスポンスを返す
	user := getUserData(uid)
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]int{
		"stones": user.Stones,
	})
}

// ガチャの処理を行う関数
func gachaHandler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得
	uid, err := getSession(r)
	if err != nil {
		http.Error(w, "ログインしてください", http.StatusUnauthorized)
		return
	}

	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	// 石の所持数をチェックして、足りない場合はエラーを返す
	if user.Stones < gachaCost {
		http.Error(w, "石が足りません！", http.StatusBadRequest)
		return
	}

	// ガチャの結果を判定する関数を呼び出して、結果を取得
	result := gachaJudgment(user)

	// DB保存
	err = saveGachaResultTx(uid, user, []GachaResult{result}, gachaCost)
	if err != nil {
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// 石を消費
	user.Stones -= gachaCost

	// 履歴に追加 (50件を超えていたら、一番古い要素を切り捨てる)
	user.GachaHistory = append(user.GachaHistory, result)
	if len(user.GachaHistory) > 50 {
		// インデックス1から最後までを残す
		user.GachaHistory = user.GachaHistory[1:]
	}

	// レスポンス作成
	sendGachaResponse(w, []GachaResult{result}, user)
}

// 10連ガチャの処理を行う関数
func gacha10Handler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得
	uid, err := getSession(r)
	if err != nil {
		http.Error(w, "ログインしてください", http.StatusUnauthorized)
		return
	}

	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	// 石の所持数をチェックして、足りない場合はエラーを返す
	if user.Stones < gachaCost*10 {
		http.Error(w, "石が足りません！", http.StatusBadRequest)
		return
	}

	var results []GachaResult
	for i := 0; i < 10; i++ {
		// ガチャの結果を判定する関数を呼び出して、結果を取得して、resultsの配列に追加
		result := gachaJudgment(user)
		results = append(results, result)

		// 履歴に追加 (50件を超えていたら、一番古い要素を切り捨てる)
		user.GachaHistory = append(user.GachaHistory, result)
		if len(user.GachaHistory) > 50 {
			// インデックス1から最後までを残す
			user.GachaHistory = user.GachaHistory[1:]
		}
	}

	// DB保存
	err = saveGachaResultTx(uid, user, results, gachaCost*10)
	if err != nil {
		http.Error(w, "サーバーエラーが発生しました", http.StatusInternalServerError)
		return
	}

	// 石を消費
	user.Stones -= gachaCost * 10

	// レスポンス作成
	sendGachaResponse(w, results, user)
}

// 天井カウンターを返すハンドラー
func limitHandler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得
	uid, err := getSession(r)
	if err != nil {
		http.Error(w, "ログインしてください", http.StatusUnauthorized)
		return
	}

	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]int{
		"star4LimitCounter": star4Limit - user.Star4LimitCounter,
		"star5LimitCounter": star5Limit - user.Star5LimitCounter,
		"stones":            user.Stones,
	})
}

// 履歴を返すハンドラー
func historyHandler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得
	uid, err := getSession(r)
	if err != nil {
		http.Error(w, "ログインしてください", http.StatusUnauthorized)
		return
	}

	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	// 履歴が空の場合は、空の配列を返す
	if len(user.GachaHistory) == 0 {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		json.NewEncoder(w).Encode([]GachaResult{})
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(user.GachaHistory)
}

// 共通のレスポンス送信処理
func sendGachaResponse(w http.ResponseWriter, results []GachaResult, user *UserData) {
	response := GachaResponse{
		Results:   results,
		Pity5Star: star5Limit - user.Star5LimitCounter, // あと何回か
		Pity4Star: star4Limit - user.Star4LimitCounter,
		Stones:    user.Stones, // 所持石数
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// ガチャの結果を判定する関数
func gachaJudgment(user *UserData) GachaResult {
	// カウンターをインクリメント
	user.Star4LimitCounter++ // 星4以上が出るまでのカウンター
	user.Star5LimitCounter++ // 星5が出るまでのカウンター

	// 0〜999の乱数を生成
	roll := rand.IntN(1000)

	// 星5の当たる確率（6/1000 = 0.6%）
	star5Prob := probBaseStar5

	// ソフトピティの確率上昇の判定
	if user.Star5LimitCounter >= pitySoftStart {
		// 74連目以降は、6%ずつ確率が上昇
		star5Prob += softPityIncrement * (user.Star5LimitCounter - (pitySoftStart - 1))
	}

	// 確率の判定
	if roll < star5Prob || user.Star5LimitCounter >= star5Limit {
		// 0.6%の確率で星5 （もしくは、天井カウンターが90連目の場合は強制的に星5）
		user.Star4LimitCounter = 0 // カウンターをリセット
		user.Star5LimitCounter = 0 // カウンターをリセット

		// ピックアップキャラクターの当選判定を行う関数を呼び出す
		return pickupJudgment(user)
	} else if roll < (star5Prob+probBaseStar4) || user.Star4LimitCounter >= star4Limit {
		// 5.1%の確率で星4 （もしくは、天井カウンターが10連目の場合は強制的に星4）
		user.Star4LimitCounter = 0 // カウンターをリセット

		pickupStar4 := getCharactersFromDB("星4", true) // ピックアップ星4キャラクターのリストをDBから取得
		randomIndex := rand.IntN(len(pickupStar4))     // ピックアップ星4キャラクターの中からランダムに選ぶ
		return GachaResult{Rarity: "星4", Character: pickupStar4[randomIndex]}
	} else {
		// 94.3%の確率で星3
		star3 := getCharactersFromDB("星3", false) // 星3キャラクターのリストをDBから取得
		randomIndex := rand.IntN(len(star3))      // 星3キャラクターの中からランダムに選ぶ
		return GachaResult{Rarity: "星3", Character: star3[randomIndex]}
	}
}

// ピックアップキャラクターの当選判定を行う関数
func pickupJudgment(user *UserData) GachaResult {
	// ピックアップキャラクターが確定している場合は、ピックアップキャラクターを返す
	if user.IsNextPickupGuaranteed {
		user.IsNextPickupGuaranteed = false // フラグをリセット

		pickupStar5 := getCharactersFromDB("星5", true)                        // ピックアップ星5キャラクターのリストをDBから取得
		randomIndex := rand.IntN(len(pickupStar5))                            // ピックアップ星5キャラクターの中からランダムに選ぶ
		return GachaResult{Rarity: "星5", Character: pickupStar5[randomIndex]} // ピックアップキャラクターの中から1体を返す（今回は1体しかいない想定）
	} else {
		// ピックアップキャラクターが確定していない場合は、50%の確率でピックアップキャラクター、50%の確率ですり抜けキャラクターを返す
		if rand.IntN(2) == 0 {
			pickupStar5 := getCharactersFromDB("星5", true) // ピックアップ星5キャラクターのリストをDBから取得
			randomIndex := rand.IntN(len(pickupStar5))     // ピックアップ星5キャラクターの中からランダムに選ぶ
			return GachaResult{Rarity: "星5", Character: pickupStar5[randomIndex]}
		} else {
			user.IsNextPickupGuaranteed = true // 次のガチャでピックアップキャラクターが確定するようにフラグをセット

			standardStar5 := getCharactersFromDB("星5", false) // すり抜け星5キャラクターのリストをDBから取得
			randomIndex := rand.IntN(len(standardStar5))      // すり抜けキャラクターの中からランダムに選ぶ
			return GachaResult{Rarity: "星5", Character: standardStar5[randomIndex]}
		}
	}
}
