package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

// GachaResult はガチャの結果を入れる「構造体」です（C/C++のstructと同じです）
type GachaResult struct {
	Rarity    string `json:"rarity"`
	Character string `json:"character"`
}

// トップページ（ / ）にアクセスされた時の処理
func rootHandler(w http.ResponseWriter, r *http.Request) {
	// Fprintf を使うと、ブラウザの画面（w）に直接文字を出力できます
	fmt.Fprintf(w, "<h1>ガチャAPI稼働中！</h1>")
	fmt.Fprintf(w, "<p>ガチャを引くにはURLの末尾に <b>/gacha</b> を付けてください。</p>")
}

// ガチャの処理を行う関数
func gachaHandler(w http.ResponseWriter, r *http.Request) {
	// 0〜99の乱数を生成
	roll := rand.Intn(100)
	var result GachaResult

	// 確率の判定（C/C++のif文と全く同じです）
	if roll < 5 {
		// 5%の確率で星5
		result = GachaResult{Rarity: "星5", Character: "ゼーレ"}
	} else if roll < 20 {
		// 15%の確率で星4
		result = GachaResult{Rarity: "星4", Character: "丹恒"}
	} else {
		// 残り80%の確率で星3
		result = GachaResult{Rarity: "星3", Character: "光円錐"}
	}

	// 結果をJSON（Webで標準的なデータ形式）に変換して、リクエスト元に返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(result)
}

func main() {
	// 「/」というURLにアクセスが来たら、rootHandlerを実行するように設定
	http.HandleFunc("/", rootHandler)
	// 「/gacha」というURLにアクセスが来たら、gachaHandlerを実行するように設定
	http.HandleFunc("/gacha", gachaHandler)

	fmt.Println("サーバーを起動しました！ ブラウザで http://localhost:8080/gacha にアクセスしてください。")
	fmt.Println("終了するにはターミナルで Ctrl + C を押します。")

	// ポート8080でサーバーを起動（ゲームのメインループのように、ここでアクセスを待ち続けます）
	http.ListenAndServe(":8080", nil)
}
