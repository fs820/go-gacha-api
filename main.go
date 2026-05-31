package main // エントリーポイント

// ライブラリのインポート
import (
	crand "crypto/rand" // 暗号用 (安全)
	"database/sql"      // データベース操作に使用
	"encoding/hex"      // セッションIDの生成に使用
	"encoding/json"     // JSONのエンコード/デコードに使用
	"fmt"               // フォーマット用 (文字列の整形など)
	"log"               // ロギングに使用
	"math/rand/v2"      // ガチャ用 (高速)
	"net/http"          // HTTPサーバーの構築に使用

	// "sync"              // データの競合を防ぐためのロックに使用

	_ "modernc.org/sqlite" // SQLiteドライバ (データベース接続のために必要、_はパッケージの初期化のみを行うための記号
)

// 定数の定義
const (
	DBFilePath = "./gacha.db" // データベースファイルのパス

	// cookieの日数
	cookieDays = 30           // セッションIDを保存するCookieの有効期限（日数）
	oneDay     = 24 * 60 * 60 // 1日の秒数（CookieのMaxAgeに使用）

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

// --- 排出キャラクターのリスト ---
// ピックアップ星5
var pickupStar5 = "ゼウス"

// すり抜け星5（7名）
var standardStar5 = []string{"ウラノス", "クロノス", "釈迦", "キリスト", "シヴァ", "ポセイドン", "ヘラクレス"}

// ピックアップ星4（3名）
var pickupStar4 = []string{"ヨハネ", "千手観音", "アキレス"}

var star3 = "武器" // 星3

// ユーザデータ
type UserData struct {
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
}

// var (
// userDB = make(map[string]*UserData) // ユーザIDをキーにしてユーザデータを保存するマップ
// dbMu   sync.Mutex                   // データの競合を防ぐためのロック
// )

var userDB *sql.DB

// メイン関数
func main() {
	// データベースの初期化
	initDB()

	// "static"フォルダの中身（HTML, CSS, JS）を、そのままブラウザに公開する設定
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// ガチャのエンドポイントを設定
	http.HandleFunc("/gacha", gachaHandler)     // 単発ガチャのエンドポイント /gacha
	http.HandleFunc("/gacha10", gacha10Handler) // 10連ガチャのエンドポイント /gacha10

	// 履歴だけを取得するエンドポイント
	http.HandleFunc("/history", historyHandler)
	// 天井カウンターを取得するエンドポイント
	http.HandleFunc("/limit", limitHandler)

	// サーバー起動のメッセージを表示
	fmt.Println("サーバーを起動しました！ ブラウザで http://localhost:8080 にアクセスしてください。")
	fmt.Println("終了するにはターミナルで Ctrl + C を押します。")

	// ポート8080でサーバーを起動（ゲームのメインループのように、ここでアクセスを待ち続けます）
	http.ListenAndServe(":8080", nil)
}

// データベースの初期化関数
func initDB() {
	var err error
	// gacha.db というファイルを開く（無ければ自動で作られる）
	userDB, err = sql.Open("sqlite", DBFilePath)
	if err != nil {
		log.Fatal(err)
	}

	// ユーザーのカウンター状態を保存するテーブルを作成
	usersTable := `
	CREATE TABLE IF NOT EXISTS users (
		uid TEXT PRIMARY KEY,
		star4_limit_counter INTEGER DEFAULT 0,
		star5_limit_counter INTEGER DEFAULT 0,
		is_next_pickup_guaranteed BOOLEAN DEFAULT 0
	);`
	_, err = userDB.Exec(usersTable)
	if err != nil {
		log.Fatal("usersテーブル作成エラー:", err)
	}

	// ガチャの履歴を保存するテーブルを作成
	historyTable := `
	CREATE TABLE IF NOT EXISTS history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		uid TEXT,
		rarity TEXT,
		character TEXT
	);`
	_, err = userDB.Exec(historyTable)
	if err != nil {
		log.Fatal("historyテーブル作成エラー:", err)
	}
}

// セッションIDを生成する関数
func generateSessionID() string {
	b := make([]byte, 16)        // 16バイトのランダムなデータを格納するためのバイトスライスを作成
	crand.Read(b)                // バイトスライスにランダムなデータを埋める
	return hex.EncodeToString(b) // バイトスライスを16進数の文字列に変換して返す
}

// CookieからセッションIDを取得、無ければ新規発行してCookieに保存する関数
func getOrCreateSession(w http.ResponseWriter, r *http.Request) string {
	// リクエストから "session_id" という名前のCookieを探す
	cookie, err := r.Cookie("session_id")

	// すでにCookieを持っている場合は、そのIDをそのまま返す
	if err == nil {
		return cookie.Value
	}

	// 新しいセッションIDを生成
	newID := generateSessionID()

	// ブラウザにCookieを保存させるための設定
	newCookie := &http.Cookie{
		Name:     "session_id",
		Value:    newID,
		Path:     "/",                 // サイト内の全ページでこのCookieを有効にする
		HttpOnly: true,                // JavaScriptからCookieを盗まれるのを防ぐセキュリティ設定
		MaxAge:   oneDay * cookieDays, // 有効期限（秒数）
	}

	// レスポンスにCookieを入れる
	http.SetCookie(w, newCookie)

	return newID
}

// ユーザーIDからデータを取得する関数
func getUserData(uid string) *UserData {
	user := &UserData{}
	var isGuaranteed int // SQLiteの整数(0か1)を安全に受け取るための一時変数

	// 1. カウンター情報の取得
	row := userDB.QueryRow("SELECT star4_limit_counter, star5_limit_counter, is_next_pickup_guaranteed FROM users WHERE uid = ?", uid)
	err := row.Scan(&user.Star4LimitCounter, &user.Star5LimitCounter, &isGuaranteed)

	if err == sql.ErrNoRows {
		// データが無い（新規ユーザー）の場合は、初期値をDBに登録
		userDB.Exec("INSERT INTO users (uid) VALUES (?)", uid)
	}

	// 整数(1)なら true、それ以外(0)なら false に変換して構造体にセット
	user.IsNextPickupGuaranteed = (isGuaranteed == 1)

	// 2. 履歴の取得（最新の50件を古い順に取り出すSQLのトリック）
	rows, err := userDB.Query("SELECT rarity, character FROM (SELECT id, rarity, character FROM history WHERE uid = ? ORDER BY id DESC LIMIT 50) AS sub ORDER BY id ASC", uid)
	if err != nil {
		log.Println("履歴取得エラー:", err)
		return user // エラーが起きたらここで中断し、サーバークラッシュを防ぐ
	}
	defer rows.Close() // 使い終わったら必ず閉じる

	for rows.Next() {
		var res GachaResult
		rows.Scan(&res.Rarity, &res.Character)
		user.GachaHistory = append(user.GachaHistory, res)
	}

	return user
}

// ガチャを引いた後、最新のカウンター状態をDBに上書き保存する
func updateUserData(uid string, user *UserData) {
	_, err := userDB.Exec("UPDATE users SET star4_limit_counter = ?, star5_limit_counter = ?, is_next_pickup_guaranteed = ? WHERE uid = ?",
		user.Star4LimitCounter, user.Star5LimitCounter, user.IsNextPickupGuaranteed, uid)
	if err != nil {
		log.Println("データ保存エラー:", err)
	}
}

// 引いたキャラクターを履歴DBに追加する
func addHistory(uid string, result GachaResult) {
	_, err := userDB.Exec("INSERT INTO history (uid, rarity, character) VALUES (?, ?, ?)",
		uid, result.Rarity, result.Character)
	if err != nil {
		log.Println("履歴保存エラー:", err)
	}
}

// ガチャの処理を行う関数
func gachaHandler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得、無ければ新規発行してブラウザに植え付ける関数を呼び出す
	uid := getOrCreateSession(w, r)
	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	// ガチャの結果を判定する関数を呼び出して、結果を取得
	result := gachaJudgment(user)

	// DBにユーザーデータを保存
	updateUserData(uid, user)
	// DBに履歴を追加
	addHistory(uid, result)

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
	// CookieからユーザーIDを取得、無ければ新規発行してブラウザに植え付ける関数を呼び出す
	uid := getOrCreateSession(w, r)
	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	var results []GachaResult
	for i := 0; i < 10; i++ {
		// ガチャの結果を判定する関数を呼び出して、結果を取得して、resultsの配列に追加
		result := gachaJudgment(user)
		results = append(results, result)

		// DBに履歴を追加
		addHistory(uid, result)

		// 履歴に追加 (50件を超えていたら、一番古い要素を切り捨てる)
		user.GachaHistory = append(user.GachaHistory, result)
		if len(user.GachaHistory) > 50 {
			// インデックス1から最後までを残す
			user.GachaHistory = user.GachaHistory[1:]
		}
	}

	// DBにユーザーデータを保存
	updateUserData(uid, user)

	// レスポンス作成
	sendGachaResponse(w, results, user)
}

// 天井カウンターを返すハンドラー
func limitHandler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得、無ければ新規発行してブラウザに植え付ける関数を呼び出す
	uid := getOrCreateSession(w, r)
	// ユーザーIDからユーザーデータを取得
	user := getUserData(uid)

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(map[string]int{
		"star4LimitCounter": star4Limit - user.Star4LimitCounter,
		"star5LimitCounter": star5Limit - user.Star5LimitCounter,
	})
}

// 履歴を返すハンドラー
func historyHandler(w http.ResponseWriter, r *http.Request) {
	// CookieからユーザーIDを取得、無ければ新規発行してブラウザに植え付ける関数を呼び出す
	uid := getOrCreateSession(w, r)
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

		randomIndex := rand.IntN(len(pickupStar4)) // ピックアップ星4キャラクターの中からランダムに選ぶ
		return GachaResult{Rarity: "星4", Character: pickupStar4[randomIndex]}
	} else {
		// 94.3%の確率で星3
		return GachaResult{Rarity: "星3", Character: star3}
	}
}

// ピックアップキャラクターの当選判定を行う関数
func pickupJudgment(user *UserData) GachaResult {
	// ピックアップキャラクターが確定している場合は、ピックアップキャラクターを返す
	if user.IsNextPickupGuaranteed {
		user.IsNextPickupGuaranteed = false // フラグをリセット
		return GachaResult{Rarity: "星5", Character: pickupStar5}
	} else {
		// ピックアップキャラクターが確定していない場合は、50%の確率でピックアップキャラクター、50%の確率ですり抜けキャラクターを返す
		if rand.IntN(2) == 0 {
			return GachaResult{Rarity: "星5", Character: pickupStar5}
		} else {
			user.IsNextPickupGuaranteed = true           // 次のガチャでピックアップキャラクターが確定するようにフラグをセット
			randomIndex := rand.IntN(len(standardStar5)) // すり抜けキャラクターの中からランダムに選ぶ
			return GachaResult{Rarity: "星5", Character: standardStar5[randomIndex]}
		}
	}
}
