package main

import (
	"bytes"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
)
import "flag"
import "path/filepath"
import "github.com/signintech/gopdf"

var (
	src = flag.String("src", "./", "the images directory path (png, jpg & jpeg) files")
	as  = flag.String("as", "./result.pdf", "the result filename")
	test  = flag.Bool("t", false, "for test")
	abort = flag.Bool("abort", true, "abort if error occurs")
	//verticalParts = flag.Int("vertical-parts", 1, "split vertical to x parts")
	//horizontalParts = flag.Int("horizontal-parts", 1, "split horizontal to x parts")
)

func init() {
	flag.Parse()
}

var width float64 = 210
var height float64 = 315
var A4 = gopdf.Rect{W: width, H: height }

func main() {
	font := "times"
	fontPath := "times.ttf"
	box := packr.New("My Resources", "./res")
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{Unit: gopdf.UnitMM, PageSize: A4})

	timesTTF, err := box.Find(fontPath)
	if err != nil {
		fmt.Printf("could not read times.ttf > %s", err.Error())
		os.Exit(1)
	}

	if err = pdf.AddTTFFontByReader(font, bytes.NewReader(timesTTF)); err != nil {
		fmt.Printf("Add font[%s] err > %s\n", font, err.Error())
		os.Exit(1)
	}

	err = pdf.SetFont(font, "", 14)
	if err != nil {
		fmt.Printf("Set font err > %s\n", err.Error())
		os.Exit(2)
	}

	fmt.Println("Reading files from (", *src, ") and saving the result as (", *as,")")
	fmt.Println("-----------------------")


	jpgs, _ := filepath.Glob(*src + "/*.jpg")
	pngs, _ := filepath.Glob(*src + "/*.png")
	jpegs, _ := filepath.Glob(*src + "/*.jpeg")
	files := append(jpgs, append(jpegs, pngs...)...)

	tempImgBuf := &bytes.Buffer{}
	for i := 0; i < len(files); i++ {
		fmt.Println(i+1, ")- adding ", files[i])
		if *test {
			continue
		}
		x := float64(0)
		if x < 0 {
			continue
		}

		fileInfo, err := os.Stat(files[i])
		if err != nil {
			fmt.Printf("Could not open file %s > %s\n", files[i], err.Error())
			if *abort {
				os.Exit(3)
			}
		}

		file, err := os.Open(files[i])
		if err != nil {
			fmt.Printf("Could not open file %s > %s\n", files[i], err.Error())
			if *abort {
				os.Exit(3)
			}
		}

		fileInBytes := make([]byte, fileInfo.Size())

		n, err := file.Read(fileInBytes)
		if err != nil || int64(n) != fileInfo.Size() {
			fmt.Printf("Could not read file %s\n", files[i])
			if *abort {
				os.Exit(3)
			}
		}
		file.Close()


		imgCfg, err := getImgConfig(bytes.NewReader(fileInBytes))
		if err != nil {
			fmt.Printf("get img config[%s] > %s\n", files[i], err.Error())
			if *abort {
				os.Exit(3)
			}
		}

		/////////////////////// 右边 开始 ///////////////////////
		tempImgBuf.Reset()
		// 先把右边的存成一张图
		err = clip(bytes.NewReader(fileInBytes), tempImgBuf, imgCfg.Width/2, 0, imgCfg.Width, imgCfg.Height)
		if err != nil {
			fmt.Printf("create clip error[%s] %s\n", files[i], err.Error())
			os.Exit(3)
		}

		pdf.AddPage()
		pdf.SetMarginLeft(width - 20)
		pdf.SetMarginTop(height - 2)
		pdf.Text(fmt.Sprintf("%d/%d", i * 2 + 1, len(files) * 2))

		imgWidth := Px2Pt(float64(imgCfg.Width)) / 2
		// 四周留 20 mm 的边距
		w := min(width - 20, imgWidth)
		var scale float64 = 1
		if w != imgWidth {
			scale = w / imgWidth
		}

		// 根据宽度同比缩放
		imgHeight := Px2Pt(float64(imgCfg.Height)) * scale
		h := min(height - 10, imgHeight)

		// 存在一种可能就是，按宽度缩放之后的高度，还是大于目标高枝，所以需要再次对宽度做同比缩放
		scale = 1
		if h != imgHeight {
			scale = h / imgHeight
		}

		// 根据高度同比缩放
		w = w * scale

		pdf.ImageByHolder(
			&imageBuff{Reader: tempImgBuf, id: files[i] + "_right"},
			(width - w) / 2,
			(height - h) / 2,
			&gopdf.Rect{W: w, H: h})
		/////////////////////// 右边 结束 ///////////////////////

		// 再把左边的存成一张图

		/////////////////////// 左边 开始 ///////////////////////
		tempImgBuf.Reset()
		// 先把右边的存成一张图
		err = clip(bytes.NewReader(fileInBytes), tempImgBuf, 0, 0, imgCfg.Width/2, imgCfg.Height)
		if err != nil {
			fmt.Printf("create clip error[%s] %s\n", files[i], err.Error())
			os.Exit(3)
		}

		pdf.AddPage()
		pdf.SetMarginLeft(width - 20)
		pdf.SetMarginTop(height - 2)
		pdf.Text(fmt.Sprintf("%d/%d", i * 2 + 2, len(files) * 2))

		pdf.ImageByHolder(
			&imageBuff{Reader: tempImgBuf, id: files[i] + "_left"},
			(width - w) / 2,
			(height - h) / 2,
			&gopdf.Rect{W: w, H: h})
		/////////////////////// 左边 结束 ///////////////////////
	}

	if *test {
		return
	}

	fmt.Println("saving to ", *as, " ...")
	pdf.WritePdf(*as)
	fmt.Println("-----------------------")
	fmt.Println("Done, have fun ;)")
	fmt.Println("Created by Mohammed Al Ashaal <https://www.alash3al.xyz>")
}

type imageBuff struct {
	id string
	io.Reader
}
func (i *imageBuff) ID() string {
	return i.id
}

func getImgConfig(r io.Reader) (*image.Config, error) {
	cfg, _, err := image.DecodeConfig(r)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func Px2Pt(px float64) float64 {
	return px * (float64(3)/float64(4))
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func clip(in io.Reader, out io.Writer, x0, y0, x1, y1 int) error {
	origin, fm, err := image.Decode(in)
	if err != nil {
		return err
	}

	switch fm {
	case "jpeg":
		img := origin.(*image.YCbCr)
		subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.YCbCr)
		return jpeg.Encode(out, subImg, nil)
	case "png":
		switch origin.(type) {
		case *image.NRGBA:
			img := origin.(*image.NRGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.NRGBA)
			return png.Encode(out, subImg)
		case *image.RGBA:
			img := origin.(*image.RGBA)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.RGBA)
			return png.Encode(out, subImg)
		case *image.Paletted:
			img := origin.(*image.Paletted)
			subImg := img.SubImage(image.Rect(x0, y0, x1, y1)).(*image.Paletted)
			return png.Encode(out, subImg)
		}

		return fmt.Errorf("unsupport sub format of png")
	default:
		return fmt.Errorf("unsupport format")
	}
}
