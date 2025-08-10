const fs = require('fs');
const vm = require('vm');
const inputFile = process.argv[2];
const outputFile = process.argv[3];

if (!inputFile) {
    console.error("Usage: node parseTweets.js <tweets-partX.js>");
    process.exit(1);
}

const code = fs.readFileSync(inputFile, 'utf-8');

// 実行環境（windowオブジェクトを用意）
const context = {
    window: {
        YTD: {
            tweet_headers: {},
            tweets: {},
            like: {},
        }
    }
};
vm.createContext(context);

try {
    vm.runInContext(code, context);
} catch (err) {
    console.error("vm 実行エラー:", err);
    process.exit(1);
}

// YTDカテゴリとパート名を取得（tweets/likes/etc）
const match = code.match(/window\.YTD\.(\w+)\.(\w+)/);
if (!match) {
    console.error("Invalid file format");
    process.exit(1);
}

const [_, category, part] = match;

// データ取得
const data = context.window.YTD[category][part];

// JSON出力
fs.writeFileSync(outputFile, JSON.stringify(data, null, 2), 'utf-8');
