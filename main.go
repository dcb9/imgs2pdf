package main

import (
	"bytes"
	"fmt"
	"github.com/gobuffalo/packr/v2"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"regexp"
)
import "flag"
import "path/filepath"
import "github.com/signintech/gopdf"

var (
	src = flag.String("src", "./", "the images directory path (png, jpg & jpeg) files")
	as  = flag.String("as", "./result.pdf", "the result filename")
	test  = flag.Bool("t", false, "for test")
	abort = flag.Bool("abort", true, "abort if error occurs")
	pageRegStr = flag.String("page-reg", "", "page regex matches page from filename")
)

func init() {
	flag.Parse()
}

var pageReg *regexp.Regexp
var width float64 = 210
var height float64 = 157
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

	_tr := bytes.NewReader(timesTTF)
	_ = _tr
	//if err = pdf.AddTTFFont(font, "times.ttf"); err != nil {
	if err = pdf.AddTTFFontByReader(font, bytes.NewReader(timesTTF)); err != nil {
		fmt.Printf("Add font[%s] err > %s\n", font, err.Error())
		os.Exit(1)
	}

	err = pdf.SetFont(font, "", 14)
	if err != nil {
		fmt.Printf("Set font err > %s\n", err.Error())
		os.Exit(2)
	}

	if *pageRegStr != "" {
		pageReg, err = regexp.Compile(*pageRegStr)
		if err != nil {
			fmt.Printf("invalid reg expression: %s with error: %s\n", *pageRegStr, err.Error())
			os.Exit(1)
		}
	}

	fmt.Println("Reading files from (", *src, ") and saving the result as (", *as,")")
	fmt.Println("-----------------------")


	jpgs, _ := filepath.Glob(*src + "/*.jpg")
	pngs, _ := filepath.Glob(*src + "/*.png")
	jpegs, _ := filepath.Glob(*src + "/*.jpeg")
	files := append(jpgs, append(jpegs, pngs...)...)
	for i := 0; i < len(files); i++ {
		fmt.Println(i+1, ")- adding ", files[i])
		if *test {
			continue
		}
		x := float64(0)
		if x < 0 {
			continue
		}
		pdf.AddPage()

		pdf.SetMarginLeft(width - 10)
		pdf.SetMarginTop(height - 2)
		// use file name as page
		page := getPage(files[i], pageReg)
		pdf.Text(page)

		imgCfg, err := getImgConfig(files[i])
		if err != nil {
			fmt.Printf("get img config[%s] > %s\n", files[i], err.Error())
			if *abort {
				os.Exit(3)
			}
		}

		imgWidth := Px2Pt(float64(imgCfg.Width))
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


		// 通过对 x 和 y 的调整，达到让图片位于中心点
		pdf.Image(files[i], (width - w) / 2, (height - h) / 2, &gopdf.Rect{W: w, H: h})
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

func getPage(path string, reg *regexp.Regexp) string {
	//regexp.Compile()
	baseName := filepath.Base(path)
	ext := filepath.Ext(baseName)

	page := baseName[:len(baseName) - len(ext)] // 去掉 extension
	if reg != nil {
		matches :=reg.FindSubmatch([]byte(page))
		if len(matches) >= 2 {
			page = string(matches[1])
		}
	}
	return page
}

func getImgConfig(imgPath string) (*image.Config, error) {
	reader, err := os.Open(imgPath)
	if err != nil {
		return nil, fmt.Errorf("open file > %w", err)
	}
	defer reader.Close()

	cfg, _, err := image.DecodeConfig(reader)
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
