package main // エントリーポイント

// ライブラリのインポート
import (
	"fmt"
	"net/http" // HTTPサーバーの構築に使用
	"strconv"
)

// 管理者専用：すべての履歴を削除するエンドポイント
func adminDeleteHistoryHandler(w http.ResponseWriter, r *http.Request) {
	// パスワード認証（URLのクエリパラメータを確認）
	password := r.URL.Query().Get("pass")
	if password != "supersecret" { // 合言葉が違う場合は弾く
		http.Error(w, "権限がありません (Unauthorized)", http.StatusUnauthorized)
		return
	}

	// 履歴テーブルのデータをすべて削除
	_, err := userDB.Exec("DELETE FROM history")
	if err != nil {
		http.Error(w, "データベースの削除に失敗しました", http.StatusInternalServerError)
		return
	}

	// AUTOINCREMENTのid番号を1にリセットする
	userDB.Exec("DELETE FROM sqlite_sequence WHERE name='history'")

	// 成功メッセージ
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte("すべてのガチャ履歴を正常に削除しました！"))
}

// 管理者専用：指定したユーザーに石を付与するエンドポイント
func adminAddStonesHandler(w http.ResponseWriter, r *http.Request) {
	// パスワード認証
	password := r.URL.Query().Get("pass")
	if password != "supersecret" {
		http.Error(w, "権限がありません (Unauthorized)", http.StatusUnauthorized)
		return
	}

	// クエリパラメータの取得
	targetUID := r.URL.Query().Get("uid")
	amountStr := r.URL.Query().Get("amount")
	if targetUID == "" || amountStr == "" {
		http.Error(w, "uidとamountを指定してください。 例: ?pass=supersecret&uid=xxx&amount=1000", http.StatusBadRequest)
		return
	}

	// 文字列のamountを整数(int)に変換
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		http.Error(w, "amountは数字で指定してください", http.StatusBadRequest)
		return
	}

	// 石を追加
	err = addStonesTx(targetUID, amount)
	if err != nil {
		http.Error(w, "石の追加に失敗しました", http.StatusInternalServerError)
		return
	}

	// 成功メッセージ (fmt.Sprintf を使って文字列の中に変数を埋め込む)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Write([]byte(fmt.Sprintf("ユーザー[%s]に石を%d個追加しました！", targetUID, amount)))
}
