package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/driver"
)

type dataTaskFlags struct {
	embed        bool
	vector       bool
	detect       bool
	segment      bool
	ner          bool
	keypoint     bool
	classify     bool
	regress      bool
	forecast     bool
	graph        bool
	nodeClassify bool
	linkPredict  bool
}

func (f dataTaskFlags) resolve() (driver.DataTask, error) {
	var task driver.DataTask
	count := 0
	pick := func(set bool, t driver.DataTask) error {
		if !set {
			return nil
		}
		count++
		if count > 1 {
			return fmt.Errorf("choose one task flag (--embed, --detect, --ner, --forecast, …)")
		}
		task = t
		return nil
	}
	for _, c := range []struct {
		set  bool
		task driver.DataTask
	}{
		{f.embed, driver.DataTaskEmbed},
		{f.vector, driver.DataTaskVector},
		{f.detect, driver.DataTaskDetect},
		{f.segment, driver.DataTaskSegment},
		{f.ner, driver.DataTaskNER},
		{f.keypoint, driver.DataTaskKeypoint},
		{f.classify, driver.DataTaskClassify},
		{f.regress, driver.DataTaskRegress},
		{f.forecast, driver.DataTaskForecast},
		{f.graph, driver.DataTaskGraph},
		{f.nodeClassify, driver.DataTaskNodeClass},
		{f.linkPredict, driver.DataTaskLinkPredict},
	} {
		if err := pick(c.set, c.task); err != nil {
			return "", err
		}
	}
	return task, nil
}

func addDataTaskFlags(cmd *cobra.Command, f *dataTaskFlags) {
	cmd.Flags().BoolVar(&f.embed, "embed", false, "text/image embeddings")
	cmd.Flags().BoolVar(&f.vector, "vector", false, "alias for --embed")
	cmd.Flags().BoolVar(&f.detect, "detect", false, "object detection on --image")
	cmd.Flags().BoolVar(&f.segment, "segment", false, "image segmentation on --image")
	cmd.Flags().BoolVar(&f.ner, "ner", false, "named entity recognition on text")
	cmd.Flags().BoolVar(&f.keypoint, "keypoint", false, "keypoint detection on --image")
	cmd.Flags().BoolVar(&f.classify, "classify", false, "tabular classification on --csv-file")
	cmd.Flags().BoolVar(&f.regress, "regress", false, "tabular regression on --csv-file")
	cmd.Flags().BoolVar(&f.forecast, "forecast", false, "time series forecast on --csv-file")
	cmd.Flags().BoolVar(&f.graph, "graph", false, "graph inference")
	cmd.Flags().BoolVar(&f.nodeClassify, "node-classify", false, "node classification on --graph-file")
	cmd.Flags().BoolVar(&f.linkPredict, "link-predict", false, "link prediction on --graph-file")
}
