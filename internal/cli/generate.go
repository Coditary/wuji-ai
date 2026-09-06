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
	var topP, minP float32
	var topK int
	var frequencyPenalty, presencePenalty, repetitionPenalty float32
	var stopSequences []string
	var seed int
	var systemPrompt string
	var contextWindow int
	var textTrain textTrainOpts
	textCmd := &cobra.Command{
		Use:   "text [prompt]",
		Short: "Generate text from a prompt",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePromptUnlessTrain(textTrain.Enabled, args); err != nil {
				return err
			}
			if textTrain.Enabled {
				return runTextTrain(cmd.Context(), app, driverID, textTrain, joinArgs(args), model)
			}
			req := driver.TextRequest{
				Prompt:            joinArgs(args),
				Model:             model,
				SystemPrompt:      systemPrompt,
				MaxTokens:         maxTokens,
				Temperature:       temperature,
				TopP:              topP,
				TopK:              topK,
				MinP:              minP,
				FrequencyPenalty:  frequencyPenalty,
				PresencePenalty:   presencePenalty,
				RepetitionPenalty: repetitionPenalty,
				StopSequences:     stopSequences,
				ContextWindow:     contextWindow,
			}
			if seed >= 0 {
				req.Seed = &seed
			}
			resp, err := app.Core.GenerateText(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Text)
			return nil
		},
	}
	textCmd.Flags().IntVar(&maxTokens, "max-tokens", 1024, "maximum tokens to generate (-1 for unlimited)")
	textCmd.Flags().Float32Var(&temperature, "temperature", 0.7, "sampling temperature (0.0–2.0)")
	textCmd.Flags().StringVar(&model, "model", "", "model name (driver-specific)")
	textCmd.Flags().Float32Var(&topP, "top-p", 0, "nucleus sampling (0.0–1.0, 0 = backend default)")
	textCmd.Flags().IntVar(&topK, "top-k", 0, "top-k sampling (0 = disabled/backend default)")
	textCmd.Flags().Float32Var(&minP, "min-p", 0, "min-p sampling filter (0.0–1.0, 0 = backend default)")
	textCmd.Flags().Float32Var(&frequencyPenalty, "frequency-penalty", 0, "frequency penalty (-2.0–2.0, 0 = off)")
	textCmd.Flags().Float32Var(&presencePenalty, "presence-penalty", 0, "presence penalty (-2.0–2.0, 0 = off)")
	textCmd.Flags().Float32Var(&repetitionPenalty, "repetition-penalty", 0, "repetition penalty multiplier (1.0 = off, 0 = backend default)")
	textCmd.Flags().StringSliceVar(&stopSequences, "stop", nil, "stop sequences (repeatable)")
	textCmd.Flags().IntVar(&seed, "seed", -1, "random seed for reproducible output (-1 = random)")
	textCmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "system instructions prepended to the request")
	textCmd.Flags().IntVar(&contextWindow, "context-window", 0, "context window size in tokens (0 = backend default)")
	addTextTrainFlags(textCmd, &textTrain)
	addDriverFlag(textCmd, &driverID)

	// image
	var width, height, steps int
	var negativePrompt, imageModel, sampler, initImage string
	var cfgScale, denoisingStrength float32
	var batchSize, batchCount, imageSeed int
	var loras []string
	var imageTrain imageTrainOpts
	imageCmd := &cobra.Command{
		Use:   "image [prompt]",
		Short: "Generate an image from a prompt",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePromptUnlessTrain(imageTrain.Enabled, args); err != nil {
				return err
			}
			if imageTrain.Enabled {
				return runImageTrain(cmd.Context(), app, driverID, imageTrain, joinArgs(args), imageModel)
			}
			req := driver.ImageRequest{
				Prompt:            joinArgs(args),
				NegativePrompt:    negativePrompt,
				Model:             imageModel,
				Width:             width,
				Height:            height,
				Steps:             steps,
				Sampler:           sampler,
				CFGScale:          cfgScale,
				BatchSize:         batchSize,
				BatchCount:        batchCount,
				DenoisingStrength: denoisingStrength,
				InitImagePath:     initImage,
				LoRAs:             loras,
			}
			if imageSeed >= 0 {
				req.Seed = &imageSeed
			}
			resp, err := app.Core.GenerateImage(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			if len(resp.Paths) > 0 {
				for _, p := range resp.Paths {
					fmt.Fprintf(os.Stdout, "%s (%s)\n", p, resp.Format)
				}
				return nil
			}
			fmt.Fprintf(os.Stdout, "%s (%s)\n", resp.Path, resp.Format)
			return nil
		},
	}
	imageCmd.Flags().IntVar(&width, "width", 512, "image width in pixels")
	imageCmd.Flags().IntVar(&height, "height", 512, "image height in pixels")
	imageCmd.Flags().IntVar(&steps, "steps", 20, "sampling steps (1–150)")
	imageCmd.Flags().StringVar(&negativePrompt, "negative-prompt", "", "what to avoid in the image")
	imageCmd.Flags().StringVar(&imageModel, "model", "", "base model (driver-specific, e.g. .safetensors name)")
	imageCmd.Flags().StringVar(&sampler, "sampler", "", "sampler algorithm (e.g. euler_a, dpm++_2m_karras)")
	imageCmd.Flags().Float32Var(&cfgScale, "cfg-scale", 0, "classifier-free guidance (1.0–20.0, 0 = backend default)")
	imageCmd.Flags().IntVar(&batchSize, "batch-size", 0, "images generated in parallel (0 = backend default)")
	imageCmd.Flags().IntVar(&batchCount, "batch-count", 1, "number of sequential batches")
	imageCmd.Flags().IntVar(&imageSeed, "seed", -1, "random seed (-1 = random)")
	imageCmd.Flags().Float32Var(&denoisingStrength, "denoising-strength", 0, "img2img denoising strength (0.0–1.0, 0 = txt2img)")
	imageCmd.Flags().StringVar(&initImage, "init-image", "", "source image path for img2img or upscaling")
	imageCmd.Flags().StringSliceVar(&loras, "lora", nil, "LoRA model paths or names (repeatable)")
	addImageTrainFlags(imageCmd, &imageTrain)
	addDriverFlag(imageCmd, &driverID)

	// video
	var duration float32
	var videoFPS, videoFrames, contextLength, videoSeed int
	var motionStrength, interpolationStrength float32
	var videoNegativePrompt, videoModel, videoSampler, videoScheduler, videoMode, initVideoImage, cameraControl string
	var interpolate bool
	var videoTrain videoTrainOpts
	videoCmd := &cobra.Command{
		Use:   "video [prompt]",
		Short: "Generate a video from a prompt",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePromptUnlessTrain(videoTrain.Enabled, args); err != nil {
				return err
			}
			if videoTrain.Enabled {
				return runVideoTrain(cmd.Context(), app, driverID, videoTrain, joinArgs(args), videoModel)
			}
			req := driver.VideoRequest{
				Prompt:                joinArgs(args),
				Duration:              duration,
				FPS:                   videoFPS,
				Frames:                videoFrames,
				MotionStrength:        motionStrength,
				ContextLength:         contextLength,
				Sampler:               videoSampler,
				Scheduler:             videoScheduler,
				Interpolate:           interpolate,
				InterpolationStrength: interpolationStrength,
				Mode:                  driver.VideoMode(videoMode),
				InitImagePath:         initVideoImage,
				CameraControl:         driver.CameraControl(cameraControl),
				NegativePrompt:        videoNegativePrompt,
				Model:                 videoModel,
			}
			if videoSeed >= 0 {
				req.Seed = &videoSeed
			}
			resp, err := app.Core.GenerateVideo(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			if resp.Frames > 0 && resp.FPS > 0 {
				fmt.Fprintf(os.Stdout, "%s (%.1fs, %d frames @ %d fps)\n", resp.Path, resp.Duration, resp.Frames, resp.FPS)
				return nil
			}
			fmt.Fprintf(os.Stdout, "%s (%.1fs)\n", resp.Path, resp.Duration)
			return nil
		},
	}
	videoCmd.Flags().IntVar(&videoFrames, "frames", 0, "total number of frames (0 = derive from duration × fps)")
	videoCmd.Flags().IntVar(&videoFPS, "fps", 24, "playback frame rate (e.g. 8, 16, 24)")
	videoCmd.Flags().Float32Var(&duration, "duration", 5, "target duration in seconds (used when --frames is 0)")
	videoCmd.Flags().Float32Var(&motionStrength, "motion-strength", 0, "motion amount / motion bucket (0 = backend default)")
	videoCmd.Flags().IntVar(&contextLength, "context-length", 0, "AnimateDiff context window in frames (0 = backend default)")
	videoCmd.Flags().StringVar(&videoSampler, "sampler", "", "video sampler (e.g. euler, lcm)")
	videoCmd.Flags().StringVar(&videoScheduler, "scheduler", "", "scheduler for temporal consistency")
	videoCmd.Flags().BoolVar(&interpolate, "interpolate", false, "enable frame interpolation (RIFE etc.)")
	videoCmd.Flags().Float32Var(&interpolationStrength, "interpolation-strength", 0, "interpolation blend strength (0 = backend default)")
	videoCmd.Flags().StringVar(&videoMode, "mode", "t2v", "generation mode: t2v (text-to-video) or i2v (image-to-video)")
	videoCmd.Flags().StringVar(&initVideoImage, "init-image", "", "first frame image path (required for i2v)")
	videoCmd.Flags().StringVar(&cameraControl, "camera", "", "camera preset: zoom_in, zoom_out, pan_left, pan_right, pan_up, pan_down")
	videoCmd.Flags().StringVar(&videoNegativePrompt, "negative-prompt", "", "what to avoid in the video")
	videoCmd.Flags().StringVar(&videoModel, "model", "", "video model (driver-specific)")
	videoCmd.Flags().IntVar(&videoSeed, "seed", -1, "random seed (-1 = random)")
	addVideoTrainFlags(videoCmd, &videoTrain)
	addDriverFlag(videoCmd, &driverID)

	// audio
	var audioDuration, overlap, audioTemperature, audioCFGScale, audioTopP float32
	var audioTopK, audioSampleRate, audioSeed int
	var lyrics, audioNegativePrompt, audioModel, taskType, referencePath string
	var audioTrain audioTrainOpts
	audioCmd := &cobra.Command{
		Use:   "audio [prompt]",
		Short: "Generate audio from a prompt",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePromptUnlessTrain(audioTrain.Enabled, args); err != nil {
				return err
			}
			if audioTrain.Enabled {
				return runAudioTrain(cmd.Context(), app, driverID, audioTrain, joinArgs(args), audioModel)
			}
			req := driver.AudioRequest{
				Prompt:         joinArgs(args),
				Lyrics:         lyrics,
				NegativePrompt: audioNegativePrompt,
				Model:          audioModel,
				Duration:       audioDuration,
				Overlap:        overlap,
				Temperature:    audioTemperature,
				CFGScale:       audioCFGScale,
				TopP:           audioTopP,
				TopK:           audioTopK,
				SampleRate:     audioSampleRate,
				TaskType:       driver.AudioTaskType(taskType),
				ReferencePath:  referencePath,
			}
			if audioSeed >= 0 {
				req.Seed = &audioSeed
			}
			resp, err := app.Core.GenerateAudio(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			if resp.SampleRate > 0 {
				fmt.Fprintf(os.Stdout, "%s (%.1fs, %d Hz, %s)\n", resp.Path, resp.Duration, resp.SampleRate, resp.Format)
				return nil
			}
			fmt.Fprintf(os.Stdout, "%s (%.1fs)\n", resp.Path, resp.Duration)
			return nil
		},
	}
	audioCmd.Flags().Float32Var(&audioDuration, "duration", 10, "audio length in seconds")
	audioCmd.Flags().StringVar(&lyrics, "lyrics", "", "song lyrics with optional tags like [Verse], [Chorus]")
	audioCmd.Flags().StringVar(&audioNegativePrompt, "negative-prompt", "", "sounds to avoid (e.g. distortion, clipping)")
	audioCmd.Flags().StringVar(&audioModel, "model", "", "audio model (driver-specific)")
	audioCmd.Flags().Float32Var(&overlap, "overlap", 0, "continuation overlap in seconds for longer tracks (0 = backend default)")
	audioCmd.Flags().Float32Var(&audioTemperature, "temperature", 0, "creativity (0.7–1.2 typical, 0 = backend default)")
	audioCmd.Flags().Float32Var(&audioCFGScale, "cfg-scale", 0, "guidance scale (3.0–15.0, 0 = backend default)")
	audioCmd.Flags().Float32Var(&audioTopP, "top-p", 0, "nucleus sampling on audio tokens (0 = backend default)")
	audioCmd.Flags().IntVar(&audioTopK, "top-k", 0, "top-k sampling on audio tokens (0 = disabled)")
	audioCmd.Flags().IntVar(&audioSampleRate, "sample-rate", 0, "output sample rate in Hz (e.g. 44100, 0 = backend default)")
	audioCmd.Flags().StringVar(&taskType, "task", "text-to-music", "task type: text-to-music, melody-to-music, text-to-sfx, tts")
	audioCmd.Flags().StringVar(&referencePath, "reference", "", "reference melody/audio/MIDI path (for melody-to-music)")
	audioCmd.Flags().IntVar(&audioSeed, "seed", -1, "random seed (-1 = random)")
	addAudioTrainFlags(audioCmd, &audioTrain)
	addDriverFlag(audioCmd, &driverID)

	// 3d
	var format string
	var asset3dTrain asset3DTrainOpts
	asset3dCmd := &cobra.Command{
		Use:   "3d [prompt]",
		Short: "Generate a 3D asset from a prompt",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePromptUnlessTrain(asset3dTrain.Enabled, args); err != nil {
				return err
			}
			if asset3dTrain.Enabled {
				return run3DTrain(cmd.Context(), app, driverID, asset3dTrain, joinArgs(args), "")
			}
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
	add3DTrainFlags(asset3dCmd, &asset3dTrain)
	addDriverFlag(asset3dCmd, &driverID)

	cmd.AddCommand(textCmd, imageCmd, videoCmd, audioCmd, asset3dCmd)
	return cmd
}

func newTTSCmd(app *App) *cobra.Command {
	var driverID, voice, voiceSample, language, emotion, ttsModel string
	var speed, temperature, repetitionPenalty float32
	var ttsSeed int
	var ttsTrain ttsTrainOpts

	cmd := &cobra.Command{
		Use:   "tts [text]",
		Short: "Convert text to speech",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requirePromptUnlessTrain(ttsTrain.Enabled, args); err != nil {
				return err
			}
			if ttsTrain.Enabled {
				return runTTSTrain(cmd.Context(), app, driverID, ttsTrain, joinArgs(args), ttsModel)
			}
			req := driver.TTSRequest{
				Text:              joinArgs(args),
				Voice:             voice,
				VoiceSamplePath:   voiceSample,
				Language:          language,
				Speed:             speed,
				Emotion:           emotion,
				Temperature:       temperature,
				RepetitionPenalty: repetitionPenalty,
				Model:             ttsModel,
			}
			if ttsSeed >= 0 {
				req.Seed = &ttsSeed
			}
			resp, err := app.Core.Synthesize(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			if resp.SampleRate > 0 {
				fmt.Fprintf(os.Stdout, "%s (%.1fs, %d Hz, %s)\n", resp.Path, resp.Duration, resp.SampleRate, resp.Format)
				return nil
			}
			fmt.Fprintf(os.Stdout, "%s (%.1fs)\n", resp.Path, resp.Duration)
			return nil
		},
	}
	cmd.Flags().StringVar(&voice, "voice", "default", "speaker ID or preset voice name")
	cmd.Flags().StringVar(&voiceSample, "voice-sample", "", "path to .wav sample for voice cloning (3–10s)")
	cmd.Flags().StringVar(&language, "lang", "", "target language (e.g. de, en, es; empty = model default)")
	cmd.Flags().Float32Var(&speed, "speed", 0, "speaking rate (0.5–2.0, 0 = 1.0)")
	cmd.Flags().StringVar(&emotion, "emotion", "", "emotion or tone tags (e.g. happy, [sad])")
	cmd.Flags().Float32Var(&temperature, "temperature", 0, "token creativity (0 = backend default)")
	cmd.Flags().Float32Var(&repetitionPenalty, "repetition-penalty", 0, "anti-stutter penalty (0 = backend default)")
	cmd.Flags().StringVar(&ttsModel, "model", "", "TTS model (driver-specific)")
	cmd.Flags().IntVar(&ttsSeed, "seed", -1, "random seed (-1 = random)")
	addTTSTrainFlags(cmd, &ttsTrain)
	addDriverFlag(cmd, &driverID)
	return cmd
}

func newSTTCmd(app *App) *cobra.Command {
	var driverID, language, sttModel, task string
	var beamSize int
	var wordTimestamps, vadEnabled bool
	var vadThreshold float32
	var sttTrain sttTrainOpts

	cmd := &cobra.Command{
		Use:   "stt [audio-file]",
		Short: "Transcribe speech to text",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sttTrain.Enabled {
				name := joinArgs(args)
				if name == "" {
					name = sttModel
				}
				return runSTTTrain(cmd.Context(), app, driverID, sttTrain, name, sttModel)
			}
			if len(args) < 1 {
				return fmt.Errorf("audio file required unless --train is set")
			}
			req := driver.STTRequest{
				AudioPath:      args[0],
				Language:       language,
				Model:          sttModel,
				Task:           driver.STTTask(task),
				BeamSize:       beamSize,
				WordTimestamps: wordTimestamps,
				VADEnabled:     vadEnabled,
				VADThreshold:   vadThreshold,
			}
			resp, err := app.Core.Transcribe(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (confidence: %.0f%%)\n", resp.Text, resp.Confidence*100)
			for _, w := range resp.Words {
				fmt.Fprintf(os.Stdout, "  %.2f–%.2fs: %s\n", w.Start, w.End, w.Word)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&language, "lang", "auto", "source language (auto, de, en, …)")
	cmd.Flags().StringVar(&sttModel, "model", "", "model size (tiny, base, small, medium, large-v3)")
	cmd.Flags().StringVar(&task, "task", "transcribe", "task: transcribe or translate")
	cmd.Flags().IntVar(&beamSize, "beam-size", 0, "beam search width (1–5, 0 = backend default)")
	cmd.Flags().BoolVar(&wordTimestamps, "word-timestamps", false, "include per-word timestamps")
	cmd.Flags().BoolVar(&vadEnabled, "vad", false, "enable voice activity detection")
	cmd.Flags().Float32Var(&vadThreshold, "vad-threshold", 0, "VAD/silence threshold (0 = backend default)")
	addSTTTrainFlags(cmd, &sttTrain)
	addDriverFlag(cmd, &driverID)
	return cmd
}

func newVoiceCmd(app *App) *cobra.Command {
	var driverID string

	cmd := &cobra.Command{
		Use:   "voice",
		Short: "Voice cloning operations",
	}

	var (
		sample, targetModel, sourcePath, mode, vocoder string
		pitchShift, chunkSize int
		indexRate, protect, denoiseStrength, crossfade float32
		denoise bool
		voiceTrain voiceTrainOpts
	)

	cloneCmd := &cobra.Command{
		Use:   "clone [name]",
		Short: "Clone a voice from an audio sample",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 && !voiceTrain.Enabled {
				return fmt.Errorf("voice name required unless --train is set")
			}
			name := ""
			if len(args) >= 1 {
				name = args[0]
			}
			if voiceTrain.Enabled {
				return runVoiceTrain(cmd.Context(), app, driverID, voiceTrain, name, targetModel)
			}
			req := driver.VoiceRequest{
				Name:            name,
				SamplePath:      sample,
				TargetModel:     targetModel,
				SourcePath:      sourcePath,
				Mode:            driver.VoiceCloneMode(mode),
				PitchShift:      pitchShift,
				IndexRate:       indexRate,
				Protect:         protect,
				Vocoder:         vocoder,
				Denoise:         denoise,
				DenoiseStrength: denoiseStrength,
				ChunkSize:       chunkSize,
				Crossfade:       crossfade,
			}
			resp, err := app.Core.CloneVoice(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "voice %q created (id: %s)\n", resp.Name, resp.VoiceID)
			if resp.OutputPath != "" {
				fmt.Fprintf(os.Stdout, "output: %s\n", resp.OutputPath)
			}
			return nil
		},
	}
	cloneCmd.Flags().StringVar(&sample, "sample", "", "reference voice audio (.wav/.mp3, required)")
	_ = cloneCmd.MarkFlagRequired("sample")
	cloneCmd.Flags().StringVar(&targetModel, "target-model", "", "pre-trained voice model path or ID (RVC .pth/.index)")
	cloneCmd.Flags().StringVar(&sourcePath, "source", "", "source audio for voice conversion mode")
	cloneCmd.Flags().StringVar(&mode, "mode", "zero-shot", "workflow: zero-shot or conversion")
	cloneCmd.Flags().IntVar(&pitchShift, "pitch-shift", 0, "pitch shift in semitones (-12 to +12)")
	cloneCmd.Flags().Float32Var(&indexRate, "index-rate", 0, "feature retrieval rate (0.0–1.0, 0 = backend default)")
	cloneCmd.Flags().Float32Var(&protect, "protect", 0, "protect voiceless consonants (0.0–0.5, 0 = backend default)")
	cloneCmd.Flags().StringVar(&vocoder, "vocoder", "", "vocoder (hifigan, vocos, nsf_hifigan, …)")
	cloneCmd.Flags().BoolVar(&denoise, "denoise", false, "enable reference noise reduction")
	cloneCmd.Flags().Float32Var(&denoiseStrength, "denoise-strength", 0, "denoising strength (0 = backend default)")
	cloneCmd.Flags().IntVar(&chunkSize, "chunk-size", 0, "audio chunk size for long inputs (0 = backend default)")
	cloneCmd.Flags().Float32Var(&crossfade, "crossfade", 0, "chunk crossfade in seconds (0 = backend default)")
	addVoiceTrainFlags(cloneCmd, &voiceTrain)
	addDriverFlag(cloneCmd, &driverID)

	cmd.AddCommand(cloneCmd)
	return cmd
}

func newTrainCmd(app *App) *cobra.Command {
	var driverID string

	root := &cobra.Command{
		Use:   "train",
		Short: "Train models per capability",
	}

	addTrainSubcommand := func(use, short string, run func(*cobra.Command, []string) error) *cobra.Command {
		cmd := &cobra.Command{Use: use, Short: short, Args: cobra.ArbitraryArgs, RunE: run}
		addDriverFlag(cmd, &driverID)
		return cmd
	}

	var textOpts textTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("text [name]", "Train a text/LLM model", func(cmd *cobra.Command, args []string) error {
			textOpts.Enabled = true
			return runTextTrain(cmd.Context(), app, driverID, textOpts, joinArgs(args), textOpts.BaseModel)
		})
		addTextTrainFlags(cmd, &textOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var imageOpts imageTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("image [name]", "Train an image/diffusion model", func(cmd *cobra.Command, args []string) error {
			imageOpts.Enabled = true
			return runImageTrain(cmd.Context(), app, driverID, imageOpts, joinArgs(args), imageOpts.BaseModel)
		})
		addImageTrainFlags(cmd, &imageOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var videoOpts videoTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("video [name]", "Train a video model", func(cmd *cobra.Command, args []string) error {
			videoOpts.Enabled = true
			return runVideoTrain(cmd.Context(), app, driverID, videoOpts, joinArgs(args), videoOpts.BaseModel)
		})
		addVideoTrainFlags(cmd, &videoOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var audioOpts audioTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("audio [name]", "Train an audio model", func(cmd *cobra.Command, args []string) error {
			audioOpts.Enabled = true
			return runAudioTrain(cmd.Context(), app, driverID, audioOpts, joinArgs(args), audioOpts.BaseModel)
		})
		addAudioTrainFlags(cmd, &audioOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var asset3dOpts asset3DTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("3d [name]", "Train a 3D asset model", func(cmd *cobra.Command, args []string) error {
			asset3dOpts.Enabled = true
			return run3DTrain(cmd.Context(), app, driverID, asset3dOpts, joinArgs(args), asset3dOpts.BaseModel)
		})
		add3DTrainFlags(cmd, &asset3dOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var ttsOpts ttsTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("tts [name]", "Train a TTS model", func(cmd *cobra.Command, args []string) error {
			ttsOpts.Enabled = true
			return runTTSTrain(cmd.Context(), app, driverID, ttsOpts, joinArgs(args), ttsOpts.BaseModel)
		})
		addTTSTrainFlags(cmd, &ttsOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var sttOpts sttTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("stt [name]", "Train an STT model", func(cmd *cobra.Command, args []string) error {
			sttOpts.Enabled = true
			return runSTTTrain(cmd.Context(), app, driverID, sttOpts, joinArgs(args), sttOpts.BaseModel)
		})
		addSTTTrainFlags(cmd, &sttOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	var voiceOpts voiceTrainOpts
	root.AddCommand(func() *cobra.Command {
		cmd := addTrainSubcommand("voice [name]", "Train a voice cloning model", func(cmd *cobra.Command, args []string) error {
			voiceOpts.Enabled = true
			return runVoiceTrain(cmd.Context(), app, driverID, voiceOpts, joinArgs(args), voiceOpts.PretrainedModel)
		})
		addVoiceTrainFlags(cmd, &voiceOpts)
		_ = cmd.MarkFlagRequired("dataset")
		return cmd
	}())

	return root
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
