from flask import Flask, request, jsonify
import requests
from urllib.parse import quote

app = Flask(__name__)

# 包装的 Google 翻译接口
TRANSLATE_URL = "https://translate.google.com/m?hl=en&sl={sl}&tl={tl}&q={q}"

@app.route('/translate_a/single', methods=['GET'])
def translate():
    # 获取请求参数
    client = request.args.get('client', 'gtx')
    dj = request.args.get('dj', '1')
    dt = request.args.get('dt', 't')
    ie = request.args.get('ie', 'UTF-8')
    q = request.args.get('q')
    sl = request.args.get('sl', 'en')  # 默认源语言为英文
    tl = request.args.get('tl', 'zh-CN')  # 默认目标语言为简体中文
    
    # URL 编码
    q_encoded = quote(q)

    # 请求 Google 翻译接口
    url = TRANSLATE_URL.format(sl=sl, tl=tl, q=q_encoded)
    response = requests.get(url)
    
    # 解析返回的 HTML，提取翻译结果
    if response.status_code == 200:
        # 提取翻译结果
        start_index = response.text.find('<div class="result-container">') + len('<div class="result-container">')
        end_index = response.text.find('</div>', start_index)
        translated_text = response.text[start_index:end_index].strip()
        
        # 构造响应数据
        result = {
            "sentences": [{
                "trans": translated_text,
                "orig": q,
                "backend": 10
            }],
            "src": sl,
            "spell": {}
        }

        # 返回 JSON 响应
        return jsonify(result)
    else:
        return jsonify({"error": "Unable to fetch translation"}), 500

if __name__ == '__main__':
    app.run(debug=True)
