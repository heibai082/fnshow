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
	// 1. 你的飞牛地址
	FN_URL = "http://192.168.100.44:5666"

	// 2. 你的身份令牌（Bearer 后面记得加空格）
	TOKEN = "Bearer 你的TOKEN粘贴在这里"

	// 3. 检查频率（1分钟检查一次）
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
			fmt.Printf("【监控启动】当前最新内容: %s\n", latest.Name)
			return
		}
		if latest.ID > lastID {
			fmt.Printf("🚀 监测到新入库：%s\n", latest.Name)
			lastID = latest.ID
		}
	}
}

func main() {
	fmt.Println("飞牛全库入库监控已启动...")
	for {
		checkNewMedia()
		time.Sleep(time.Duration(INTERVAL) * time.Minute)
	}
}
