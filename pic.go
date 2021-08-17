package file

import (
	"bytes"
	"fmt"
	"image"
	"math/rand"
	"path"
	"time"

	"math"

	"image/color"
	"image/png"
	"io/ioutil"
	"os"

	"github.com/golang/freetype"
	"github.com/xuri/excelize/v2"
)

func FileWithWatermarkExcel(body [][]interface{}, text string, fontaddr string) (*bytes.Buffer, error) {
	if len(body) < 1 {
		return nil, fmt.Errorf("empty body： %v", body)
	}

	const sheetName = "Sheet1"

	// write body
	f := excelize.NewFile()
	for row, line := range body {
		if err := f.SetSheetRow(sheetName, fmt.Sprintf("A%d", row+1), &line); err != nil {
			return nil, fmt.Errorf("set line err: %v", err)
		}
	}
	img, err := BuildTextPNG(text, fontaddr)
	if err != nil {
		return nil, fmt.Errorf("build text png err: %v", err)
	}
	if err := f.SetSheetBackground(sheetName, img); err != nil {
		return nil, fmt.Errorf("set sheet background err: %v", err)
	}

	return f.WriteToBuffer()
}

const (
	dx       = 400 // 图片的大小 宽度
	dy       = 400 // 图片的大小 高度
	fontSize = 30  // 字体尺寸
	fontDPI  = 120 // 屏幕每英寸的分辨率
)

func BuildTextPNG(text string, fontlink string) (string, error) {

	// 需要保存的文件
	imgfile, err := randomTmpFile()
	if err != nil {
		return "", err
	}
	defer imgfile.Close()

	// 新建一个 指定大小的 RGBA位图
	img := image.NewNRGBA(image.Rect(0, 0, dx, dy))

	// 读字体数据
	fontBytes, err := ioutil.ReadFile(fontlink)
	if err != nil {
		return "", err
	}
	font, err := freetype.ParseFont(fontBytes)
	if err != nil {

		return "", err
	}

	c := freetype.NewContext()
	c.SetDPI(fontDPI)
	c.SetFont(font)
	c.SetFontSize(fontSize)
	c.SetClip(img.Bounds())
	c.SetDst(img)
	c.SetSrc(image.NewUniform(color.CMYK{10, 4, 0, 38}))
	c.SetSrc(image.Black)

	centerx, centery := (dx-(len(text))*fontSize)/2, (dy-fontSize)/2+fontSize
	pt := freetype.Pt(centerx, centery) // 字出现的位置

	_, err = c.DrawString(text, pt)
	if err != nil {
		return "", err
	}

	// // 0,0 ----------
	// //     |
	// //     | D33097
	// //     |
	rotate := image.NewRGBA(image.Rect(-dx, -dy, dx*2, dy*2))
	for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
	    for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
	        //  设置像素点
			// 与x的夹角
			if img.At(x, y) == image.Transparent {
				continue
			}
			pointx := x - centerx
			pointy := y - centery
			

			radius := math.Sqrt(float64(pointx*pointx+pointy*pointy))
			angle := math.Acos(float64(pointy)/radius) - 45

			newx := math.Sin(angle) * radius + float64(pointx)
			newy := math.Cos(angle) * radius + float64(pointy)

	        rotate.Set(int(newx), int(newy), img.At(x, y))
	    }
	}

	// 以PNG格式保存文件
	err = png.Encode(imgfile, rotate)
	return imgfile.Name(), err
}

func randomTmpFile() (*os.File, error) {
	randnum := rand.Int()
	var file string
	for i := 0; i < 3; i++ {
		file = path.Join(os.TempDir(), fmt.Sprintf("%d%d%d.png", randnum, os.Getpid(), time.Now().Nanosecond()))
		if FileExist(file) {
			continue
		}
	}
	imgfile, err := os.Create(file)
	if err != nil {
		return nil, err
	}
	return imgfile, nil
}

func FileExist(path string) bool {
	_, err := os.Lstat(path)
	return !os.IsNotExist(err)
}
