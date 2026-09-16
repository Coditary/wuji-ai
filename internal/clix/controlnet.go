package clix

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/driver"
)

type controlTypeSpec struct {
	canonical driver.ImageControlType
	aliases   []string
}

var imageControlSpecs = []controlTypeSpec{
	{driver.ImageControlCanny, []string{"canny", "edge"}},
	{driver.ImageControlDepth, []string{"depth"}},
	{driver.ImageControlPose, []string{"pose", "openpose"}},
	{driver.ImageControlLineart, []string{"lineart"}},
	{driver.ImageControlLineartAnime, []string{"lineart-anime"}},
	{driver.ImageControlScribble, []string{"scribble"}},
	{driver.ImageControlNormal, []string{"normal"}},
	{driver.ImageControlSegmentation, []string{"segmentation"}},
	{driver.ImageControlMLSD, []string{"mlsd"}},
	{driver.ImageControlTile, []string{"tile"}},
}

type controlFlagState struct {
	values map[driver.ImageControlType]*[]string
	params map[driver.ImageControlType]*driver.ControlUnitParams
	low    float32
	high   float32
	mode   string
}

type ImageControlOptions struct {
	state controlFlagState
}

func RegisterImageControlFlags(cmd *cobra.Command, opts *ImageControlOptions) {
	st := &opts.state
	st.values = make(map[driver.ImageControlType]*[]string, len(imageControlSpecs))
	st.params = make(map[driver.ImageControlType]*driver.ControlUnitParams, len(imageControlSpecs))

	for _, spec := range imageControlSpecs {
		values := new([]string)
		st.values[spec.canonical] = values
		usage := fmt.Sprintf("%s control: image path or variant (e.g. --%s=full --%s path.png)", spec.canonical, spec.aliases[0], spec.aliases[0])
		for _, name := range spec.aliases {
			cmd.Flags().StringArrayVar(values, name, nil, usage)
		}

		params := new(driver.ControlUnitParams)
		st.params[spec.canonical] = params
		for _, name := range spec.aliases {
			cmd.Flags().Float32Var(&params.Weight, name+"-weight", 0, fmt.Sprintf("weight for --%s / synonym flags (0 = backend default)", name))
			cmd.Flags().StringVar(&params.Model, name+"-model", "", fmt.Sprintf("ControlNet model override for --%s", name))
			cmd.Flags().Float32Var(&params.GuidanceStart, name+"-start", 0, fmt.Sprintf("guidance start for --%s (0 = backend default)", name))
			cmd.Flags().Float32Var(&params.GuidanceEnd, name+"-end", 0, fmt.Sprintf("guidance end for --%s (0 = backend default)", name))
		}
	}

	cmd.Flags().Float32Var(&st.low, "canny-low", 0, "Canny/edge low threshold (0 = backend default)")
	cmd.Flags().Float32Var(&st.low, "edge-low", 0, "alias for --canny-low")
	cmd.Flags().Float32Var(&st.high, "canny-high", 0, "Canny/edge high threshold (0 = backend default)")
	cmd.Flags().Float32Var(&st.high, "edge-high", 0, "alias for --canny-high")
	cmd.Flags().StringVar(&st.mode, "control-mode", "", "control priority: balanced, prompt, control")
}

func (opts ImageControlOptions) finalizeParams(cmd *cobra.Command) {
	st := &opts.state
	for _, spec := range imageControlSpecs {
		p := st.params[spec.canonical]
		for _, name := range spec.aliases {
			if f := cmd.Flags().Lookup(name + "-weight"); f != nil && f.Changed {
				p.WeightSet = true
			}
			if f := cmd.Flags().Lookup(name + "-start"); f != nil && f.Changed {
				p.StartSet = true
			}
			if f := cmd.Flags().Lookup(name + "-end"); f != nil && f.Changed {
				p.EndSet = true
			}
		}
		if spec.canonical == driver.ImageControlCanny {
			for _, name := range []string{"canny-low", "edge-low"} {
				if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
					p.ThresholdASet = true
					p.ThresholdA = st.low
					break
				}
			}
			for _, name := range []string{"canny-high", "edge-high"} {
				if f := cmd.Flags().Lookup(name); f != nil && f.Changed {
					p.ThresholdBSet = true
					p.ThresholdB = st.high
					break
				}
			}
		}
	}
}

func (opts ImageControlOptions) BuildUnits(cmd *cobra.Command) ([]driver.ControlNetUnit, error) {
	opts.finalizeParams(cmd)
	st := opts.state

	var units []driver.ControlNetUnit
	for _, spec := range imageControlSpecs {
		values := st.values[spec.canonical]
		if values == nil || len(*values) == 0 {
			continue
		}
		params := *st.params[spec.canonical]
		unit, err := driver.BuildControlUnit(spec.canonical, *values, params)
		if err != nil {
			return nil, err
		}
		units = append(units, unit)
	}
	return units, nil
}

func (opts ImageControlOptions) ControlMode() (driver.ControlNetMode, error) {
	if opts.state.mode == "" {
		return "", nil
	}
	return driver.ParseControlNetMode(opts.state.mode)
}
