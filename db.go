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
		username TEXT UNIQUE,
		password_hash,
		stones INTEGER DEFAULT 30000,
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

	// キャラクターデータを保存するテーブルを作成
	charactersTable := `
	CREATE TABLE IF NOT EXISTS characters (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		rarity TEXT,
		is_pickup BOOLEAN DEFAULT 0
	);`
	_, err = userDB.Exec(charactersTable)
	if err != nil {
		log.Fatal("charactersテーブル作成エラー:", err)
	}

	// もしキャラクターテーブルが空の場合初期化
	var count int
	userDB.QueryRow("SELECT COUNT(*) FROM characters").Scan(&count)
	if count == 0 {
		log.Println("キャラクターの初期データを挿入します...")
		initialData := []struct {
			name     string
			rarity   string
			isPickup bool
		}{
			{"ゼウス", "星5", true},
			{"ウラノス", "星5", false},
			{"クロノス", "星5", false},
			{"釈迦", "星5", false},
			{"シヴァ", "星5", false},
			{"ポセイドン", "星5", false},
			{"ヘラクレス", "星5", false},
			{"キリスト", "星5", false},
			{"ヨハネ", "星4", true},
			{"千手観音", "星4", true},
			{"アキレス", "星4", true},
			{"武器", "星3", false},
		}

		for _, c := range initialData {
			userDB.Exec("INSERT INTO characters (name, rarity, is_pickup) VALUES (?, ?, ?)",
				c.name, c.rarity, c.isPickup)
		}
	}
}

// ユーザーIDからデータを取得する関数
func getUserData(uid string) *UserData {
	user := &UserData{}
	var isGuaranteed int // SQLiteの整数(0か1)を安全に受け取るための一時変数

	// カウンター情報の取得
	row := userDB.QueryRow("SELECT stones, star4_limit_counter, star5_limit_counter, is_next_pickup_guaranteed FROM users WHERE uid = ?", uid)
	err := row.Scan(&user.Stones, &user.Star4LimitCounter, &user.Star5LimitCounter, &isGuaranteed)
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

// DBに新規ユーザーを登録する関数 ※パスワードはhash済文字列
func insertUser(uid string, username string, hashedPassword string) error {
	// データベースに新しいユーザーを保存
	_, err := userDB.Exec("INSERT INTO users (uid, username, password_hash) VALUES (?, ?, ?)",
		uid, username, hashedPassword)
	return err
}

// ユーザー名からDBを検索して uidとhash済パスワードを返す関数
func findUser(username string) (string, string, error) {
	var uid, hash string
	err := userDB.QueryRow("SELECT uid, password_hash FROM users WHERE username = ?", username).Scan(&uid, &hash)
	return uid, hash, err
}

// ガチャの結果を保存する関数 （トランザクション）
func saveGachaResultTx(uid string, user *UserData, results []GachaResult, cost int) error {
	// トランザクションの開始
	tx, err := userDB.Begin()
	if err != nil {
		return err
	}

	// 石を消費してカウンターを進める
	_, err = tx.Exec("UPDATE users SET stones = stones - ?, star4_limit_counter = ?, star5_limit_counter = ?, is_next_pickup_guaranteed = ? WHERE uid = ?",
		cost, user.Star4LimitCounter, user.Star5LimitCounter, user.IsNextPickupGuaranteed, uid)
	if err != nil {
		tx.Rollback() // エラーが起きたらロールバック
		return err
	}

	// ガチャの結果を履歴テーブルに保存
	for _, res := range results {
		_, err = tx.Exec("INSERT INTO history (uid, rarity, character) VALUES (?, ?, ?)", uid, res.Rarity, res.Character)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	// コミットして確定
	return tx.Commit()
}

// ガチャ石を追加する関数 （トランザクション）
func addStones(uid string, stonesToAdd int) error {
	// 石を追加する
	_, err := userDB.Exec("UPDATE users SET stones = stones + ? WHERE uid = ?", stonesToAdd, uid)
	return err
}

// DBから指定したレアリティとピックアップ条件に合うキャラクターの配列を取得する関数
func getCharactersFromDB(rarity string, isPickup bool) []string {
	var chars []string

	// boolをSQLite用の整数(0か1)に変換
	pickupInt := 0
	if isPickup {
		pickupInt = 1
	}

	// DBから検索
	rows, err := userDB.Query("SELECT name FROM characters WHERE rarity = ? AND is_pickup = ?", rarity, pickupInt)
	if err != nil {
		log.Println("キャラクター取得エラー:", err)
		return chars
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		rows.Scan(&name)
		chars = append(chars, name)
	}

	return chars
}
