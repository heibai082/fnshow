// 寻找并替换 triggerTestNotification 函数
func triggerTestNotification() {
	fmt.Println("🔔 收到 Web 端发起的【测试入库通知】请求！")

	// 1. 自动从系统环境变量读取你在飞牛配置的 URL
	// 对应你截图中的 NOTIFY_API_URL
	notifyURL := os.Getenv("NOTIFY_API_URL")
	if notifyURL == "" {
		fmt.Println("❌ 错误：未在环境变量中找到 NOTIFY_API_URL")
		return
	}

	// 2. 构造测试数据（模拟原版 Python 发送的格式）
	testData := map[string]interface{}{
		"title":   "【监控测试】",
		"message": "这是一条来自 Web 测试面板的入库通知测试，证明通道已打通！",
	}
	jsonData, _ := json.Marshal(testData)

	// 3. 发送 POST 请求
	resp, err := http.Post(notifyURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("❌ 发送失败: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Println("✅ 测试通知已成功推送到目标接口！")
	} else {
		fmt.Printf("⚠️ 接口返回异常状态码: %d\n", resp.StatusCode)
	}
}
