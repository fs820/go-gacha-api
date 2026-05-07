package main // エントリーポイント

// ライブラリのインポート
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

// ガチャの結果を入れる構造体 変数名の先頭が大文字にすると外部からアクセスできる（JSONに変換するために必要）
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
	// ガチャの結果を判定する関数を呼び出して、結果を取得
	result := gachaJudgment()

	// 結果をJSONに変換して、リクエスト元に返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(result)
}

// 10連ガチャの処理を行う関数
func gacha10Handler(w http.ResponseWriter, r *http.Request) {
	var results []GachaResult

	for i := 0; i < 10; i++ {
		// ガチャの結果を判定する関数を呼び出して、結果を取得して、resultsの配列に追加
		results = append(results, gachaJudgment())
	}

	// 結果をJSONに変換して、リクエスト元に返す
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(results)
}

// ガチャの結果を判定する関数
func gachaJudgment() GachaResult {
	// 0〜999の乱数を生成
	roll := rand.Intn(1000)

	// 確率の判定
	if roll < 6 {
		// 0.6%の確率で星5
		return GachaResult{Rarity: "星5", Character: "ゼーレ"}
	} else if roll < 57 {
		// 5.1%の確率で星4
		return GachaResult{Rarity: "星4", Character: "丹恒"}
	} else {
		// 94.3%の確率で星3
		return GachaResult{Rarity: "星3", Character: "光円錐"}
	}
}
