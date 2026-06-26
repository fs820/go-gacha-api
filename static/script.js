// ページが読み込まれたときに、ログイン状態を確認する
window.onload = async function() {
    try {
        // サーバーに「現在のCookieは有効ですか？」と尋ねる
        const response = await fetch("/check_auth");
        
        if (response.ok) {
            // 有効だった場合：ログイン画面を隠して、ガチャ画面を表示！
            document.getElementById("auth-area").style.display = "none";
            document.getElementById("gacha-app").style.display = "block";
            
            // ユーザーデータを読み込む
            loadHistoryFromServer();
            loadLimitFromServer();
        } else {
            // 無効だった場合（未ログイン）：何もしない（ログイン画面が出たままになる）
        }
    } catch (error) {
        console.error("認証チェックに失敗:", error);
    }
};

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
        // サーバーに「注文」を作成してもらう（ユーザーとしての操作）
        const response = await fetch("/checkout", { method: "POST" });
        if (!response.ok) {
            alert("ログインしてください");
            return;
        }
        const data = await response.json();
        const orderId = data.order_id;

        // ユーザーへの案内
        alert(`注文を作成しました！(注文番号: ${orderId})\n\n※セキュリティ上の理由により、ボタンを押しただけでは石は増えません。決済会社からのWebhook（裏通信）が必要です。`);

        // 開発者（あなた）が決済会社になりきるためのコマンドを出力
        console.log("%c【決済システム（Stripe等）シミュレーター】", "color: blue; font-size: 16px; font-weight: bold;");
        console.log("以下のコードをコピーして、このコンソールに貼り付けて実行（Enter）し、決済を完了させてください：\n\n");
        
        console.log(`fetch("/webhook/payment", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Stripe-Signature": "secret_stripe_key_123" },
    body: JSON.stringify({ order_id: "${orderId}" })
}).then(r => r.text()).then(text => {
    console.log("Webhookからの返事:", text);
    alert(text + "\\nガチャ画面をリロードして石を確認してください！");
});`);

    } catch (error) {
        console.error("注文エラー:", error);
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