package common

import (
	"jugg-tool-box-service/model"
)

func autoMigrate() {
	// 建表
	db.AutoMigrate(
		&model.User{},
		&model.Captcha{},
		&model.EmailRecord{},
		&model.Tool{},
		&model.Task{},
		&model.Point{},
		&model.Seo{},
	)

	ToolInit()
	SeoInit()
}

func ToolInit() {
	// tools表 插入默认数据
	datas := model.ToolSlice{
		{
			ID:    1,
			Pid:   0,
			Title: "AI 图片工具",
			Icon: `<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" strokeWidth={1.5} stroke="currentColor" class="size-5 text-green-500">
					<path strokeLinecap="round" strokeLinejoin="round" d="m2.25 15.75 5.159-5.159a2.25 2.25 0 0 1 3.182 0l5.159 5.159m-1.5-1.5 1.409-1.409a2.25 2.25 0 0 1 3.182 0l2.909 2.909m-18 3.75h16.5a1.5 1.5 0 0 0 1.5-1.5V6a1.5 1.5 0 0 0-1.5-1.5H3.75A1.5 1.5 0 0 0 2.25 6v12a1.5 1.5 0 0 0 1.5 1.5Zm10.5-11.25h.008v.008h-.008V8.25Zm.375 0a.375.375 0 1 1-.75 0 .375.375 0 0 1 .75 0Z" />
				</svg>`,
			Description: "",
			Url:         "",
		},
		{ID: 2, Pid: 1, Title: "图片高清修复", Description: "通过智能算法提升图片细节与清晰度", Icon: "text-[#eedb95]", Url: "/img-resize", Tool: "resize"},
		{ID: 3, Pid: 1, Title: "自动去除背景", Description: "通过深度学习算法，智能识别图像中的前景与背景，精确去除背景。", Icon: "text-[#de7660]", Url: "/rembg", Tool: "rembg"},
		{ID: 4, Pid: 1, Title: "文生图", Description: "根据输入的文本描述生成对应的图像。", Icon: "text-[#7891b1]", Url: "/txt2img", Tool: "txt2img"},
		{ID: 5, Pid: 1, Title: "图生图", Description: "通过输入一张图像，生成与之相关或具有某种变化的另一张图像。", Icon: "text-[#f5bfc8]", Url: "/img2img", Tool: "img2img"},
	}

	for _, data := range datas {
		db.FirstOrCreate(&data, model.Tool{Title: data.Title})
	}
}

func SeoInit() {
	// tools表 插入默认数据
	datas := []model.Seo{
		{ID: 1, Url: "/home", Title: "JUGG工具箱", Keywords: "JUGG工具箱, AI在线工具, AI图片生成工具, AI在线工具", Description: "JUGG工具箱集合了多种实用的AI在线工具，帮助您轻松解决各种问题。"},
		{ID: 2, Url: "/img-resize", Title: "图片高清修复", Keywords: "图片修复, 高清修复, 图片清晰度, 图像增强", Description: "通过智能算法提升图片细节与清晰度，让您的图片焕然一新，适用于各种场景。"},
		{ID: 3, Url: "/rembg", Title: "自动去除背景", Keywords: "去除背景, 图片背景, 背景移除, 深度学习", Description: "通过深度学习算法，智能识别图像中的前景与背景，精确去除背景，轻松制作透明背景图。"},
		{ID: 4, Url: "/txt2img", Title: "文生图", Keywords: "文本生成图像, AI生成图像, 文生图, 文本到图像", Description: "根据输入的文本描述生成对应的图像，帮助您将创意迅速转化为视觉艺术。"},
		{ID: 5, Url: "/img2img", Title: "图生图", Keywords: "图像生成, 图像变化, AI图像生成, 图片重绘", Description: "通过输入一张图像，生成与之相关或具有某种变化的另一张图像，拓展您的创作空间。"},
		{ID: 6, Url: "/tasks", Title: "任务记录", Keywords: "", Description: ""},
		{ID: 7, Url: "/points", Title: "积分记录", Keywords: "", Description: ""},
	}

	for _, data := range datas {
		db.FirstOrCreate(&data, model.Tool{Url: data.Url})
	}
}
