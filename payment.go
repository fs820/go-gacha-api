package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
)

// セッションIDを生成する関数
func generateOrderID() string {
	b := make([]byte, 16)                   // 16バイトのランダムなデータを格納するためのバイトスライスを作成
	rand.Read(b)                            // バイトスライスにランダムなデータを埋める
	return "ORDER_" + hex.EncodeToString(b) // バイトスライスを16進数の文字列に変換して返す
}

// ユーザーが石を購入したいと仮オーダーを出す関数
func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	// POSTリクエストのみ
	if r.Method != http.MethodPost {
		http.Error(w, "POSTのみ", http.StatusMethodNotAllowed)
		return
	}

	// ログインが必要
	uid, err := getSession(r)
	if err != nil {
		http.Error(w, "未ログイン", http.StatusUnauthorized)
		return
	}

	// ランダムな注文番号を生成
	orderID := generateOrderID()

	// DBに未払い('pending')状態で記録（3000個の石を購入したい）
	_, err = userDB.Exec("INSERT INTO orders (order_id, uid, amount, status) VALUES (?, ?, ?, 'pending')", orderID, uid, 3000)
	if err != nil {
		http.Error(w, "注文の作成に失敗しました", http.StatusInternalServerError)
		return
	}

	// ユーザーに注文番号だけを返す（この時点ではまだ石は増えない！）
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"order_id": orderID})
}

// 決済会社の決済完了通知（Webhook）を受け取るハンドラー
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	// POSTリクエストのみ
	if r.Method != http.MethodPost {
		http.Error(w, "POSTのみ", http.StatusMethodNotAllowed)
		return
	}

	// 決済会社（Stripe等）からの通信であることを証明する署名を確認する
	signature := r.Header.Get("Stripe-Signature")
	if signature != "secret_stripe_key_123" {
		http.Error(w, "不正な署名です (Unauthorized Webhook)", http.StatusUnauthorized)
		return
	}

	// 注文番号を読み取る
	var req struct {
		OrderID string `json:"order_id"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// DBで決済の完了と石の付与をする
	err := completeOrderTx(req.OrderID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 成功メッセージ
	w.Write([]byte("Webhook処理成功：石を付与しました"))
}
