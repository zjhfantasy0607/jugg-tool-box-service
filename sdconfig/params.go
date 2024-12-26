package sdconfig

import (
	"fmt"
	"jugg-tool-box-service/util"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

const jsonDir = "./sdconfig" // 指定读取配置的文件夹路径
const fileExt = ".json"      // 配置文件使用的后缀

var paramsMap map[string]string = map[string]string{}

// 初始化将所有 json 配置写入内存中
func init() {
	// 读取文件夹内容
	entries, err := os.ReadDir(jsonDir)
	if err != nil {
		util.LogErr(errors.WithStack(fmt.Errorf("读取文件失败: %w", err)), "./log/imageQueue.log")
		return
	}

	// 遍历文件夹中的文件
	for _, entry := range entries {
		// 只处理 .json 文件
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), fileExt) {

			// 获取文件名
			fileName := strings.TrimSuffix(entry.Name(), fileExt)

			// 构建文件路径
			filePath := filepath.Join(jsonDir, entry.Name())

			// 读取文件内容
			content, err := os.ReadFile(filePath)
			if err != nil {
				util.LogErr(errors.WithStack(fmt.Errorf("读取文件失败: %w", err)), "./log/imageQueue.log")
				continue
			}

			paramsMap[fileName] = string(content)
		}
	}
}

func Get(key string) string {
	if value, exists := paramsMap[key]; exists {
		return value
	}
	return ""
}
