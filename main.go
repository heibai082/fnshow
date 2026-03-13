package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ================= 配置区 =================
const (
	FN_URL = "http://192.168.100.44:5666"
	TOKEN  = "Bearer 1ae45321e2ae42ed8a5723f39d62cc4"
	INTERVAL = 1
)
// ==========================================

var lastID int = 0

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
		}
	}
}

// 供 Web 按钮调用的测试功能
func triggerTestNotification() {
	fmt.Println("🔔 收到 Web 端发起的【测试入库通知】请求！")
	// 未来你的微信/PushDeer推送代码可以写在这里
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
			<button onclick="sendTest()">发送测试通知</button>
			<div id="msg"></div>
			<script>
				function sendTest() {
					fetch('/test').then(res => res.text()).then(text => {
						document.getElementById('msg').innerText = text + " (" + new Date().toLocaleTimeString() + ")";
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
		w.Write([]byte("✅ 测试请求已发送，请查看 Docker 日志！"))
	})

	// 监听 5001 端口
	http.ListenAndServe(":5001", nil)
}
