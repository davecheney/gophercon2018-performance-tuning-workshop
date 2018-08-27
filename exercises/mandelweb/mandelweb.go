// mandelbrot example code adapted from Francesc Campoy's mandelbrot package.
// https://github.com/campoy/mandelbrot
package main

import (
	"image"
	"image/color"
	"image/png"
	"log"
	"runtime"
	"sync"
	"time"

	"net/http"
	_ "net/http/pprof"
)

func main() {
	http.HandleFunc("/mandelbrot", mandelbrot)
	log.Println("listening on http://127.0.0.1:8080/")
	http.ListenAndServe(":8080", logRequest(http.DefaultServeMux))
}

func logRequest(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		start := time.Now()
		h.ServeHTTP(w, req)
		log.Println(req.RemoteAddr, req.RequestURI, time.Since(start))
	})
}

func mandelbrot(w http.ResponseWriter, req *http.Request) {
	const height, width = 512, 512
	c := make([][]color.RGBA, height)
	for i := range c {
		c[i] = make([]color.RGBA, width)
	}
	img := &img{h: height, w: width, m: c}

	fillImage(img, runtime.NumCPU())
	png.Encode(w, img)
}

type img struct {
	h, w int
	m    [][]color.RGBA
}

func (m *img) At(x, y int) color.Color { return m.m[x][y] }
func (m *img) ColorModel() color.Model { return color.RGBAModel }
func (m *img) Bounds() image.Rectangle { return image.Rect(0, 0, m.h, m.w) }

func fillImage(m *img, workers int) {
	c := make(chan int, m.h)
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for i := range c {
				for j := range m.m[i] {
					fillPixel(m, i, j)
				}
			}
		}()
	}

	for i := range m.m {
		c <- i
	}
	close(c)
	wg.Wait()
}

func fillPixel(m *img, x, y int) {
	const n = 1000
	const Limit = 2.0
	const Zoom = 4
	Zr, Zi, Tr, Ti := 0.0, 0.0, 0.0, 0.0
	Cr := (Zoom*float64(x)/float64(n) - 1.5)
	Ci := (Zoom*float64(y)/float64(n) - 1.0)

	for i := 0; i < n && (Tr+Ti <= Limit*Limit); i++ {
		Zi = 2*Zr*Zi + Ci
		Zr = Tr - Ti + Cr
		Tr = Zr * Zr
		Ti = Zi * Zi
	}
	paint(&m.m[x][y], Tr, Ti)
}

func paint(c *color.RGBA, x, y float64) {
	n := byte(x * y * 4)
	c.R, c.G, c.B, c.A = n, n, n, 255
}
