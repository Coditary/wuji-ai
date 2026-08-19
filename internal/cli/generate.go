package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji/internal/driver"
)

func joinArgs(args []string) string {
	return strings.Join(args, " ")
}

func addDriverFlag(cmd *cobra.Command, driverID *string) {
	cmd.Flags().StringVarP(driverID, "driver", "d", "", "driver to use (default: configured default)")
}

func newGenerateCmd(app *App) *cobra.Command {
	var driverID string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate content using AI backends",
	}

	// text
	var model string
	var maxTokens int
	var temperature float32
	textCmd := &cobra.Command{
		Use:   "text [prompt]",
		Short: "Generate text from a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.GenerateText(context.Background(), driverID, driver.TextRequest{
				Prompt: joinArgs(args), Model: model, MaxTokens: maxTokens, Temperature: temperature,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Text)
			return nil
		},
	}
	textCmd.Flags().IntVar(&maxTokens, "max-tokens", 1024, "maximum tokens to generate (-1 for unlimited)")
	textCmd.Flags().Float32Var(&temperature, "temperature", 0.7, "sampling temperature")
	textCmd.Flags().StringVar(&model, "model", "", "model name (driver-specific)")
	addDriverFlag(textCmd, &driverID)

	// image
	var width, height, steps int
	imageCmd := &cobra.Command{
		Use:   "image [prompt]",
		Short: "Generate an image from a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.GenerateImage(context.Background(), driverID, driver.ImageRequest{
				Prompt: joinArgs(args), Width: width, Height: height, Steps: steps,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%s)\n", resp.Path, resp.Format)
			return nil
		},
	}
	imageCmd.Flags().IntVar(&width, "width", 512, "image width")
	imageCmd.Flags().IntVar(&height, "height", 512, "image height")
	imageCmd.Flags().IntVar(&steps, "steps", 20, "diffusion steps")
	addDriverFlag(imageCmd, &driverID)

	// video
	var duration float32
	var fps int
	videoCmd := &cobra.Command{
		Use:   "video [prompt]",
		Short: "Generate a video from a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.GenerateVideo(context.Background(), driverID, driver.VideoRequest{
				Prompt: joinArgs(args), Duration: duration, FPS: fps,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%.1fs)\n", resp.Path, resp.Duration)
			return nil
		},
	}
	videoCmd.Flags().Float32Var(&duration, "duration", 5, "video duration in seconds")
	videoCmd.Flags().IntVar(&fps, "fps", 24, "frames per second")
	addDriverFlag(videoCmd, &driverID)

	// audio
	audioCmd := &cobra.Command{
		Use:   "audio [prompt]",
		Short: "Generate audio from a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.GenerateAudio(context.Background(), driverID, driver.AudioRequest{
				Prompt: joinArgs(args), Duration: duration,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%.1fs)\n", resp.Path, resp.Duration)
			return nil
		},
	}
	audioCmd.Flags().Float32Var(&duration, "duration", 10, "audio duration in seconds")
	addDriverFlag(audioCmd, &driverID)

	// 3d
	var format string
	asset3dCmd := &cobra.Command{
		Use:   "3d [prompt]",
		Short: "Generate a 3D asset from a prompt",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.Generate3D(context.Background(), driverID, driver.Asset3DRequest{
				Prompt: joinArgs(args), Format: format,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%s)\n", resp.Path, resp.Format)
			return nil
		},
	}
	asset3dCmd.Flags().StringVar(&format, "format", "glb", "output format")
	addDriverFlag(asset3dCmd, &driverID)

	cmd.AddCommand(textCmd, imageCmd, videoCmd, audioCmd, asset3dCmd)
	return cmd
}

func newTTSCmd(app *App) *cobra.Command {
	var driverID, voice string

	cmd := &cobra.Command{
		Use:   "tts [text]",
		Short: "Convert text to speech",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.Synthesize(context.Background(), driverID, driver.TTSRequest{
				Text: joinArgs(args), Voice: voice,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%.1fs)\n", resp.Path, resp.Duration)
			return nil
		},
	}
	cmd.Flags().StringVar(&voice, "voice", "default", "voice to use")
	addDriverFlag(cmd, &driverID)
	return cmd
}

func newSTTCmd(app *App) *cobra.Command {
	var driverID, language string

	cmd := &cobra.Command{
		Use:   "stt [audio-file]",
		Short: "Transcribe speech to text",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.Transcribe(context.Background(), driverID, driver.STTRequest{
				AudioPath: args[0], Language: language,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (confidence: %.0f%%)\n", resp.Text, resp.Confidence*100)
			return nil
		},
	}
	cmd.Flags().StringVar(&language, "lang", "auto", "language code")
	addDriverFlag(cmd, &driverID)
	return cmd
}

func newVoiceCmd(app *App) *cobra.Command {
	var driverID string

	cmd := &cobra.Command{
		Use:   "voice",
		Short: "Voice cloning operations",
	}

	cloneCmd := &cobra.Command{
		Use:   "clone [name]",
		Short: "Clone a voice from an audio sample",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			sample, _ := cmd.Flags().GetString("sample")
			resp, err := app.Core.CloneVoice(context.Background(), driverID, driver.VoiceRequest{
				Name: args[0], SamplePath: sample,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "voice %q created (id: %s)\n", resp.Name, resp.VoiceID)
			return nil
		},
	}
	cloneCmd.Flags().String("sample", "", "path to voice sample audio (required)")
	_ = cloneCmd.MarkFlagRequired("sample")
	addDriverFlag(cloneCmd, &driverID)

	cmd.AddCommand(cloneCmd)
	return cmd
}

func newTrainCmd(app *App) *cobra.Command {
	var driverID, datasetID, modelType string
	var epochs int

	cmd := &cobra.Command{
		Use:   "train",
		Short: "Train a model",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.Train(context.Background(), driverID, driver.TrainRequest{
				DatasetID: datasetID, ModelType: modelType, Epochs: epochs,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "job %s: %s\n", resp.JobID, resp.Status)
			return nil
		},
	}
	cmd.Flags().StringVar(&datasetID, "dataset", "", "dataset ID (required)")
	cmd.Flags().StringVar(&modelType, "model", "default", "model type")
	cmd.Flags().IntVar(&epochs, "epochs", 10, "training epochs")
	_ = cmd.MarkFlagRequired("dataset")
	addDriverFlag(cmd, &driverID)
	return cmd
}

func newDatasetCmd(app *App) *cobra.Command {
	var driverID string

	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Manage datasets",
	}

	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List datasets",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.ManageDataset(context.Background(), driverID, driver.DatasetRequest{
				Action: driver.DatasetList,
			})
			if err != nil {
				return err
			}
			for _, ds := range resp.Datasets {
				fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%d bytes\n", ds.ID, ds.Name, ds.Path, ds.Size)
			}
			return nil
		},
	}
	addDriverFlag(listCmd, &driverID)

	createCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("path")
			resp, err := app.Core.ManageDataset(context.Background(), driverID, driver.DatasetRequest{
				Action: driver.DatasetCreate, Name: args[0], Path: path,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Message)
			return nil
		},
	}
	createCmd.Flags().String("path", "", "dataset path (required)")
	_ = createCmd.MarkFlagRequired("path")
	addDriverFlag(createCmd, &driverID)

	deleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.ManageDataset(context.Background(), driverID, driver.DatasetRequest{
				Action: driver.DatasetDelete, Name: args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Message)
			return nil
		},
	}
	addDriverFlag(deleteCmd, &driverID)

	cmd.AddCommand(listCmd, createCmd, deleteCmd)
	return cmd
}
