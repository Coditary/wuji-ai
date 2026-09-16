package cli

import (
	"fmt"

	"github.com/coditary/wuji-core/pkg/batchplan"
	"github.com/coditary/wuji-core/pkg/driver"
)

type imageBatchOptions struct {
	total           int
	size            int
	count           int
	sizeSet         bool
	countSet        bool
	width           int
	height          int
	controlUnits    int
	availableVRAMMB int
}

func resolveImageBatch(opts imageBatchOptions, task driver.ImageTask) (size, count int, err error) {
	if opts.total > 0 && !taskSupportsBatch(task) {
		return 0, 0, fmt.Errorf("--batch is not supported for image task %q", task)
	}

	res, err := batchplan.Resolve(batchplan.Input{
		Total:           opts.total,
		Size:            opts.size,
		Count:           opts.count,
		SizeSet:         opts.sizeSet,
		CountSet:        opts.countSet,
		Width:           opts.width,
		Height:          opts.height,
		ControlUnits:    opts.controlUnits,
		AvailableVRAMMB: opts.availableVRAMMB,
	})
	if err != nil {
		return 0, 0, err
	}
	return res.Size, res.Count, nil
}

func taskSupportsBatch(task driver.ImageTask) bool {
	switch task {
	case driver.ImageTaskUpscale:
		return false
	default:
		return true
	}
}
