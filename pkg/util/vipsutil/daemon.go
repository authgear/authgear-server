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
	ImageMetadata *vips.ImageMetadata
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

func NewDaemon(numWorker int) *Daemon {
	return &Daemon{
		numWorker: numWorker,
		queue:     make(chan task),
	}
}

func (d *Daemon) Open() {
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
}

func (v *Daemon) runInput(i Input) (o *Output, err error) {
	imageRef, err := vips.NewImageFromReader(i.Reader)
	if err != nil {
		return
	}

	// FIXME(vips): process the input according to options.
	data, metadata, err := imageRef.ExportNative()
	if err != nil {
		return
	}

	o = &Output{
		ImageMetadata: metadata,
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
