package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

const (
	signInURL = "https://bbs.topfeel.com/api/gift/day_sign"
	referer   = "https://bbs.topfeel.com/h5/"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"
	secCHUA   = `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Printf("警告: 未找到 .env 文件或加载失败: %v\n", err)
	}

	token := os.Getenv("TOPFEEL_TOKEN")
	if token == "" {
		fmt.Println("错误: 未设置 TOPFEEL_TOKEN 环境变量")
		os.Exit(1)
	}

	now := time.Now().UnixMilli()
	newTime := now + int64(rand.Intn(4)+3)*1000

	body := map[string]interface{}{
		"oldtime": now,
		"newtime": newTime,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		fmt.Printf("序列化请求体失败: %v\n", err)
		os.Exit(1)
	}

	req, err := http.NewRequest("POST", signInURL, bytes.NewReader(bodyBytes))
	if err != nil {
		fmt.Printf("创建请求失败: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Referer", referer)
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("sec-ch-ua-platform", `"Windows"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua", secCHUA)
	req.Header.Set("token", token)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("网络请求失败: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取响应失败: %v\n", err)
		os.Exit(1)
	}

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP 状态码异常 %d: %s\n", resp.StatusCode, string(respBody))
		os.Exit(1)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		fmt.Printf("JSON 解析失败: %v\n", err)
		os.Exit(1)
	}

	msg, ok := result["msg"].(string)
	if !ok {
		fmt.Printf("签到结果: %v\n", result)
		os.Exit(1)
	}

	switch msg {
	case "签到成功":
		fmt.Println("签到成功")
	case "已经签到过了":
		fmt.Println("今日已签到")
	default:
		fmt.Printf("签到失败: %s\n", msg)
		os.Exit(1)
	}
}
