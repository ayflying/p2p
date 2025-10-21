package main

import (
	"image"
	"os"

	"github.com/icza/icox" // 替换为icox库
	"github.com/nfnt/resize"
)

func main() {
	// 1. 读取PNG文件
	inputPath := "manifest/images/logo.png"     // 输入PNG路径
	outputPath := "manifest/images/favicon.ico" // 输出ICO路径

	pngFile, err := os.Open(inputPath)
	if err != nil {
		panic("无法打开PNG文件: " + err.Error())
	}
	defer pngFile.Close()

	// 2. 解码PNG为image.Image对象
	img, _, err := image.Decode(pngFile)
	if err != nil {
		panic("PNG解码失败: " + err.Error())
	}

	// 3. 定义ICO需要包含的尺寸（常见尺寸）
	sizes := []uint{16, 32, 64, 128} // 可根据需求添加更多尺寸
	var icoImages []image.Image

	// 4. 缩放图像到每个目标尺寸并收集
	for _, size := range sizes {
		// 使用Lanczos3算法缩放（高质量）
		resized := resize.Resize(size, size, img, resize.Lanczos3)
		icoImages = append(icoImages, resized)
	}

	// 5. 编码为ICO并写入文件（关键修改：使用icox.Encode）
	icoFile, err := os.Create(outputPath)
	if err != nil {
		panic("无法创建ICO文件: " + err.Error())
	}
	defer icoFile.Close()

	// icox.Encode直接支持多尺寸切片[]image.Image
	if err := icox.Encode(icoFile, icoImages); err != nil {
		panic("ICO编码失败: " + err.Error())
	}
}
