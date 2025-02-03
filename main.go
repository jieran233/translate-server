package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// 翻译结果响应结构体
type TranslateResponse struct {
	Sentences []struct {
		Trans   string `json:"trans"`
		Orig    string `json:"orig"`
		Backend int    `json:"backend"`
	} `json:"sentences"`
	Src   string                 `json:"src"`
	Spell map[string]interface{} `json:"spell"`
}

// Google 翻译请求的 URL 模板
const translateURL = "https://translate.google.com/m?hl=en&sl=%s&tl=%s&q=%s"

// 提取翻译内容的函数
func extractTranslation(body []byte) (string, error) {
	// 解析 HTML
	doc, err := html.Parse(strings.NewReader(string(body)))
	if err != nil {
		return "", err
	}

	// 深度遍历 HTML 树，寻找 class="result-container" 的 div
	var f func(*html.Node)
	var translation string

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "div" {
			// 找到 class="result-container" 的 div
			for _, a := range n.Attr {
				if a.Key == "class" && a.Val == "result-container" {
					translation = n.FirstChild.Data
				}
			}
		}
		// 继续遍历子节点
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	// 如果没有找到翻译内容，返回错误
	if translation == "" {
		return "", fmt.Errorf("failed to extract translation")
	}

	return translation, nil
}

// 翻译请求的处理函数
func TranslateHandler(w http.ResponseWriter, r *http.Request) {
	// 获取请求中的 URL 参数
	q := r.URL.Query().Get("q")
	sl := r.URL.Query().Get("sl")
	tl := r.URL.Query().Get("tl")

	// 如果缺少 q, sl, 或 tl 参数，返回 400 错误
	if q == "" || sl == "" || tl == "" {
		http.Error(w, "Missing required parameters: q, sl, tl", http.StatusBadRequest)
		return
	}

	// URL 编码待翻译文本
	qEncoded := url.QueryEscape(q)

	// 构建请求 URL
	requestURL := fmt.Sprintf(translateURL, sl, tl, qEncoded)

	// 发起请求到 Google 翻译接口
	resp, err := http.Get(requestURL)
	if err != nil {
		http.Error(w, "Failed to fetch translation from Google", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// 读取响应的 HTML 内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read response body", http.StatusInternalServerError)
		return
	}

	// 提取翻译结果
	translatedText, err := extractTranslation(body)
	if err != nil {
		http.Error(w, "Failed to parse translation", http.StatusInternalServerError)
		return
	}

	// 构造响应的数据
	result := TranslateResponse{
		Sentences: []struct {
			Trans   string `json:"trans"`
			Orig    string `json:"orig"`
			Backend int    `json:"backend"`
		}{
			{
				Trans:   translatedText,
				Orig:    q,
				Backend: 10,
			},
		},
		Src:   sl,
		Spell: make(map[string]interface{}),
	}

	// 设置响应头为 JSON 格式
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// 返回 JSON 格式的翻译响应
	err = json.NewEncoder(w).Encode(result)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func main() {
	// 使用 flag 包解析命令行参数
	port := flag.String("port", "5000", "Port to run the server on")
	flag.Parse()

	// 设置路由和处理函数
	http.HandleFunc("/translate_a/single", TranslateHandler)

	// 启动服务器，使用指定的端口
	fmt.Printf("Server started on :%s\n", *port)
	if err := http.ListenAndServe(":"+*port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
	}
}
