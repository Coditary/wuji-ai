package clix

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/driver"
)

type DataTaskFlags struct {
	Embed        bool
	Vector       bool
	Detect       bool
	Segment      bool
	NER          bool
	Keypoint     bool
	Classify     bool
	Regress      bool
	Forecast     bool
	Graph        bool
	NodeClassify bool
	LinkPredict  bool
}

func (f DataTaskFlags) Resolve() (driver.DataTask, error) {
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
		{f.Embed, driver.DataTaskEmbed},
		{f.Vector, driver.DataTaskVector},
		{f.Detect, driver.DataTaskDetect},
		{f.Segment, driver.DataTaskSegment},
		{f.NER, driver.DataTaskNER},
		{f.Keypoint, driver.DataTaskKeypoint},
		{f.Classify, driver.DataTaskClassify},
		{f.Regress, driver.DataTaskRegress},
		{f.Forecast, driver.DataTaskForecast},
		{f.Graph, driver.DataTaskGraph},
		{f.NodeClassify, driver.DataTaskNodeClass},
		{f.LinkPredict, driver.DataTaskLinkPredict},
	} {
		if err := pick(c.set, c.task); err != nil {
			return "", err
		}
	}
	return task, nil
}

func AddDataTaskFlags(cmd *cobra.Command, f *DataTaskFlags) {
	cmd.Flags().BoolVar(&f.Embed, "embed", false, "text/image embeddings")
	cmd.Flags().BoolVar(&f.Vector, "vector", false, "alias for --embed")
	cmd.Flags().BoolVar(&f.Detect, "detect", false, "object detection on --image")
	cmd.Flags().BoolVar(&f.Segment, "segment", false, "image segmentation on --image")
	cmd.Flags().BoolVar(&f.NER, "ner", false, "named entity recognition on text")
	cmd.Flags().BoolVar(&f.Keypoint, "keypoint", false, "keypoint detection on --image")
	cmd.Flags().BoolVar(&f.Classify, "classify", false, "tabular classification on --csv-file")
	cmd.Flags().BoolVar(&f.Regress, "regress", false, "tabular regression on --csv-file")
	cmd.Flags().BoolVar(&f.Forecast, "forecast", false, "time series forecast on --csv-file")
	cmd.Flags().BoolVar(&f.Graph, "graph", false, "graph inference")
	cmd.Flags().BoolVar(&f.NodeClassify, "node-classify", false, "node classification on --graph-file")
	cmd.Flags().BoolVar(&f.LinkPredict, "link-predict", false, "link prediction on --graph-file")
}
