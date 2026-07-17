package gui

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

// GenerateTrayIcon 生成 study 托盘图标（16x16 PNG）
// 设计：圆角方形，深蓝底 + 青色书本折角
func GenerateTrayIcon() []byte {
	const S = 16
	img := image.NewRGBA(image.Rect(0, 0, S, S))

	bg := color.RGBA{R: 30, G: 30, B: 46, A: 255}
	accent := color.RGBA{R: 137, G: 180, B: 250, A: 255}
	light := color.RGBA{R: 186, G: 210, B: 255, A: 255}

	// 填充圆角背景
	fillRounded(img, bg, S, S, 3)

	// 绘制书本：左页 + 右页
	leftPage := image.Rect(2, 3, 8, 14)
	rightPage := image.Rect(8, 3, 14, 14)
	draw.Draw(img, leftPage, image.NewUniform(accent), image.Point{}, draw.Over)
	draw.Draw(img, rightPage, image.NewUniform(light), image.Point{}, draw.Over)

	// 书脊线
	for y := 3; y < 14; y++ {
		img.Set(8, y, bg)
	}

	// 折角效果（右下角小三角）
	img.Set(12, 12, bg)
	img.Set(12, 13, bg)
	img.Set(13, 13, bg)

	// 顶部文字行（模拟书页文字）
	textColor := color.RGBA{R: 30, G: 30, B: 46, A: 180}
	for y := 5; y <= 5; y++ {
		for x := 3; x <= 7; x++ {
			img.Set(x, y, textColor)
		}
	}
	for y := 7; y <= 7; y++ {
		for x := 9; x <= 13; x++ {
			img.Set(x, y, textColor)
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// GenerateAppIcon 生成应用窗口图标（32x32 PNG）
func GenerateAppIcon() []byte {
	const S = 32
	img := image.NewRGBA(image.Rect(0, 0, S, S))

	bg := color.RGBA{R: 30, G: 30, B: 46, A: 255}
	accent := color.RGBA{R: 137, G: 180, B: 250, A: 255}
	light := color.RGBA{R: 186, G: 210, B: 255, A: 255}

	fillRounded(img, bg, S, S, 6)

	// 书本：左页 + 右页（16x16 的2倍比例）
	leftPage := image.Rect(5, 6, 16, 27)
	rightPage := image.Rect(16, 6, 27, 27)
	draw.Draw(img, leftPage, image.NewUniform(accent), image.Point{}, draw.Over)
	draw.Draw(img, rightPage, image.NewUniform(light), image.Point{}, draw.Over)

	// 书脊
	for y := 6; y < 27; y++ {
		img.Set(16, y, bg)
	}

	// 折角
	for x := 24; x < 27; x++ {
		for y := 24; y < 27; y++ {
			if x+y >= 49 {
				img.Set(x, y, bg)
			}
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

// fillRounded 在图像中填充圆角矩形
func fillRounded(img *image.RGBA, c color.RGBA, w, h, r int) {
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			// 检查是否在圆角内
			px, py := x, y
			if x < r && y < r {
				if inCircle(px, py, r-1, r-1, r) {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
			} else if x >= w-r && y < r {
				if inCircle(px, py, w-r, r-1, r) {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
			} else if x < r && y >= h-r {
				if inCircle(px, py, r-1, h-r, r) {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
			} else if x >= w-r && y >= h-r {
				if inCircle(px, py, w-r, h-r, r) {
					img.Set(x, y, color.RGBA{0, 0, 0, 0})
					continue
				}
			}
			img.Set(x, y, c)
		}
	}
}

func inCircle(x, y, cx, cy, r int) bool {
	dx := float64(x - cx)
	dy := float64(y - cy)
	return math.Sqrt(dx*dx+dy*dy) > float64(r)
}
