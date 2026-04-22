import requests

OLLAMA_URL = "http://10.186.38.170:11434/api/generate"
MODEL = "qwen2:7b"

messages = [
    {"role": "system", "content": "You are a helpful assistant."},
    {"role": "user", "content": "你好，请介绍你自己。"}
]

# 手动构造 Qwen2 chat 模板格式
text = ""
for msg in messages:
    role_tag = "system" if msg["role"] == "system" else msg["role"]
    text += f"<|im_start|>{role_tag}\n{msg['content']}<|im_end|>\n"
text += "<|im_start|>assistant\n"

print("编码后的输入文本:")
print(repr(text))
print()

# 调用 Ollama generate 接口
payload = {
    "model": MODEL,
    "prompt": text,
    "stream": False,
    "options": {
        "num_predict": 512
    }
}

resp = requests.post(OLLAMA_URL, json=payload)
result = resp.json()

print("模型的回答:")
print(result.get("response", "未获取到回答"))
