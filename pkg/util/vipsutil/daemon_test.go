package vipsutil

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
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

type ErrorReader struct {
	Error error
}

func (r ErrorReader) Read(p []byte) (n int, err error) {
	err = r.Error
	return
}

func TestDaemonGoroutineCharacteristics(t *testing.T) {
	LibvipsInit()

	// Tests in different packages are run in parallel.
	// So comparing number of goroutine inside a test is very fragile and often fail.
	SkipConvey("Daemon goroutine characteristics", t, func() {
		// Since tests in different packages are run in parallel.
		// If we compare the exact number of running goroutine,
		// this test may sometimes fail.
		// Adding some skew will make the test less likely to fail.
		SKEW := 5

		// n0 is the initial number of running goroutine.
		n0 := runtime.NumGoroutine()

		numWorker := 4
		d := OpenDaemon(numWorker)

		// n1 is the number of running goroutine after the daemon is opened.
		// n1 = n0 + numWorker
		n1 := runtime.NumGoroutine()
		So(n1-n0, ShouldAlmostEqual, numWorker, SKEW)

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
		So(n2-n1, ShouldAlmostEqual, numRequest, SKEW)

		wg.Wait()

		// n3 is the number of running goroutine after all requests have been processed.
		// n3 = n1
		n3 := runtime.NumGoroutine()
		So(n3, ShouldAlmostEqual, n1, SKEW)

		_ = d.Close()
		// Allow the daemon threads to run and die.
		runtime.Gosched()

		// n4 is the number of running goroutine after the daemon is closed.
		// n4 = n0
		n4 := runtime.NumGoroutine()
		So(n4, ShouldAlmostEqual, n0, SKEW)
	})
}

func TestDaemonProcess(t *testing.T) {
	LibvipsInit()

	Convey("Daemon Process", t, func() {
		numWorker := 1
		d := OpenDaemon(numWorker)
		defer d.Close()

		f, err := os.Open("testdata/image-cat-coffee.jpg")
		So(err, ShouldBeNil)
		defer f.Close()

		input := Input{
			Reader: f,
			Options: Options{
				Width:            500,
				Height:           500,
				ResizingModeType: ResizingModeTypeCover,
			},
		}

		output, err := d.Process(input)
		So(err, ShouldBeNil)

		expected, err := os.Open("testdata/image-cat-coffee.expected.jpg")
		So(err, ShouldBeNil)
		defer expected.Close()

		expectedImage, err := jpeg.Decode(expected)
		So(err, ShouldBeNil)

		actualImage, err := jpeg.Decode(bytes.NewReader(output.Data))
		So(err, ShouldBeNil)

		So(expectedImage.Bounds(), ShouldResemble, actualImage.Bounds())
	})

	Convey("Daemon Process does not panic on invalid input", t, func() {
		numWorker := 1
		d := OpenDaemon(numWorker)
		defer d.Close()

		_, err := d.Process(Input{
			Reader: bytes.NewBuffer(nil),
			Options: Options{
				Width:            500,
				Height:           500,
				ResizingModeType: ResizingModeTypeCover,
			},
		})
		So(err, ShouldNotBeNil)
	})

	Convey("Daemon Process does not panic on io error", t, func() {
		numWorker := 1
		d := OpenDaemon(numWorker)
		defer d.Close()

		_, err := d.Process(Input{
			Reader: ErrorReader{
				Error: fmt.Errorf("some io error"),
			},
			Options: Options{
				Width:            500,
				Height:           500,
				ResizingModeType: ResizingModeTypeCover,
			},
		})
		So(err, ShouldBeError, "some io error")
	})

	Convey("Daemon Process does not error on zero options", t, func() {
		numWorker := 1
		d := OpenDaemon(numWorker)
		defer d.Close()

		f, err := os.Open("testdata/image-cat-coffee.jpg")
		So(err, ShouldBeNil)
		defer f.Close()

		input := Input{
			Reader: f,
		}

		_, err = d.Process(input)
		So(err, ShouldBeNil)
	})
}
