package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/robfig/cron/v3"
)

const (
	signInURL = "https://bbs.topfeel.com/api/gift/day_sign"
	referer   = "https://bbs.topfeel.com/h5/"
	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36"
	secCHUA   = `"Google Chrome";v="143", "Chromium";v="143", "Not A(Brand";v="24"`
)

func signIn(token string) (string, error) {
	now := time.Now().UnixMilli()
	newTime := now + int64(rand.Intn(4)+3)*1000

	body := map[string]any{
		"oldtime": now,
		"newtime": newTime,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("序列化请求体失败: %w", err)
	}

	req, err := http.NewRequest("POST", signInURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
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
		return "", fmt.Errorf("网络请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP 状态码异常 %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("JSON 解析失败: %w", err)
	}

	msg, ok := result["msg"].(string)
	if !ok {
		return "", fmt.Errorf("签到结果: %v", result)
	}

	switch msg {
	case "签到成功", "已经签到过了":
		return msg, nil
	default:
		return "", fmt.Errorf("签到失败: %s", msg)
	}
}

func doSignIn() {
	tokensStr := os.Getenv("TOPFEEL_TOKEN")
	if tokensStr == "" {
		fmt.Println("错误: 未设置 TOPFEEL_TOKEN 环境变量")
		return
	}

	tokens := strings.Split(tokensStr, ",")
	successCount := 0
	failCount := 0

	for i, token := range tokens {
		token = strings.TrimSpace(token)
		if token == "" {
			continue
		}

		accountID := fmt.Sprintf("账号%d", i+1)
		fmt.Printf("正在为 %s 签到...\n", accountID)

		msg, err := signIn(token)
		if err != nil {
			fmt.Printf("%s 签到失败: %v\n", accountID, err)
			failCount++
		} else {
			fmt.Printf("%s %s\n", accountID, msg)
			successCount++
		}
	}

	fmt.Printf("\n签到完成: 成功 %d 个, 失败 %d 个\n", successCount, failCount)
}

func main() {
	once := flag.Bool("once", false, "只执行一次签到后退出")
	flag.Parse()

	_ = godotenv.Load()

	if *once {
		doSignIn()
		return
	}

	fmt.Println("启动每日签到定时任务，每天 08:00 执行")

	c := cron.New(cron.WithLocation(func() *time.Location {
		loc, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			fmt.Printf("加载时区失败: %v，使用本地时间\n", err)
			return time.Local
		}
		return loc
	}()))

	c.AddFunc("0 8 * * *", doSignIn)

	c.Start()

	select {}
}
