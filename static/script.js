// ページが読み込まれたときに、天井カウンターと履歴をサーバーから取得してUIを更新する
window.onload = function() {
    loadHistoryFromServer();
    loadLimitFromServer();
};

// 単発ガチャボタンが押された時の処理
async function drawGacha() {
    // 結果表示エリアを「通信中...」にして、ユーザーに待ってもらう
    const resultArea = document.getElementById("result-area");
    resultArea.innerHTML = "通信中...";

    try {
        // サーバーの /gacha エンドポイントにリクエストを送る
        const response = await fetch("/gacha");
        const data = await response.json();

        // 結果を表示する
        const res = data.results[0];
        let colorClass = res.rarity === "星5" ? "star5" : (res.rarity === "星4" ? "star4" : "star3");
        resultArea.innerHTML = `<span class="${colorClass}">【${res.rarity}】 ${res.character}</span>`;
        
        // UI更新
        updatePityUI(data.pity5Star, data.pity4Star);
        updateHistoryUI(res.character, res.rarity);
    } catch (error) {
        resultArea.innerHTML = "エラーが発生しました";
    }
}

// 10連ガチャボタンが押された時の処理
async function drawGacha10() {
    // 結果表示エリアを「通信中...」にして、ユーザーに待ってもらう
    const resultArea = document.getElementById("result-area");
    resultArea.innerHTML = "通信中...";

    try {
        // サーバーの /gacha10 エンドポイントにリクエストを送る
        const response = await fetch("/gacha10");
        const data = await response.json();

        resultArea.innerHTML = "";
        // 10個の結果をループ
        data.results.forEach(res => {
            // レアリティに応じて色を変える
            let colorClass = res.rarity === "星5" ? "star5" : (res.rarity === "星4" ? "star4" : "star3");
            resultArea.innerHTML += `<div class="${colorClass}">【${res.rarity}】 ${res.character}</div>`;

            // 履歴UIも更新
            updateHistoryUI(res.character, res.rarity);
        });

        // UI更新
        updatePityUI(data.pity5Star, data.pity4Star);
    } catch (error) {
        resultArea.innerHTML = "エラーが発生しました";
    }
}

// 天井カウンターをサーバーから取得してUIを更新する関数
async function loadLimitFromServer() {
    try {
        // サーバーの /limit エンドポイントにリクエストを送る
        const response = await fetch("/limit");
        const data = await response.json();
        // 天井カウンターを更新する
        updatePityUI(data.star5LimitCounter, data.star4LimitCounter);
    } catch (error) {
        console.error("天井カウンターの取得に失敗:", error);
    }
}

// 履歴をサーバーから取得してUIを更新する関数
async function loadHistoryFromServer() {
    try {
        // サーバーの /history エンドポイントにリクエストを送る
        const response = await fetch("/history");
        const data = await response.json();
        // 履歴表示を更新する
        const historyArea = document.getElementById("history-area");
        historyArea.innerHTML = "";
        data.reverse().forEach(item => {
            let colorClass = item.rarity === "星5" ? "star5" : (item.rarity === "星4" ? "star4" : "star3");
            historyArea.innerHTML += `<div class="history-item ${colorClass}">【${item.rarity}】 ${item.character}</div>`;
        });
    } catch (error) {
        console.error("履歴の取得に失敗:", error);
    }
}

// 天井表示を更新する共通関数
function updatePityUI(pity5, pity4) {
    document.getElementById("pity5-count").innerText = pity5;
    document.getElementById("pity4-count").innerText = pity4;
}

// 履歴表示を更新する共通関数
function updateHistoryUI(character, rarity) {
    const historyArea = document.getElementById("history-area");
    let colorClass = "star3";
    if (rarity === "星5") colorClass = "star5";
    if (rarity === "星4") colorClass = "star4";

    // 新しい履歴を一番上に追加
    const item = `<div class="history-item ${colorClass}">【${rarity}】 ${character}</div>`;
    historyArea.innerHTML = item + historyArea.innerHTML;
    // 履歴が50件を超えたら、古いものを削除
    if (historyArea.children.length > 50) {
        historyArea.removeChild(historyArea.lastChild);
    }
}