package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"time"
)

func main() {
	os.Exit(run())
}

func run() int {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s image_file\n", os.Args[0])
		return 1
	}
	filePath := os.Args[1]
	context, err := ExtractContext(filePath)
	if err != nil {
		return 2
	}
	//fmt.Println(context)

	// 第0世代コンテクストの保存
	context.Save("stage0.json")

	outGif := &gif.GIF{}

	// 世代数かつコンテクスト内のライフがある限り繰り返し
	for i := 0; i < context.Epoch && context.CountLives() > 0; i++ {
		if context.Verbose {
			context.Print()
		}
		if context.OutFrames {
			err = context.SaveImage(fmt.Sprintf("out%03d.jpg", i))
			if err != nil {
				return 2
			}
		}
		img := context.GetImage()
		outGif.Image = append(outGif.Image, img.(*image.Paletted))
		if i == 0 {
			outGif.Delay = append(outGif.Delay, 100)
		} else {
			outGif.Delay = append(outGif.Delay, 0)
		}
		context.Next()
	}

	// Gif アニメーションをファイルに出力する
	var w *os.File
	w, err = os.Create("out.gif")
	if err != nil {
		fmt.Println(err)
		return 3
	}
	defer w.Close()
	gif.EncodeAll(w, outGif)

	return 0
}

func ExtractContext(filePath string) (context *Context, err error) {
	var f *os.File
	f, err = os.Open(filePath)
	if err != nil {
		return
	}
	defer f.Close()

	var img image.Image
	var format string
	img, format, err = image.Decode(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	fmt.Println("format =", format)
	fmt.Println("Min.X =", img.Bounds().Min.X)
	fmt.Println("Min.Y =", img.Bounds().Min.Y)
	fmt.Println("Max.X =", img.Bounds().Max.X)
	fmt.Println("Max.Y =", img.Bounds().Max.Y)
	width := img.Bounds().Max.X
	height := img.Bounds().Max.Y
	lives := make([]int, width*height)

	for i := 0; i < width*height; i++ {
		x := i % width
		y := i / width
		r, g, b, a := img.At(x, y).RGBA()
		//fmt.Printf("[%04X,%04X,%04X,%04X]\n", r, g, b, a)
		c := int(math.Sqrt(float64(r*r + g*g + b*b)))
		//fmt.Printf("[%04X,%04X]\n", c, a)
		if c >= int(a)/2 {
			lives[i] = 1
		}
	}
	context = &Context{
		Epoch:  300,
		Width:  width,
		Height: height,
		Lives:  lives,
	}
	return
}

func xrun() int {

	/*
		width := 10
		height := 9
		initLives := []int{
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
			0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
		}
		Context := New(width, height, initLives)
	*/

	// 設定ファイルの読み込み
	context, err := Load("input.json")
	if err != nil {
		fmt.Println(err)
		return 1
	}

	// 第0世代コンテクストの保存
	context.Save("stage0.json")

	outGif := &gif.GIF{}

	// 世代数かつコンテクスト内のライフがある限り繰り返し
	for i := 0; i < context.Epoch && context.CountLives() > 0; i++ {
		if context.Verbose {
			context.Print()
		}
		if context.OutFrames {
			err = context.SaveImage(fmt.Sprintf("out%03d.jpg", i))
			if err != nil {
				return 2
			}
		}
		img := context.GetImage()
		outGif.Image = append(outGif.Image, img.(*image.Paletted))
		outGif.Delay = append(outGif.Delay, 0)
		context.Next()
	}

	// Gif アニメーションをファイルに出力する
	var w *os.File
	w, err = os.Create("out.gif")
	if err != nil {
		fmt.Println(err)
		return 3
	}
	defer w.Close()
	gif.EncodeAll(w, outGif)

	return 0
}

// Context : ライフゲームのコンテクストを保持するデータ
type Context struct {
	Epoch     int   `json:"epoch"`
	Width     int   `json:"width"`
	Height    int   `json:"height"`
	Lives     []int `json:"lives"`
	Ops       []int `json:"ops"`
	OutFrames bool  `json:"out_frames"`
	Verbose   bool  `json:"verbose"`
	nextLives []int
}

// New : 新たなコンテクストデータを作成する
func New(width, height int, initLives []int) *Context {
	if width < 1 || height < 1 || width*height != len(initLives) {
		return nil
	}
	return &Context{
		Width:  width,
		Height: height,
		Lives:  initLives,
	}
}

// Load ライフゲームの過去のコンテクストデータをファイルから読み出す
func Load(filePath string) (r *Context, err error) {
	var bytes []byte
	bytes, err = ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	var context Context
	r = &context
	err = json.Unmarshal(bytes, &context)
	if err != nil {
		return
	}
	if context.Width < 1 || context.Height < 1 {
		err = fmt.Errorf("Abnormal context data. width=%d, height=%d, len(lives)=%d", context.Width, context.Height, len(context.Lives))
		return
	}
	if len(context.Lives) < context.Width*context.Height {
		tmp := make([]int, context.Width*context.Height)
		for i := 0; i < len(context.Lives); i++ {
			tmp[i] = context.Lives[i]
		}
		context.Lives = tmp
	} else if len(context.Lives) > context.Width*context.Height {
		context.Lives = context.Lives[0 : context.Width*context.Height]
	}

	rand.Seed(time.Now().UnixNano())

	// 初期ステージでのライフのオペレーション
	for _, op := range context.Ops {
		switch op {
		case 0: // ランダムな座標のライフを 1 にする
			pos := rand.Intn(len(context.Lives))
			context.Lives[pos] = 1
		case 1: // ランダムな座標のライフを 0 にする
			pos := rand.Intn(len(context.Lives))
			context.Lives[pos] = 0
		case 2: // ランダムな y 座標の横一行のライフを 1 にする
			y := rand.Intn(context.Height)
			for x := 0; x < context.Width; x++ {
				pos := y*context.Width + x
				context.Lives[pos] = 1
			}
		case 3: // ランダムな x 座標の縦一列のライフを 1 にする
			x := rand.Intn(context.Width)
			for y := 0; y < context.Height; y++ {
				pos := y*context.Width + x
				//fmt.Println("x =", x, "y =", y, "pos =", pos, "len(Lives) =", len(context.Lives))
				context.Lives[pos] = 1
			}
		}
	}
	return
}

// Save コンテクストデータをファイルに書き出す
func (l *Context) Save(filePath string) (err error) {
	// Ops をリセットしておく
	l.Ops = []int{}
	var bytes []byte
	bytes, err = json.MarshalIndent(l, "", "  ")
	if err != nil {
		return
	}
	var w *os.File
	w, err = os.Create(filePath)
	if err != nil {
		return
	}
	defer w.Close()
	w.Write(bytes)
	return
}

// Get : 指定された座標のライフを取得する
func (l *Context) Get(x, y int) int {
	return l.Lives[y*l.Width+x]
}
func (l *Context) setNext(x, y, v int) {
	if v > 0 {
		v = 1
	} else {
		v = 0
	}
	l.nextLives[y*l.Width+x] = v
}

// Neibor : 指定された座標の近傍のライフの数を取得する
func (l *Context) Neibor(x, y int) int {
	count := 0
	// NW
	if x > 0 && y > 0 && l.Get(x-1, y-1) == 1 {
		count++
	}
	// N
	if y > 0 && l.Get(x, y-1) == 1 {
		count++
	}
	// NE
	if x < l.Width-1 && y > 0 && l.Get(x+1, y-1) == 1 {
		count++
	}
	// E
	if x < l.Width-1 && l.Get(x+1, y) == 1 {
		count++
	}
	// SE
	if x < l.Width-1 && y < l.Height-1 && l.Get(x+1, y+1) == 1 {
		count++
	}
	// S
	if y < l.Height-1 && l.Get(x, y+1) == 1 {
		count++
	}
	// SW
	if x > 0 && y < l.Height-1 && l.Get(x-1, y+1) == 1 {
		count++
	}
	// W
	if x > 0 && l.Get(x-1, y) == 1 {
		count++
	}
	return count
}

// CountLives : ライフの個数を計算する
func (l *Context) CountLives() (r int) {
	for i := 0; i < len(l.Lives); i++ {
		r += l.Lives[i]
	}
	return
}

// Print : コンテクストを標準出力に表示する
func (l *Context) Print() {
	fmt.Println("------------------")
	for i := 0; i < len(l.Lives); i++ {
		fmt.Print(l.Lives[i])
		if i%l.Width == l.Width-1 {
			fmt.Println("")
		}
	}
}

// Next : コンテクストを次の世代に更新する
func (l *Context) Next() {
	l.nextLives = make([]int, len(l.Lives))
	for i := 0; i < len(l.Lives); i++ {
		x := i % l.Width
		y := i / l.Width
		neibor := l.Neibor(x, y)
		life := l.Get(x, y)
		l.setNext(x, y, rule(life, neibor))
		/*
			if l.Get(x, y) == 1 {
				switch neibor {
				case 2, 3:

				}
			} else {
				if neibor == 3 {
					l.setNext(x, y, 1)
				}
			}*/
	}
	l.Lives = l.nextLives
}

func rule(life, neibor int) (nextLife int) {
	if life == 1 {
		switch neibor {
		case 2, 3:
			nextLife = 1
		}
	} else {
		if neibor == 3 {
			nextLife = 1
		}
	}
	return
}

func (l *Context) SaveImage(filePath string) (err error) {
	img := image.NewRGBA(image.Rect(0, 0, l.Width, l.Height))
	col := color.RGBA{255, 255, 255, 0}
	for i := 0; i < len(l.Lives); i++ {
		if l.Lives[i] == 0 {
			continue
		}
		x := i % l.Width
		y := i / l.Width
		img.Set(x, y, col)
	}
	// 出力用ファイル作成(エラー処理は略)
	var w *os.File
	w, err = os.Create(filePath)
	if err != nil {
		return
	}
	defer w.Close()

	// JPEGで出力(100%品質)
	err = jpeg.Encode(w, img, &jpeg.Options{Quality: 100})
	return
}

func (l *Context) GetImage() (r image.Image) {
	var palette = []color.Color{
		color.RGBA{0x00, 0x00, 0x00, 0xff},
		color.RGBA{0x00, 0x00, 0xff, 0xff},
		color.RGBA{0x00, 0xff, 0x00, 0xff},
		color.RGBA{0x00, 0xff, 0xff, 0xff},
		color.RGBA{0xff, 0x00, 0x00, 0xff},
		color.RGBA{0xff, 0x00, 0xff, 0xff},
		color.RGBA{0xff, 0xff, 0x00, 0xff},
		color.RGBA{0xff, 0xff, 0xff, 0xff},
	}

	//img := image.NewRGBA(image.Rect(0, 0, l.Width, l.Height))
	img := image.NewPaletted(image.Rect(0, 0, l.Width, l.Height), palette)
	col := palette[7]
	for i := 0; i < len(l.Lives); i++ {
		if l.Lives[i] == 0 {
			continue
		}
		x := i % l.Width
		y := i / l.Width
		img.Set(x, y, col)
	}
	r = img
	return
}
