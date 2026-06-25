package main // エントリーポイント

// ライブラリのインポート
import (
	"fmt"      // フォーマット用 (文字列の整形など)
	"net/http" // HTTPサーバーの構築に使用
)

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

	// 石を追加するエンドポイント（デバッグ用）
	http.HandleFunc("/add_stones", addStonesHandler)

	// 履歴だけを取得するエンドポイント
	http.HandleFunc("/history", historyHandler)
	// 天井カウンターを取得するエンドポイント
	http.HandleFunc("/limit", limitHandler)

	// 認証用エンドポイント
	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/check_auth", checkAuthHandler)

	// 管理者用エンドポイント
	http.HandleFunc("/admin/delete_history", adminDeleteHistoryHandler)
	http.HandleFunc("/admin/add_stones", adminAddStonesHandler)

	// サーバー起動のメッセージを表示
	fmt.Println("サーバーを起動しました！ ブラウザで http://localhost:8080 にアクセスしてください。")
	fmt.Println("終了するにはターミナルで Ctrl + C を押します。")

	// ポート8080でサーバーを起動（ゲームのメインループのように、ここでアクセスを待ち続けます）
	http.ListenAndServe(":8080", nil)
}
