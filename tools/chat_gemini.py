#!/usr/bin/env python3
import json
import os
import subprocess
import sys
import urllib.request
import urllib.error

PROJECT_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
MCP_SERVER_BIN = os.path.join(PROJECT_ROOT, "bin", "mcp-server")

class MCPServerClient:
    def __init__(self, bin_path):
        if not os.path.exists(bin_path):
            print(f"⚠️  {bin_path} が見つかりません。ビルドします...")
            subprocess.run(["go", "build", "-o", bin_path, "cmd/mcp-server/main.go"], cwd=PROJECT_ROOT, check=True)
            print("✅ ビルド完了")
        
        self.proc = subprocess.Popen(
            [bin_path],
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.DEVNULL,
            text=True,
            bufsize=1
        )
        self.req_id = 1
        self._initialize()

    def _send_request(self, method, params=None):
        payload = {
            "jsonrpc": "2.0",
            "id": self.req_id,
            "method": method,
            "params": params or {}
        }
        self.req_id += 1
        self.proc.stdin.write(json.dumps(payload) + "\n")
        self.proc.stdin.flush()

        line = self.proc.stdout.readline()
        if not line:
            raise RuntimeError("MCP server closed stdout unexpectedly")
        return json.loads(line)

    def _initialize(self):
        self._send_request("initialize", {
            "protocolVersion": "2024-11-05",
            "capabilities": {},
            "clientInfo": {"name": "chat_gemini", "version": "1.0"}
        })

    def list_tools(self):
        res = self._send_request("tools/list")
        return res.get("result", {}).get("tools", [])

    def call_tool(self, name, arguments):
        res = self._send_request("tools/call", {
            "name": name,
            "arguments": arguments
        })
        content_list = res.get("result", {}).get("content", [])
        output = []
        for c in content_list:
            if c.get("type") == "text":
                output.append(c.get("text", ""))
        return "\n".join(output)

    def close(self):
        if self.proc:
            self.proc.terminate()

def mcp_tools_to_gemini_declarations(mcp_tools):
    declarations = []
    for tool in mcp_tools:
        dec = {
            "name": tool["name"],
            "description": tool.get("description", ""),
        }
        input_schema = tool.get("inputSchema", {})
        if input_schema and input_schema.get("properties"):
            properties = {}
            for prop_name, prop_info in input_schema["properties"].items():
                prop_type = prop_info.get("type", "string").upper()
                if prop_type == "STRING":
                    p_type = "STRING"
                elif prop_type in ("INTEGER", "NUMBER"):
                    p_type = "INTEGER"
                elif prop_type == "BOOLEAN":
                    p_type = "BOOLEAN"
                else:
                    p_type = "STRING"
                
                properties[prop_name] = {
                    "type": p_type,
                    "description": prop_info.get("description", "")
                }
            dec["parameters"] = {
                "type": "OBJECT",
                "properties": properties,
                "required": input_schema.get("required", [])
            }
        declarations.append(dec)
    return declarations

def list_available_models(api_key):
    url = f"https://generativelanguage.googleapis.com/v1beta/models?key={api_key}"
    req = urllib.request.Request(url)
    try:
        with urllib.request.urlopen(req) as res:
            data = json.loads(res.read().decode("utf-8"))
            models = []
            for m in data.get("models", []):
                name = m.get("name", "").replace("models/", "")
                methods = m.get("supportedGenerationMethods", [])
                if "generateContent" in methods:
                    models.append(name)
            return models
    except Exception as e:
        print(f"⚠️  モデル一覧の取得中に警告: {e}")
        return []

def normalize_model_name(model, available_models):
    # もしAPIキーでモデル一覧が取得できた場合
    if available_models:
        # 完全一致
        if model in available_models:
            return model
        
        # 部分一致検索
        for am in available_models:
            if model in am or am in model:
                return am
        
        # Pro系を求めている場合
        if "pro" in model:
            for am in available_models:
                if "pro" in am:
                    return am

        # Flash系
        for am in available_models:
            if "2.0-flash" in am or "1.5-flash" in am:
                return am
        
        return available_models[0]

    # モデル一覧取得が失敗した場合のハードコードエイリアス
    hardcoded_map = {
        "gemini-1.5-pro": "gemini-1.5-pro-002",
        "gemini-2.0-pro": "gemini-2.0-pro-exp-02-05",
        "gemini-1.5-flash": "gemini-1.5-flash-002",
    }
    return hardcoded_map.get(model, "gemini-2.0-flash")

def call_gemini_api(api_key, contents, tools_declarations, model="gemini-2.0-flash", max_retries=3, available_models=None):
    real_model = normalize_model_name(model, available_models)
    url = f"https://generativelanguage.googleapis.com/v1beta/models/{real_model}:generateContent?key={api_key}"
    payload = {
        "contents": contents,
        "tools": [{"functionDeclarations": tools_declarations}],
        "systemInstruction": {
            "parts": [{
                "text": "あなたはユーザーの過去のツイート（Twilog / Xアーカイブ）データを検索・分析するパーソナルAIアシスタントです。過去の投稿内容や日付に関する質問を受けた場合は、必ず用意されているツール (search_tweets, get_tweets_by_date, get_latest_tweets) を呼び出して検索し、そのデータに基づいて親切かつ正確に回答してください。"
            }]
        }
    }

    import time

    for attempt in range(1, max_retries + 1):
        req = urllib.request.Request(
            url,
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"}
        )

        try:
            with urllib.request.urlopen(req) as res:
                return json.loads(res.read().decode("utf-8"))
        except urllib.error.HTTPError as e:
            err_body = e.read().decode("utf-8")
            if e.code == 429:
                wait_sec = attempt * 3
                if attempt < max_retries:
                    print(f"⏳ レート制限 (429 Too Many Requests) に達しました。{wait_sec}秒待機して再試行します ({attempt}/{max_retries})...")
                    time.sleep(wait_sec)
                    continue
                else:
                    # 2.0-flash が上限に達した場合は 1.5-flash や 2.0-flash-lite をフォールバック試行
                    if model != "gemini-1.5-flash":
                        print("🔄 gemini-1.5-flash にフォールバックして再試行します...")
                        return call_gemini_api(api_key, contents, tools_declarations, model="gemini-1.5-flash", max_retries=2)
            
            print(f"\n❌ Gemini API エラー: {e.code}\n{err_body}")
            sys.exit(1)

def main():
    api_key = os.environ.get("GEMINI_API_KEY")
    user_requested_model = os.environ.get("GEMINI_MODEL", "gemini-2.0-flash")
    if not api_key:
        print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        print("⚠️  GEMINI_API_KEY 環境変数が設定されていません。")
        print("以下のように API キーを設定して実行してください：")
        print("  export GEMINI_API_KEY=\"your_api_key_here\"")
        print("  python3 tools/chat_gemini.py")
        print("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
        sys.exit(1)

    available_models = list_available_models(api_key)
    if available_models:
        print(f"📋 利用可能モデル一覧: {', '.join(available_models[:5])}...")
    target_model = normalize_model_name(user_requested_model, available_models)

    print("🚀 MCP サーバーを起動中...")
    mcp_client = MCPServerClient(MCP_SERVER_BIN)
    
    try:
        mcp_tools = mcp_client.list_tools()
        gemini_tools = mcp_tools_to_gemini_declarations(mcp_tools)
        
        print(f"✨ Gemini (モデル: {target_model}) 連携対話チャット (Twilog Archive MCP)")
        if user_requested_model != target_model:
            print(f"ℹ️  指定された '{user_requested_model}' を '{target_model}' に調整しました。")
        print("ツイートに関する質問を入力してください（終了するには 'exit' や 'quit' と入力）。")
        print("------------------------------------------------------------")

        contents_history = []

        while True:
            try:
                user_input = input("\n💬 あなた: ").strip()
            except (KeyboardInterrupt, EOFError):
                print("\n終了します。")
                break

            if not user_input:
                continue
            if user_input.lower() in ("exit", "quit"):
                print("バイバイ！👋")
                break

            contents_history.append({
                "role": "user",
                "parts": [{"text": user_input}]
            })

            # Gemini API 呼出
            res = call_gemini_api(api_key, contents_history, gemini_tools, model=target_model, available_models=available_models)
            candidates = res.get("candidates", [])
            if not candidates:
                print("🤖 Gemini: 応答が得られませんでした。")
                continue

            first_candidate = candidates[0]
            content = first_candidate.get("content", {})
            parts = content.get("parts", [])

            # Function Call チェック
            function_call = None
            for p in parts:
                if "functionCall" in p:
                    function_call = p["functionCall"]
                    break

            if function_call:
                fn_name = function_call["name"]
                fn_args = function_call.get("args", {})
                print(f"🔍 ツール実行中: {fn_name}({fn_args}) ...")

                # MCP ツール実行
                tool_result_text = mcp_client.call_tool(fn_name, fn_args)

                # 履歴に追加
                contents_history.append(content)
                contents_history.append({
                    "role": "user",
                    "parts": [{
                        "functionResponse": {
                            "name": fn_name,
                            "response": {"output": tool_result_text}
                        }
                    }]
                })

                # 再度 Gemini を呼び出して最終回答を生成
                res_after_tool = call_gemini_api(api_key, contents_history, gemini_tools, model=target_model, available_models=available_models)
                final_parts = res_after_tool["candidates"][0]["content"].get("parts", [])
                final_text = "".join(p.get("text", "") for p in final_parts)
                
                contents_history.append({
                    "role": "model",
                    "parts": [{"text": final_text}]
                })
                print(f"\n🤖 Gemini: {final_text}")
            else:
                final_text = "".join(p.get("text", "") for p in parts)
                contents_history.append({
                    "role": "model",
                    "parts": [{"text": final_text}]
                })
                print(f"\n🤖 Gemini: {final_text}")

    finally:
        mcp_client.close()

if __name__ == "__main__":
    main()
