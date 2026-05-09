package main // エントリーポイント

// ライブラリのインポート
import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
)

// 定数の定義
const (
	// ガチャの確率を定数として定義
	probBaseStar5 = 6  // 星5の当たる基本確率（6/1000 = 0.6%）
	probBaseStar4 = 51 // 星4の当たる基本確率（51/1000 = 5.1%）

	// キャラクターの名前を定数として定義
	star5Character = "ゼーレ" // 星5のキャラクター
	star4Character = "丹恒"  // 星4のキャラクター
	star3Character = "光円錐" // 星3のキャラクター

	// 天井の回数を定数として定義
	star4Limit = 10 // 星4以上が出るまでの回数
	star5Limit = 90 // 星5が出るまでの回数

	// ソフトピティの設定
	pitySoftStart     = 74 // 確率が上がり始める回数
	softPityIncrement = 6  // 確率が上がる割合 (6%ずつ増加)
)

// --- 排出キャラクターのリスト ---
// ピックアップ星5
var pickupStar5 = "銀狼Lv999"

// すり抜け星5（7名）
var standardStar5 = []string{"銀狼", "雲璃", "アルジェンティ", "ゼーレ", "符玄", "刃", "ヴェルト"}

// ピックアップ星4（3名）
var pickupStar4 = []string{"フック", "雪衣", "ギャラガー"}

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

// 天井のカウンター
var star4LimitCounter int       // 星4以上が出るまでのカウンター
var star5LimitCounter int       // 星5が出るまでのカウンター
var isNextPickupGuaranteed bool // 次のガチャでピックアップキャラクターが確定しているかどうかのフラグ

// DBモック
var gachaHistory []GachaResult // ガチャの履歴を保存するスライス

// メイン関数
func main() {
	// "static"フォルダの中身（HTML, CSS, JS）を、そのままブラウザに公開する設定
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/", fs)

	// ガチャのエンドポイントを設定
	http.HandleFunc("/gacha", gachaHandler)     // 単発ガチャのエンドポイント /gacha
	http.HandleFunc("/gacha10", gacha10Handler) // 10連ガチャのエンドポイント /gacha10

	// 履歴だけを取得するエンドポイント
	http.HandleFunc("/history", historyHandler)

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

	// 履歴に追加
	gachaHistory = append(gachaHistory, result)

	// レスポンス作成
	sendGachaResponse(w, []GachaResult{result})
}

// 10連ガチャの処理を行う関数
func gacha10Handler(w http.ResponseWriter, r *http.Request) {
	var results []GachaResult

	for i := 0; i < 10; i++ {
		// ガチャの結果を判定する関数を呼び出して、結果を取得して、resultsの配列に追加
		result := gachaJudgment()
		results = append(results, result)

		// 履歴に追加
		gachaHistory = append(gachaHistory, result)
	}

	// レスポンス作成
	sendGachaResponse(w, results)
}

// 履歴を返すハンドラー
func historyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(gachaHistory)
}

// 共通のレスポンス送信処理
func sendGachaResponse(w http.ResponseWriter, results []GachaResult) {
	response := GachaResponse{
		Results:   results,
		Pity5Star: (star5Limit + 1) - star5LimitCounter, // あと何回か
		Pity4Star: (star4Limit + 1) - star4LimitCounter,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(response)
}

// ガチャの結果を判定する関数
func gachaJudgment() GachaResult {
	// カウンターをインクリメント
	star4LimitCounter++ // 星4以上が出るまでのカウンター
	star5LimitCounter++ // 星5が出るまでのカウンター

	// 0〜999の乱数を生成
	roll := rand.Intn(1000)

	// 星5の当たる確率（6/1000 = 0.6%）
	star5Prob := probBaseStar5

	// ソフトピティの確率上昇の判定
	if star5LimitCounter >= pitySoftStart {
		// 74連目以降は、6%ずつ確率が上昇
		star5Prob += softPityIncrement * (star5LimitCounter - (pitySoftStart - 1))
	}

	// 確率の判定
	if roll < star5Prob || star5LimitCounter >= star5Limit {
		// 0.6%の確率で星5 （もしくは、天井カウンターが90連目の場合は強制的に星5）
		star4LimitCounter = 0 // カウンターをリセット
		star5LimitCounter = 0 // カウンターをリセット

		// ピックアップキャラクターの当選判定を行う関数を呼び出す
		return pickupJudgment()
	} else if roll < (star5Prob+probBaseStar4) || star4LimitCounter >= star4Limit {
		// 5.1%の確率で星4 （もしくは、天井カウンターが10連目の場合は強制的に星4）
		star4LimitCounter = 0 // カウンターをリセット

		randomIndex := rand.Intn(len(pickupStar4)) // ピックアップ星4キャラクターの中からランダムに選ぶ
		return GachaResult{Rarity: "星4", Character: pickupStar4[randomIndex]}
	} else {
		// 94.3%の確率で星3
		return GachaResult{Rarity: "星3", Character: "光円錐"}
	}
}

// ピックアップキャラクターの当選判定を行う関数
func pickupJudgment() GachaResult {
	// ピックアップキャラクターが確定している場合は、ピックアップキャラクターを返す
	if isNextPickupGuaranteed {
		isNextPickupGuaranteed = false // フラグをリセット
		return GachaResult{Rarity: "星5", Character: pickupStar5}
	} else {
		// ピックアップキャラクターが確定していない場合は、50%の確率でピックアップキャラクター、50%の確率ですり抜けキャラクターを返す
		if rand.Intn(2) == 0 {
			return GachaResult{Rarity: "星5", Character: pickupStar5}
		} else {
			isNextPickupGuaranteed = true                // 次のガチャでピックアップキャラクターが確定するようにフラグをセット
			randomIndex := rand.Intn(len(standardStar5)) // すり抜けキャラクターの中からランダムに選ぶ
			return GachaResult{Rarity: "星5", Character: standardStar5[randomIndex]}
		}
	}
}
