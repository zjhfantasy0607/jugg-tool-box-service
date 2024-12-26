package controller

import (
	"encoding/json"
	"fmt"
	"jugg-tool-box-service/common"
	"jugg-tool-box-service/middleware"
	"jugg-tool-box-service/model"
	"jugg-tool-box-service/util"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

func ClientProgressHandler(c *gin.Context) {
	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := common.CreateWS(c)
	if err != nil {
		util.LogErr(err, "./log/socket.log")
		return
	}
	defer func() { // 未正常关闭 socket 链接时对链接进行关闭
		conn.Close()
	}()

	// 连接开始读取超时设置为 3s 避免积累过多不通过用户验证的链接
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	// 开始连接时会收到客户端发来的第一条携带用户 token 的消息
	_, message, err := conn.ReadMessage()
	if err != nil {
		return
	}
	go func() {
		_, _, err := conn.ReadMessage()
		if err != nil {
			return
		}
	}()

	// 执行中间件做用户校验
	// 中间件错误时会劫持 socket 链接报错， 暂不处理
	token := "Bearer " + string(message)
	c.Request.Header.Set("Authorization", token) // 设置 header token
	if middleware.AuthMiddleware()(c); c.IsAborted() {
		return
	}

	// 中间件验证过后重新设置回原来的超时设置
	conn.SetReadDeadline(time.Now().Add(common.ReadTimeout))

	userData, exists := c.Get("user")
	if !exists || userData == nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	// 将 user 断言为正确的类型
	user, ok := userData.(model.User)
	if !ok {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}
	userID := user.UID

	// 创建一个每 1 秒触发一次的 ticker
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		// 检测是否结束
		isEnd := true

		// 获取用户的任务 ID 列表
		taskIds, err := common.Rdb.SMembers(common.Ctx, userID).Result()
		if err != nil {
			util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
			return
		}

		tasks := gin.H{}
		for _, taskId := range taskIds {
			task, err := common.Rdb.HGetAll(common.Ctx, taskId).Result()
			if err != nil {
				util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
				break
			}

			// 将字符串转为 float64，保留小数点后两位
			progress, err := strconv.ParseFloat(task["progress"], 64)
			if err != nil {
				util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
				break
			}
			progress = math.Round(progress*100) / 100

			// 获取当前 taskId 在 tasks 中的排位
			rank, err := common.Rdb.ZRank(common.Ctx, "tasks", taskId).Result()
			if err != nil {
				util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
				break
			}

			if task["status"] == "pending" || task["status"] == "producing" {
				isEnd = false
			} else if task["status"] == "success" || task["status"] == "falied" {
				// 当任务完成并被读取过时，可以直接从人物队列中删除任务
				delTaskToRedis(taskId)
			}

			tasks[taskId] = gin.H{
				"status":   task["status"],
				"progress": progress,
				"rank":     int(rank + 1),
				"endtime":  task["endtime"],
				"output":   task["output"],
				"params":   task["params"],
			}
		}

		// 获取全局任务有序集合中的任务总数
		totalTasks, err := common.Rdb.ZCard(common.Ctx, "tasks").Result()
		if err != nil {
			util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
			break
		}

		progressMessage := gin.H{"tasks": tasks, "isEnd": isEnd, "totalRank": totalTasks}

		if err := conn.WriteJSON(progressMessage); err != nil {
			break
		}

		// 结束，终止 websocket 的通讯
		if isEnd {
			break
		}
	}

	// 发送关闭消息关闭链接
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
}

func SDCallbackHandler(c *gin.Context) {
	var connTaskId string

	// 将 HTTP 连接升级为 WebSocket 连接
	conn, err := common.CreateWS(c)
	if err != nil {
		util.LogErr(err, "./log/socket.log")
		return
	}
	defer func() {
		// 处理因为异常导致的 websocket 链接关闭
		if conn.Close() == nil {
			// 将最后通讯的任务设置为失败，进行用户的积分返还和任务状态设置
			err := failHandler(connTaskId)
			if err != nil {
				util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
			}
		}
	}()

	// 定义消息超时间，超过这个时间未收到除 Ping/Pong 外的消息就关闭连接
	messageTimeout := 1 * time.Minute
	lastMessageTime := time.Now()

	// 覆盖默认的 pong 处理
	conn.SetPongHandler(func(appData string) error {
		// 当前业务检查是否超时
		if time.Since(lastMessageTime) > messageTimeout {
			fmt.Println("超过一定时间未收到正常消息，关闭连接")
			// 发送关闭消息
			conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			return nil
		}

		fmt.Println("pong")

		conn.SetReadDeadline(time.Now().Add(common.ReadTimeout))
		return nil
	})

	// 设置 正常关闭 事件处理程序
	conn.SetCloseHandler(func(code int, text string) error {
		conn.Close()
		return nil
	})

	// 启动一个无限循环来处理 WebSocket 消息
	for {
		// 更新业务超时时间
		lastMessageTime = time.Now()

		// 读取客户端发来的消息
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 打印接收到的消息
		jsonParsed := gjson.Parse(string(message))
		jsonStr := jsonParsed.String()
		api := gjson.Get(jsonStr, "api").String()
		taskId := gjson.Get(jsonStr, "task_id").String()
		status := gjson.Get(jsonStr, "status").Int()
		body := gjson.Get(jsonStr, "body").String()

		// 解密出正确的 taskId
		taskId, err = util.Decrypt(taskId, viper.GetString("appkey"))
		if err != nil {
			err = fmt.Errorf("taskId decrypt error: %w", err)
			util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
			continue
		}

		if api == "sdapi/v1/progress" {
			connTaskId = taskId
		} else {
			connTaskId = ""
		}

		if status != http.StatusOK {
			err := failHandler(taskId) // 任务失败，进行用户的积分返还和任务状态设置
			if err != nil {
				util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
				continue
			}

			err = fmt.Errorf("sd server status: %d, error: %s", status, body)
			util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
			continue
		}

		// 调用 api 所对应的handler
		callSuccessHandler(api, taskId, body)
	}
}

func callSuccessHandler(api string, taskId string, body string) {
	switch api {
	case "sdapi/v1/progress":
		progressHandler(taskId, body)
	case "sdapi/v1/txt2img", "sdapi/v1/img2img":
		successHandler(taskId, body)
	case "sdapi/v1/extra-single-image", "rembg":
		extrasHandler(taskId, body)
	}
}

func progressHandler(taskId string, progress string) {
	// 检查哈希键是否存在
	exists, err := common.Rdb.HExists(common.Ctx, taskId, "status").Result()
	if err != nil {
		err = fmt.Errorf("HExists error: %w", err)
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	if !exists {
		return
	}

	// 获取当前的状态
	status, err := common.Rdb.HGet(common.Ctx, taskId, "status").Result()
	if err != nil {
		err = fmt.Errorf("HGet error: %w", err)
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	// 如果状态是 success 或 failed，直接返回
	if status == "success" || status == "failed" {
		return
	}

	err = common.Rdb.HSet(common.Ctx, taskId, "status", "producing").Err()
	if err != nil {
		err = fmt.Errorf("progress status HSet error: %w", err)
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	err = common.Rdb.HSet(common.Ctx, taskId, "progress", progress).Err()
	if err != nil {
		err = fmt.Errorf("progress progress HSet error: %w", err)
		util.LogErr(err, "./log/imageQueue.log")
		return
	}
}

func successHandler(taskId string, response string) {
	// 使用 gjson 提取 "" 结果数组
	imagesJson := gjson.Get(response, "images")

	// 保存图片到本地服务器
	outputPath := viper.GetString("image.output")
	var output []string
	if imagesJson.IsArray() {
		imagesJson.ForEach(func(key, value gjson.Result) bool {
			savePath, err := saveOutputImage(value.String(), outputPath)
			if err != nil { // 图片保存错误
				err = fmt.Errorf("image output error: failed to save image: %w", err)
				util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
			}
			output = append(output, savePath)
			return true // 继续下一个元素
		})
	}

	// 将保存的图片地址编码为json
	outputJson, err := json.Marshal(output)
	if err != nil {
		err = errors.WithStack(fmt.Errorf("image output error: failed to marshal json: %w", err))
		util.LogErr(err, "./log/imageQueue.log")
		return
	}

	// 重新使用 gjson 从已经反序列的 JSON 中读取seed值
	infoJson := gjson.Parse(gjson.Get(response, "info").String())
	seed := gjson.Get(infoJson.String(), "seed")

	// 查询对应的任务记录
	db := common.GetDB()
	task := &model.Task{}
	result := db.Where("task_id = ?", taskId).First(task)
	if result.Error != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	// 更新数据库值
	task.Status = "success"
	task.EndTime = time.Now()
	task.Output = string(outputJson)
	task.UsedTime = task.EndTime.Sub(task.StartTime).Milliseconds() // 计算生成消耗的时间
	task.Params, err = sjson.Set(task.Params, "seed", seed.Int())

	if result.Error != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	// 更新任务数据库记录
	db.Save(task)

	err = successChangeRedis(task)
	if err != nil {
		util.LogErr(err, "./log/imageQueue.log")
	}
}

func failHandler(taskId string) error {
	var err error

	// 进行数据库的更新
	db := common.GetDB()
	task := &model.Task{}

	// - 查询对应的任务记录
	result := db.Where("task_id = ?", taskId).First(task)
	if result.Error != nil {
		return errors.WithStack(result.Error)
	}

	// - 查询对应的用户信息
	user := &model.User{}
	result = db.Where("uid = ?", task.UID).First(user)
	if result.Error != nil {
		return errors.WithStack(result.Error)
	}

	// - 更新任务
	tx := db.Begin() // 开始事务
	task.Status = "failed"
	task.EndTime = time.Now()
	task.UsedTime = task.EndTime.Sub(task.StartTime).Milliseconds() // 计算生成消耗的时间
	tx.Save(task)

	// - 失败返还损失的积分
	pointRecord := model.Point{
		TaskId:       task.TaskId,
		UID:          task.UID,
		BeforePoints: user.Points,
		Points:       task.UsedPoints,
		Tool:         task.Tool,
		Remark:       "任务失败积分返还",
	}
	err = pointRecord.T_AddRecord(tx, user)

	if err != nil {
		return errors.WithStack(err)
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		return errors.WithStack(err)
	}

	// 修改redis记录中的状态
	err = common.Rdb.HSet(common.Ctx, taskId, "status", "failed").Err()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.Rdb.HSet(common.Ctx, taskId, "endtime", task.EndTime.Format("2006-01-02 15:04:05")).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	// 5秒后从redis的队列中删除当前任务映射
	time.AfterFunc(5*time.Second, func() {
		err = delTaskToRedis(taskId)
		if err != nil {
			util.LogErr(err, "./log/imageQueue.log")
		}
	})

	return nil
}

func extrasHandler(taskId string, response string) {
	imageJson := gjson.Get(response, "image")

	// 保存图片到本地服务器
	outputPath := viper.GetString("image.output")
	output, err := saveOutputImage(imageJson.String(), outputPath)
	if err != nil { // 图片保存错误
		err = fmt.Errorf("image output error: failed to save image: %w", err)
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
	}

	// 将保存的图片地址编码为json
	outputJson, err := json.Marshal([1]string{output})
	if err != nil {
		err = errors.WithStack(fmt.Errorf("image output error: failed to marshal json: %w", err))
		util.LogErr(err, "./log/imageQueue.log")
		return
	}

	// 查询对应的任务记录
	db := common.GetDB()
	task := &model.Task{}
	result := db.Where("task_id = ?", taskId).First(task)
	if result.Error != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	// 更新数据库值
	task.Status = "success"
	task.EndTime = time.Now()
	task.Output = string(outputJson)
	task.UsedTime = task.EndTime.Sub(task.StartTime).Milliseconds() // 计算生成消耗的时间

	if result.Error != nil {
		util.LogErr(errors.WithStack(err), "./log/imageQueue.log")
		return
	}

	// 更新任务数据库记录
	db.Save(task)

	successChangeRedis(task)
}

func successChangeRedis(task *model.Task) error {
	var err error

	// 修改redis记录中的状态
	err = common.Rdb.HSet(common.Ctx, task.TaskId, "status", "success").Err()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.Rdb.HSet(common.Ctx, task.TaskId, "progress", 1).Err()
	if err != nil {
		return fmt.Errorf("progress progress HSet error: %w", err)
	}
	err = common.Rdb.HSet(common.Ctx, task.TaskId, "endtime", task.EndTime.Format("2006-01-02 15:04:05")).Err()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.Rdb.HSet(common.Ctx, task.TaskId, "output", task.Output).Err()
	if err != nil {
		return errors.WithStack(err)
	}
	err = common.Rdb.HSet(common.Ctx, task.TaskId, "params", task.Params).Err()
	if err != nil {
		return errors.WithStack(err)
	}

	// 5秒后从redis的队列中删除当前任务映射
	time.AfterFunc(5*time.Second, func() {
		err = delTaskToRedis(task.TaskId)
		if err != nil {
			util.LogErr(err, "./log/imageQueue.log")
		}
	})

	return nil
}
