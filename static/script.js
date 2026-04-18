async function drawGacha() {
    const resultArea = document.getElementById("result-area");
    resultArea.innerHTML = "通信中...";

    try {
        const response = await fetch("/gacha");
        const data = await response.json();

        let colorClass = "star3";
        if (data.rarity === "星5") colorClass = "star5";
        if (data.rarity === "星4") colorClass = "star4";

        resultArea.innerHTML = `<span class="${colorClass}">【${data.rarity}】 ${data.character} が出ました！</span>`;
    } catch (error) {
        resultArea.innerHTML = "エラーが発生しました";
    }
}