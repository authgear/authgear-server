package vipsutil

import (
	"io"

	"github.com/davidbyttow/govips/v2/vips"
)

type Input struct {
	Reader  io.Reader
	Options Options
}

type Output struct {
	FileExtension string
	Data          []byte
}

type task struct {
	Input      Input
	OutputChan chan interface{}
}

type Daemon struct {
	numWorker int
	queue     chan task
}

var _ io.Closer = &Daemon{}

func OpenDaemon(numWorker int) *Daemon {
	d := &Daemon{
		numWorker: numWorker,
		queue:     make(chan task),
	}
	for i := 0; i < d.numWorker; i++ {
		go func(workerID int) {
			for task := range d.queue {
				output, err := d.runInput(task.Input)
				if err != nil {
					task.OutputChan <- err
				} else {
					task.OutputChan <- output
				}
			}
		}(i)
	}
	return d
}

func (v *Daemon) runInput(i Input) (o *Output, err error) {
	imageRef, err := vips.NewImageFromReader(i.Reader)
	if err != nil {
		return
	}

	// Consume Exif orientation.
	err = imageRef.AutoRotate()
	if err != nil {
		return
	}

	// Resize only if it is OK to do so.
	// This checking avoids division by zero.
	if i.Options.ShouldResize() {
		resizeMode := ResizingModeFromType(i.Options.ResizingModeType)
		resizeDimen := ResizeDimensions{
			Width:  i.Options.Width,
			Height: i.Options.Height,
		}
		imageDimen := ImageDimensions{
			Width:  imageRef.Metadata().Width,
			Height: imageRef.Metadata().Height,
		}
		resizeResult := resizeMode.Resize(imageDimen, resizeDimen)
		err = applyResize(resizeResult, imageRef, vips.KernelAuto)
		if err != nil {
			return
		}
	}

	data, metadata, err := export(imageRef)
	if err != nil {
		return
	}

	o = &Output{
		FileExtension: metadata.Format.FileExt(),
		Data:          data,
	}
	return
}

func (v *Daemon) Close() error {
	close(v.queue)
	return nil
}

func (v *Daemon) Process(i Input) (*Output, error) {
	task := task{
		Input:      i,
		OutputChan: make(chan interface{}),
	}

	v.queue <- task

	result := <-task.OutputChan
	switch result := result.(type) {
	case error:
		return nil, result
	case *Output:
		return result, nil
	default:
		panic("unreachable")
	}
}

func export(imageRef *vips.ImageRef) ([]byte, *vips.ImageMetadata, error) {
	imageType := imageRef.Format()
	switch imageType {
	case vips.ImageTypeJPEG:
		return imageRef.ExportJpeg(&vips.JpegExportParams{
			StripMetadata: true,
			Quality:       80,
			Interlace:     true,
			SubsampleMode: vips.VipsForeignSubsampleOn,
		})
	case vips.ImageTypePNG:
		return imageRef.ExportPng(&vips.PngExportParams{
			StripMetadata: true,
			Compression:   6,
		})
	case vips.ImageTypeGIF:
		return imageRef.ExportGIF(&vips.GifExportParams{
			StripMetadata: true,
			Quality:       75,
		})
	case vips.ImageTypeWEBP:
		return imageRef.ExportWebp(&vips.WebpExportParams{
			StripMetadata:   true,
			Quality:         75,
			ReductionEffort: 4,
		})
	default:
		return imageRef.ExportNative()
	}
}

func applyResize(r ResizeResult, imageRef *vips.ImageRef, kernel vips.Kernel) (err error) {
	if r.Scale != 1.0 {
		err = imageRef.Resize(r.Scale, kernel)
		if err != nil {
			return
		}
	}

	if r.Crop != nil {
		dx := r.Crop.Dx()
		dy := r.Crop.Dy()
		x := r.Crop.Min.X
		y := r.Crop.Min.Y
		err = imageRef.ExtractArea(x, y, dx, dy)
		if err != nil {
			return
		}
	}

	return nil
}
