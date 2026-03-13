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
	FN_URL   = "http://192.168.100.44:5666"
	TOKEN    = "Bearer 1ae45321e2ae42ed8a5723f39d62cc4"
	INTERVAL = 1
)

// ==========================================

var lastID int = 0

// 核心监控逻辑：轮询飞牛 API
func checkNewMedia() {
	url := fmt.Sprintf("%s/api/v1/item/list", FN_URL)
	payload := map[string]interface{}{
		"page_size":   1,
		"sort_column": "create_time",
		"sort_type":   "DESC",
		"tags": map[string]interface{}{
			"type": []string{"Movie", "TV"},
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", TOKEN)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result struct {
		Data struct {
			Items []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			} `json:"items"`
		} `json:"data"`
	}
	json.Unmarshal(body, &result)

	if len(result.Data.Items) > 0 {
		latest := result.Data.Items[0]

		if lastID == 0 {
			lastID = latest.ID
			fmt.Printf("【监控已启动】当前最新内容: %s\n", latest.Name)
			return
		}

		if latest.ID > lastID {
			fmt.Printf("🚀 监测到新片入库：%s\n", latest.Name)
			lastID = latest.ID
			// 自动触发通知（可选：如果你想入库时也自动发通知，可以调用 triggerTestNotification）
		}
	}
}

// 供 Web 按钮调用的测试功能：向你配置的 NOTIFY_API_URL 发送通知
func triggerTestNotification() {
	fmt.Println("🔔 收到 Web 端发起的【测试入库通知】请求！")

	// 自动从系统环境变量读取你在飞牛配置的 URL
	notifyURL := os.Getenv("NOTIFY_API_URL")
	if notifyURL == "" {
		fmt.Println("❌ 错误：未在容器环境变量中找到 NOTIFY_API_URL")
		return
	}

	// 构造测试数据
	testData := map[string]interface{}{
		"title":   "【监控测试】",
		"message": "这是一条来自 Web 测试面板的入库通知测试，证明通道已打通！",
	}
	jsonData, _ := json.Marshal(testData)

	// 发送 POST 请求到你的通知接口
	resp, err := http.Post(notifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 发送失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == 201 {
		fmt.Println("✅ 测试通知已成功推送到目标接口！")
	} else {
		fmt.Printf("⚠️ 接口返回异常状态码: %d\n", resp.StatusCode)
	}
}

func main() {
	fmt.Println("--------------------------------")
	fmt.Println("  飞牛 fnshow 增强版监控 运行中  ")
	fmt.Println("  Web 测试面板已在 5001 端口就绪 ")
	fmt.Println("--------------------------------")

	// 1. 开启后台静默监控
	go func() {
		for {
			checkNewMedia()
			time.Sleep(time.Duration(INTERVAL) * time.Minute)
		}
	}()

	// 2. 提供 Web 界面
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>入库通知测试</title>
			<style>
				body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; background-color: #f4f4f9; }
				button { padding: 15px 30px; font-size: 18px; color: white; background-color: #007bff; border: none; border-radius: 5px; cursor: pointer; }
				button:hover { background-color: #0056b3; }
				#msg { margin-top: 20px; font-size: 16px; color: green; font-weight: bold; }
			</style>
		</head>
		<body>
			<h2>飞牛入库通知 - 独立测试面板</h2>
			<p>点击下方按钮，测试是否能收到通知消息</p>
			<button onclick="sendTest()">发送测试通知</button>
			<div id="msg"></div>
			<script>
				function sendTest() {
					const msgDiv = document.getElementById('msg');
					msgDiv.innerText = "正在发送...";
					fetch('/test').then(res => res.text()).then(text => {
						msgDiv.innerText = text + " (" + new Date().toLocaleTimeString() + ")";
					}).catch(err => {
						msgDiv.innerText = "请求失败，请检查 Docker 日志";
					});
				}
			</script>
		</body>
		</html>`
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})

	http.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		triggerTestNotification()
		w.Write([]byte("✅ 测试请求已处理，请查看手机或 Docker 日志！"))
	})

	// 监听 5001 端口
	err := http.ListenAndServe(":5001", nil)
	if err != nil {
		fmt.Printf("❌ Web 服务器启动失败: %v\n", err)
	}
}
