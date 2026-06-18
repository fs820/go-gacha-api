package main // エントリーポイント

// ライブラリのインポート
import (
	"net/http" // HTTPサーバーの構築に使用
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
