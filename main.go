package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

// ================= 配置区 =================
const (
	FN_URL   = "http://192.168.100.44:5666" // 飞牛地址
	INTERVAL = 1                          // 监控间隔（分钟）
)

var (
	currentToken string
	lastID       int = 0
)

// 自动登录获取 Token
func getAutoToken() string {
	user, pwd := os.Getenv("FN_USER"), os.Getenv("FN_PWD")
	if user == "" || pwd == "" {
		fmt.Println("⚠️ 错误：未在环境变量中配置 FN_USER 或 FN_PWD")
		return ""
	}

	loginURL := fmt.Sprintf("%s/api/v1/auth/login", FN_URL)
	payload := map[string]string{"username": user, "password": pwd}
	jsonData, _ := json.Marshal(payload)

	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 登录飞牛失败: %v\n", err)
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Data struct{ Token string } `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Data.Token != "" {
		fmt.Println("✅ 自动登录成功，获取到新 Token")
		return "Bearer " + result.Data.Token
	}
	return ""
}

// 通用发送通知
func sendNotify(title, message string) {
	notifyURL := os.Getenv("NOTIFY_API_URL")
	if notifyURL == "" {
		return
	}

	payload := map[string]string{"title": title, "message": message}
	jsonData, _ := json.Marshal(payload)
	
	http.Post(notifyURL, "application/json", bytes.NewBuffer(jsonData))
	fmt.Printf("🚀 通知已发出: [%s] %s\n", title, message)
}

// 检查新片入库
func checkNewMedia() {
	if currentToken == "" {
		currentToken = getAutoToken()
		if currentToken == "" { return }
	}

	url := fmt.Sprintf("%s/api/v1/item/list", FN_URL)
	payload := []byte(`{"page_size":1,"sort_column":"create_time","sort_type":"DESC","tags":{"type":["Movie","TV"]}}`)
	
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	req.Header.Set("Authorization", currentToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		currentToken = "" 
		return
	}

	var result struct {
		Data struct {
			Items []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	json.NewDecoder(resp.Body).Decode(&result)

	if len(result.Data.Items) > 0 {
		latest := result.Data.Items[0]
		if lastID != 0 && latest.ID > lastID {
			sendNotify("🎬 新片入库", "发现新内容："+latest.Name)
		}
		lastID = latest.ID
	}
}

// 接收飞牛 Webhook
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return
	}

	eventType, _ := body["event"].(string)
	data, ok := body["data"].(map[string]interface{})
	if !ok { return }

	itemName, _ := data["item_name"].(string)
	userName, _ := data["user_name"].(string)

	if itemName != "" {
		title := "🔔 飞牛提醒"
		if eventType == "item.play" { title = "▶️ 正在播放" }
		if eventType == "item.stop" { title = "⏹️ 停止播放" }
		sendNotify(title, fmt.Sprintf("用户 [%s]: %s", userName, itemName))
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	fmt.Println("🚀 Go版监控助手已就绪...")

	go func() {
		for {
			checkNewMedia()
			time.Sleep(time.Duration(INTERVAL) * time.Minute)
		}
	}()

	// 5001 端口测试面板
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body style="text-align:center;padding-top:50px;">
				<h2>飞牛通知测试控制台</h2>
				<button style="padding:1
