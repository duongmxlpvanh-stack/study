// +build ignore

// 临时工具：将多张 PNG 合并为多分辨率 .ico 文件
package main

import (
	"encoding/binary"
	"os"
)

func main() {
	sizes := []struct {
		path   string
		width  uint8
		height uint8
	}{
		{"icons8-保存图书-100.png", 100, 100},
		{"icons8-保存图书-96.png", 96, 96},
		{"icons8-保存图书-48.png", 48, 48},
	}

	var pngData [][]byte
	for _, s := range sizes {
		data, err := os.ReadFile(s.path)
		if err != nil {
			panic(err)
		}
		pngData = append(pngData, data)
	}

	// 计算 offset：header(6) + directory(16*N)
	offset := 6 + 16*len(pngData)

	out, err := os.Create("icon.ico")
	if err != nil {
		panic(err)
	}
	defer out.Close()

	// ICO header
	writeLE(out, uint16(0))   // reserved
	writeLE(out, uint16(1))   // type: ICO
	writeLE(out, uint16(len(pngData))) // count

	// Directory entries
	for i, data := range pngData {
		out.Write([]byte{sizes[i].width})  // width
		out.Write([]byte{sizes[i].height}) // height
		out.Write([]byte{0})       // colors
		out.Write([]byte{0})       // reserved
		writeLE(out, uint16(1))    // planes
		writeLE(out, uint16(32))   // bpp
		writeLE(out, uint32(len(data))) // size
		writeLE(out, uint32(offset))     // offset
		offset += len(data)
	}

	// Image data (PNG directly)
	for _, data := range pngData {
		out.Write(data)
	}
}

func writeLE(f *os.File, v any) {
	binary.Write(f, binary.LittleEndian, v)
}
