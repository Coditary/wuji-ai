package clix

import (
	"fmt"

	"github.com/coditary/wuji-core/pkg/batchplan"
	"github.com/coditary/wuji-core/pkg/driver"
)

type ImageBatchOptions struct {
	Total           int
	Size            int
	Count           int
	SizeSet         bool
	CountSet        bool
	Width           int
	Height          int
	ControlUnits    int
	AvailableVRAMMB int
}

func ResolveImageBatch(opts ImageBatchOptions, task driver.ImageTask) (size, count int, err error) {
	if opts.Total > 0 && !TaskSupportsBatch(task) {
		return 0, 0, fmt.Errorf("--batch is not supported for image task %q", task)
	}

	res, err := batchplan.Resolve(batchplan.Input{
		Total:           opts.Total,
		Size:            opts.Size,
		Count:           opts.Count,
		SizeSet:         opts.SizeSet,
		CountSet:        opts.CountSet,
		Width:           opts.Width,
		Height:          opts.Height,
		ControlUnits:    opts.ControlUnits,
		AvailableVRAMMB: opts.AvailableVRAMMB,
	})
	if err != nil {
		return 0, 0, err
	}
	return res.Size, res.Count, nil
}

func TaskSupportsBatch(task driver.ImageTask) bool {
	switch task {
	case driver.ImageTaskUpscale:
		return false
	default:
		return true
	}
}
