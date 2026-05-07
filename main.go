package main // エントリーポイント

// ライブラリのインポート
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

// チャの結果を入れる構造体 変数名の先頭が大文字にすると外部からアクセスできる（JSONに変換するために必要）
type GachaResult struct {
	Rarity    string `json:"rarity"`    // レアリティ (`json:"rarity"`は、JSONに変換するときのキー名)
	Character string `json:"character"` // キャラクター名 (`json:"character"`は、JSONに変換するときのキー名)
}

// メイン関数
func main() {
	// "static"フォルダの中身（HTML, CSS, JS）を、そのままブラウザに公開する設定
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// ガチャのエンドポイントを設定
	http.HandleFunc("/gacha", gachaHandler)     // 単発ガチャのエンドポイント /gacha
	http.HandleFunc("/gacha10", gacha10Handler) // 10連ガチャのエンドポイント /gacha10

	// サーバー起動のメッセージを表示
	fmt.Println("サーバーを起動しました！ ブラウザで http://localhost:8080 にアクセスしてください。")
	fmt.Println("終了するにはターミナルで Ctrl + C を押します。")

	// ポート8080でサーバーを起動（ゲームのメインループのように、ここでアクセスを待ち続けます）
	http.ListenAndServe(":8080", nil)
}

// ガチャの処理を行う関数
func gachaHandler(w http.ResponseWriter, r *http.Request) {
	var result GachaResult

	// 0〜99の乱数を生成
	roll := rand.Intn(100)

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

	// 結果をJSONに変換して、リクエスト元に返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(result)
}

// 10連ガチャの処理を行う関数
func gacha10Handler(w http.ResponseWriter, r *http.Request) {
	var results []GachaResult

	for i := 0; i < 10; i++ {
		var result GachaResult

		// 0〜99の乱数を生成
		roll := rand.Intn(100)

		// 確率の判定
		if roll < 5 {
			// 5%の確率で星5
			result = GachaResult{Rarity: "星5", Character: "ゼーレ"}
		} else if roll < 20 {
			// 15%の確率で星4
			result = GachaResult{Rarity: "星4", Character: "丹恒"}
		} else {
			// 80%の確率で星3
			result = GachaResult{Rarity: "星3", Character: "光円錐"}
		}

		// 結果を配列に追加
		results = append(results, result)
	}

	// 結果をJSONに変換して、リクエスト元に返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(results)
}
