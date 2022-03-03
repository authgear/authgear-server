package vipsutil

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"runtime"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func NewPNG(width int, height int) []byte {
	c := color.RGBA{}
	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			img.Set(x, y, c)
		}
	}
	w := bytes.Buffer{}
	_ = png.Encode(&w, img)
	return w.Bytes()
}

func TestDaemonGoroutineCharacteristics(t *testing.T) {
	LibvipsInit()

	Convey("Daemon goroutine characteristics", t, func() {
		numWorker := 4
		d := NewDaemon(numWorker)

		// n0 is the initial number of running goroutine.
		n0 := runtime.NumGoroutine()
		d.Open()

		// n1 is the number of running goroutine after the daemon is opened.
		// n1 = n0 + numWorker
		n1 := runtime.NumGoroutine()
		So(n1-n0, ShouldEqual, numWorker)

		numRequest := 100
		var wg sync.WaitGroup

		for i := 0; i < numRequest; i++ {
			wg.Add(1)
			go func(i int) {
				img := NewPNG(1, 1)
				input := Input{Reader: bytes.NewReader(img)}
				_, _ = d.Process(input)
				wg.Done()
			}(i)
		}

		// n2 is the number of running goroutine after some requests are submitted to the daemon.
		// n2 = n1 + numRequest
		n2 := runtime.NumGoroutine()
		So(n2-n1, ShouldEqual, numRequest)

		wg.Wait()

		// n3 is the number of running goroutine after all requests have been processed.
		// n3 = n1
		n3 := runtime.NumGoroutine()
		So(n3, ShouldEqual, n1)

		_ = d.Close()
		// Allow the daemon threads to run and die.
		runtime.Gosched()

		// n4 is the number of running goroutine after the daemon is closed.
		// n4 = n0
		n4 := runtime.NumGoroutine()
		So(n4, ShouldEqual, n0)
	})
}
