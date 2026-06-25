// 新規登録処理
async function registerUser() {
    const user = document.getElementById("username").value;
    const pass = document.getElementById("password").value;
    const msgArea = document.getElementById("auth-message");

    try {
        const response = await fetch("/register", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: user, password: pass })
        });
        const text = await response.text();
        msgArea.style.color = response.ok ? "blue" : "red";
        msgArea.innerText = text;
    } catch (error) {
        msgArea.innerText = "通信エラーが発生しました";
    }
}

// ログイン処理
async function loginUser() {
    const user = document.getElementById("username").value;
    const pass = document.getElementById("password").value;
    const msgArea = document.getElementById("auth-message");

    try {
        const response = await fetch("/login", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: user, password: pass })
        });
        const text = await response.text();
        
        if (response.ok) {
            // ログイン成功時：ログイン画面を隠して、ガチャ画面を表示する！
            document.getElementById("auth-area").style.display = "none";
            document.getElementById("gacha-app").style.display = "block";
            
            // ログインした人のデータを読み込む
            loadHistoryFromServer();
            loadLimitFromServer();
        } else {
            // エラー時
            msgArea.style.color = "red";
            msgArea.innerText = text;
        }
    } catch (error) {
        msgArea.innerText = "通信エラーが発生しました";
    }
}

// ページが読み込まれたときに、天井カウンターと履歴をサーバーから取得してUIを更新する
window.onload = function() {

};

// 単発ガチャボタンが押された時の処理
async function drawGacha() {
    // 結果表示エリアを「通信中...」にして、ユーザーに待ってもらう
    const resultArea = document.getElementById("result-area");
    resultArea.innerHTML = "通信中...";

    try {
        // サーバーの /gacha エンドポイントにリクエストを送る
        const response = await fetch("/gacha");
        // 石不足などのエラーをキャッチして、ユーザーに知らせる
        if (!response.ok) {
            const errorText = await response.text();
            alert(errorText);
            resultArea.innerHTML = "キャンセルされました";
            return;
        }
        const data = await response.json();

        // 結果を表示する
        const res = data.results[0];
        let colorClass = res.rarity === "星5" ? "star5" : (res.rarity === "星4" ? "star4" : "star3");
        resultArea.innerHTML = `<span class="${colorClass}">【${res.rarity}】 ${res.character}</span>`;
        
        // UI更新
        updatePityUI(data.pity5Star, data.pity4Star, data.stones);
        updateHistoryUI(res.character, res.rarity);
    } catch (error) {
        // エラーの具体的な中身（error.message）を画面に出す！
        resultArea.innerHTML = "エラー詳細: " + error.message;
        console.error("通信エラー:", error);
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
        // 石不足などのエラーをキャッチして、ユーザーに知らせる
        if (!response.ok) {
            const errorText = await response.text();
            alert(errorText);
            resultArea.innerHTML = "キャンセルされました";
            return;
        }
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
        updatePityUI(data.pity5Star, data.pity4Star, data.stones);
    } catch (error) {
        // エラーの具体的な中身（error.message）を画面に出す！
        resultArea.innerHTML = "エラー詳細: " + error.message;
        console.error("通信エラー:", error);
    }
}

// 石を追加するボタンが押された時の処理（デバッグ用）
async function addStones() {
    try {
        // サーバーの /add_stones エンドポイントにPOSTリクエストを送る
        const response = await fetch("/add_stones", {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            }
        });
        if (!response.ok) {
            const errorText = await response.text();
            alert("購入に失敗しました: " + errorText);
            return;
        }
        const data = await response.json();

        // 石の所持数を更新する
        document.getElementById("stone-count").innerText = data.stones;
        alert("石を1000個購入しました！");
    } catch (error) {
        // エラーの具体的な中身（error.message）を画面に出す！
        console.error("石の追加に失敗:", error);
    }
}

// 天井カウンターをサーバーから取得してUIを更新する関数
async function loadLimitFromServer() {
    try {
        // サーバーの /limit エンドポイントにリクエストを送る
        const response = await fetch("/limit");
        const data = await response.json();
        // 天井カウンターを更新する
        updatePityUI(data.star5LimitCounter, data.star4LimitCounter, data.stones);
    } catch (error) {
        // エラーの具体的な中身（error.message）を画面に出す！
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
function updatePityUI(pity5, pity4, stones) {
    document.getElementById("pity5-count").innerText = pity5;
    document.getElementById("pity4-count").innerText = pity4;
    if(stones !== undefined) {
        document.getElementById("stone-count").innerText = stones;
    }
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