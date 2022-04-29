package main

import (
	"fmt"
	// 	"bytes"
	"image"
	// 	"image/color"
	"log"
	"os"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"gocv.io/x/gocv"
)

type tomlConfig struct {
	URL map[string]url
	SaveCount int
	ContourArea float64
}
type url struct {
	DevID string
	Rtsp  string
}

func main() {
	tomlPath := "./conf.toml"
	imgPath := "PICTURE"
	var tc tomlConfig
	var wg sync.WaitGroup
	if _, err := toml.DecodeFile(tomlPath, &tc); err != nil {
		log.Fatal(err)
	}
	nv := 1
	for _, v := range tc.URL {
		newPath := fmt.Sprintf("%s/%s/%s", imgPath, v.DevID, time.Now().Format("20060102"))
		if Exists(newPath) == false {
			os.MkdirAll(newPath, os.ModePerm)
			log.Printf("mkdir suc %v", newPath)
		}
		log.Printf("%v--%v--%v", newPath, v.DevID, v.Rtsp)
		wg.Add(1)
		go func() {
			webcam, err := gocv.OpenVideoCapture(v.Rtsp)
			if err != nil {
				log.Fatalf("open webcam error:%v", err)
			}

			defer webcam.Close()
			defer wg.Done()

			img := gocv.NewMat()
			defer img.Close()

			// window := gocv.NewWindow(fmt.Sprintf("win_%d", nv))
			// window.ResizeWindow(640, 360)
			// defer window.Close()

			nv++
			defaultCount := 3
			count := defaultCount
			fps := webcam.Get(gocv.VideoCaptureFPS)
			durCount := int(fps / 2)
			durBegin := 0
			saveBegin := 0
			var bgMat []gocv.Mat

			background := img.Clone()
			defer background.Close()
			rectimg := img.Clone()
			defer rectimg.Close()
			for {
				if Exists(newPath) == false {
					os.MkdirAll(newPath, os.ModePerm)
					log.Printf("mkdir suc %v", newPath)
				}
				// fmt.Printf("initial MatProfile count: %v\n,fps:%v,durCount:%v\n", gocv.MatProfile.Count(), fps, durBegin)
				// var b bytes.Buffer
				// gocv.MatProfile.WriteTo(&b, 1)
				// fmt.Print(b.String())
				if durBegin < durCount {
					durBegin++
					continue
				} else {
					durBegin = 0
				}
				if ok := webcam.Read(&img); !ok || img.Empty() {
					log.Println("read device error")
					continue
				}
				gray := img.Clone()
				gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
				// 高斯滤波
				gocv.GaussianBlur(gray, &gray, image.Point{5, 5}, 0, 0, gocv.BorderConstant)
				if count == defaultCount {
					background = gray
				}
				frameDelta := img.Clone()
				gocv.AbsDiff(background, gray, &frameDelta)
				gocv.Threshold(frameDelta, &frameDelta, 25, 255, gocv.ThresholdBinary)
				cnts := gocv.FindContours(frameDelta, gocv.RetrievalExternal, gocv.ChainApproxSimple)
				// gocv.Polylines(&img, cnts, true, color.RGBA{0, 0, 255, 0}, 0)
				// fmt.Println(cnts.Size())
				for i := 0; i < cnts.Size(); i++ {
					pt := cnts.At(i)
					c := gocv.ContourArea(pt)
					if c>500{
					    log.Printf("[count] %v--%v",c,saveBegin)
					}
					if c < tc.ContourArea {
						continue
					}
					// 	rect := gocv.BoundingRect(pt)
					// 	gocv.Rectangle(&rectimg, rect, color.RGBA{0, 0, 255, 0}, 3)
					if saveBegin < tc.SaveCount {
						saveBegin++
						continue
					} else {
						saveBegin = 0
						file := fmt.Sprintf(newPath+"/%d.jpg", time.Now().UnixNano())
						log.Printf("[file] %v", file)
						err := gocv.IMWrite(file, img)
						if err != true {
							fmt.Println("write err")
						}
					}
				}
				// if rectimg.Empty() {
				// 	rectimg = img
				// }
				frameDelta.Close()
				if count == defaultCount {
					for _, v := range bgMat {
						v.Close()
					}
					if len(bgMat) > 0 {
						bgMat = bgMat[1:]
					}
					bgMat = append(bgMat, gray)
					count--
				} else if count <= 1 {
					count = defaultCount
					gray.Close()
				} else {
					count--
					gray.Close()
				}
				// window.IMShow(rectimg)
				// window.WaitKey(1)
			}
		}()
		time.Sleep(1 * time.Second)
	}
	wg.Wait()
}

// Exists 判断文件夹是否存在
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
