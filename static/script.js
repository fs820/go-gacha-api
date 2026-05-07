// 単発ガチャボタンが押された時の処理
async function drawGacha() {
    // 結果表示エリアを「通信中...」にして、ユーザーに待ってもらう
    const resultArea = document.getElementById("result-area");
    resultArea.innerHTML = "通信中...";

    try {
        // バックエンドの単発用APIを叩く
        const response = await fetch("/gacha");
        // 返ってきたJSONをJavaScriptのオブジェクトに変換
        const data = await response.json();

        // レアリティに応じて、文字の色を変えるためのクラス名を決める
        let colorClass = "star3";
        if (data.rarity === "星5") colorClass = "star5";
        if (data.rarity === "星4") colorClass = "star4";

        // 結果表示エリアに、レアリティとキャラクター名を表示するHTMLを作る
        resultArea.innerHTML = `<span class="${colorClass}">【${data.rarity}】 ${data.character} が出ました！</span>`;
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
        // バックエンドの10連用APIを叩く
        const response = await fetch("/gacha10");
        // 今回の dataList は、10個のデータが入った「配列」になる
        const dataList = await response.json();

        // 画面の「通信中...」を一旦リセットして空にする
        resultArea.innerHTML = "";

        // 配列の中身を1つずつ取り出してループ処理
        dataList.forEach(data => {
            let colorClass = "star3";
            if (data.rarity === "星5") colorClass = "star5";
            if (data.rarity === "星4") colorClass = "star4";

            // <div>タグで改行しながら、レアリティとキャラクター名を表示するHTMLを作る
            resultArea.innerHTML += `<div class="${colorClass}">【${data.rarity}】 ${data.character}</div>`;
        });

    } catch (error) {
        resultArea.innerHTML = "エラーが発生しました";
    }
}