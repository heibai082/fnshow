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
// ==========================================

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
		fmt.Println("✅ 自动登录成功，Token 已更新")
		return "Bearer " + result.Data.Token
	}
	return ""
}

// 通用发送通知函数
func sendNotify(title, message string) {
	notifyURL := os.Getenv("NOTIFY_API_URL")
	if notifyURL == "" {
		fmt.Println("⚠️ 未配置 NOTIFY_API_URL，无法发送通知")
		return
	}

	payload := map[string]string{"title": title, "message": message}
	jsonData, _ := json.Marshal(payload)
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(notifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 通知发送失败: %v\n", err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("🚀 通知已送达: [%s] %s\n", title, message)
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
	if err != nil || resp.StatusCode == 401 {
		currentToken = "" // 失效则清空以便下次重连
		return
	}
	defer resp.Body.Close()

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

// 接收飞牛 Webhook（处理播放、停止等）
func handleWebhook(w http.ResponseWriter, r *http.Request) {
	var body map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return
	}

	// 提取飞牛发过来的事件内容（根据飞牛标准 webhook 格式）
	eventType, _ := body["event"].(string)
	data, _ := body["data"].(map[string]interface{})
	itemName, _ := data["item_name"].(string)
	userName, _ := data["user_name"].(string)

	var title, msg string
	switch eventType {
	case "item.play":
		title = "▶️ 正在播放"
		msg = fmt.Sprintf("用户 [%s] 正在观看: %s", userName, itemName)
	case "item.stop":
		title = "⏹️ 停止播放"
		msg = fmt.Sprintf("用户 [%s] 停止观看: %s", userName, itemName)
	default:
		title = "🔔 飞牛提醒"
		msg = fmt.Sprintf("事件: %s, 内容: %s", eventType, itemName)
	}

	if itemName != "" {
		sendNotify(title, msg)
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	fmt.Println("🚀 飞牛全能监控助手已启动...")

	// 1. 周期性检查新片
	go func() {
		for {
			checkNewMedia()
			time.Sleep(time.Duration(INTERVAL) * time.Minute)
		}
	}()

	// 2. 监听 5000 端口（接收飞牛原有的播放/停止 Webhook）
	http.HandleFunc("/webhook", handleWebhook) // 建议飞牛 Webhook 地址填这个
	http.HandleFunc("/", handleWebhook)        // 兼容模式：直接填 IP:端口 也能收到

	// 3. 监听 5001 端口（提供 Web 测试按钮）
	go func() {
		testMux := http.NewServeMux()
		testMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, `<html><body style="text-align:center;padding-top:50px;">
				<h2>飞牛通知测试控制台</h2>
				<button style="padding:15px 30px;" onclick="fetch('/test').then(()=>alert('发送请求成功，请检查手机'))">点击发送测试通知</button>
			</body></html>`)
		})
		testMux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
			sendNotify("🧪 手动测试", "这是一条来自测试按钮的消息")
			w.Write([]byte("ok"))
		})
		http.ListenAndServe(":5001", testMux)
	}()

	// 启动主服务（5000 端口）
	http.ListenAndServe(":5000", nil)
}
