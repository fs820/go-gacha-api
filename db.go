package main // エントリーポイント

// ライブラリのインポート
import (
	"database/sql" // データベース操作に使用
	"log"          // ロギングに使用

	_ "modernc.org/sqlite" // SQLiteドライバ (データベース接続)
)

const DBFilePath = "./gacha.db" // データベースファイルのパス

var userDB *sql.DB

// データベースの初期化関数
func initDB() {
	var err error
	// DBファイルを開く
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

// ユーザーIDからデータを取得する関数
func getUserData(uid string) *UserData {
	user := &UserData{}
	var isGuaranteed int // SQLiteの整数(0か1)を安全に受け取るための一時変数

	// カウンター情報の取得
	row := userDB.QueryRow("SELECT star4_limit_counter, star5_limit_counter, is_next_pickup_guaranteed FROM users WHERE uid = ?", uid)
	err := row.Scan(&user.Star4LimitCounter, &user.Star5LimitCounter, &isGuaranteed)
	if err == sql.ErrNoRows {
		// データが無い（新規ユーザー）の場合は、初期値をDBに登録
		userDB.Exec("INSERT INTO users (uid) VALUES (?)", uid)
	}

	// intからboolにする
	user.IsNextPickupGuaranteed = (isGuaranteed == 1)

	// 履歴の取得新しいものを50件取得して、古い順に並び替えるs
	rows, err := userDB.Query("SELECT rarity, character FROM (SELECT id, rarity, character FROM history WHERE uid = ? ORDER BY id DESC LIMIT 50) AS sub ORDER BY id ASC", uid)
	if err != nil {
		log.Println("履歴取得エラー:", err)
		return user // エラーが起きたらここで中断
	}
	defer rows.Close() // 使い終わったら必ず閉じる

	// 取得した履歴をUserDataのGachaHistoryに追加
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
