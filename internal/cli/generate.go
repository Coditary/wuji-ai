package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-ai/internal/clix"
	"github.com/coditary/wuji-core/pkg/capability"
	"github.com/coditary/wuji-core/pkg/driver"
)

func newTextCmd(app *App) *cobra.Command {
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
	var useStdin bool
	var textFile, messagesFile, imagePath, videoPath, audioPath, documentPath string
	var translate bool
	var targetLang, language string
	var audioInference clix.AudioInferenceFields
	var textTrain textTrainOpts
	var loraEntries []string
	var loraWeight float32
	var textLoRAState loraFlagState
	var useStream bool
	cmd := &cobra.Command{
		Use:   "text [prompt]",
		Short: "Generate plain text from a prompt or media file",
		Long: `Produce plain text on stdout from text or media input.

Input (exactly one primary source):
  wuji text "hello"                  prompt text
  wuji text --text article.txt       read prompt from file
  cat article.txt | wuji text        read prompt from stdin (pipe)
  wuji text --stdin                  read prompt from stdin (explicit)
  wuji text --image photo.jpg        describe image (optional prompt for questions)
  wuji text --video clip.mp4         video understanding (driver must support it)
  wuji text --audio speech.wav       transcribe audio (video files extract audio first)
  wuji text --document scan.pdf      extract text from document

Modifiers:
  --translate                        translate text input (prompt, --text, or stdin)
  --target-lang de                   translation target language (default: English)
  --lang de                          source language for --audio or --translate

Audio transcription (--audio only):
  --quality fast|balanced|careful|1-5   speed vs. accuracy
  --trim-silence                        skip long silent passages
  --silence-level 0.0-1.0               how aggressively to trim silence

Media flags are mutually exclusive. Multi-step workflows are composed manually via the shell.`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if textTrain.Enabled {
				if err := requirePromptUnlessTrain(true, args); err != nil {
					return err
				}
				if textTrain.BaseModel == "" {
					textTrain.BaseModel = model
				}
				if textTrain.Seed == -1 {
					textTrain.Seed = seed
				}
				return runTextTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.TextGeneration), textTrain, clix.JoinArgs(args), model)
			}
			loras, err := resolveLoRAs(cmd, app, loraEntries, loraWeight, textLoRAState)
			if err != nil {
				return err
			}
			req, err := clix.BuildTextRequest(args, false, clix.TextRequestFields{
				UseStdin: useStdin,
				TextFile: textFile, MessagesFile: messagesFile, ImagePath: imagePath, VideoPath: videoPath,
				AudioPath: audioPath, DocumentPath: documentPath,
				Translate: translate, TargetLang: targetLang, Language: language,
				Quality: audioInference.Quality, TrimSilence: audioInference.TrimSilence,
				SilenceLevel: audioInference.SilenceLevel,
				Model:        model, SystemPrompt: systemPrompt, MaxTokens: maxTokens,
				Temperature: temperature, TopP: topP, TopK: topK, MinP: minP,
				FrequencyPenalty: frequencyPenalty, PresencePenalty: presencePenalty,
				RepetitionPenalty: repetitionPenalty, StopSequences: stopSequences,
				Seed: seed, ContextWindow: contextWindow, LoRAs: loras,
			})
			if err != nil {
				return err
			}
			driverID := app.resolveDriver(cmd, capability.TextGeneration)
			if useStream {
				_, err := app.Core.GenerateTextStream(context.Background(), driverID, req, func(delta string) error {
					_, werr := os.Stdout.WriteString(delta)
					return werr
				})
				if err != nil {
					return err
				}
				fmt.Fprintln(os.Stdout)
				return nil
			}
			resp, err := app.Core.GenerateText(context.Background(), driverID, req)
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Text)
			return nil
		},
	}
	cmd.Flags().BoolVar(&useStream, "stream", false, "stream tokens to stdout as they are generated")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "read prompt text from stdin (also auto-detected when input is piped)")
	cmd.Flags().StringVar(&textFile, "text", "", "read prompt text from file (alternative to positional prompt)")
	cmd.Flags().StringVar(&messagesFile, "messages", "", "read chat messages JSON from file (alternative to prompt; for multi-turn/agent use)")
	cmd.Flags().StringVar(&imagePath, "image", "", "image file for vision/OCR input")
	cmd.Flags().StringVar(&videoPath, "video", "", "video file for native video-to-text (no automatic fallback)")
	cmd.Flags().StringVar(&audioPath, "audio", "", "audio or video file to transcribe to plain text")
	cmd.Flags().StringVar(&documentPath, "document", "", "document file to extract plain text from")
	cmd.Flags().BoolVar(&translate, "translate", false, "translate text input to --target-lang (prompt, --text, or stdin)")
	cmd.Flags().StringVar(&targetLang, "target-lang", "", "translation target language (default: English)")
	cmd.Flags().StringVar(&language, "lang", "auto", "source language for --audio or --translate (auto, de, en, …)")
	clix.AddAudioInferenceFlags(cmd, &audioInference)
	cmd.Flags().IntVar(&maxTokens, "max-tokens", 1024, "maximum tokens to generate (-1 for unlimited)")
	cmd.Flags().Float32Var(&temperature, "temperature", 0.7, "sampling temperature (0.0–2.0)")
	cmd.Flags().StringVar(&model, "model", "", "model name (driver-specific)")
	cmd.Flags().Float32Var(&topP, "top-p", 0, "nucleus sampling (0.0–1.0, 0 = backend default)")
	cmd.Flags().IntVar(&topK, "top-k", 0, "top-k sampling (0 = disabled/backend default)")
	cmd.Flags().Float32Var(&minP, "min-p", 0, "min-p sampling filter (0.0–1.0, 0 = backend default)")
	cmd.Flags().Float32Var(&frequencyPenalty, "frequency-penalty", 0, "frequency penalty (-2.0–2.0, 0 = off)")
	cmd.Flags().Float32Var(&presencePenalty, "presence-penalty", 0, "presence penalty (-2.0–2.0, 0 = off)")
	cmd.Flags().Float32Var(&repetitionPenalty, "repetition-penalty", 0, "repetition penalty multiplier (1.0 = off, 0 = backend default)")
	cmd.Flags().StringSliceVar(&stopSequences, "stop", nil, "stop sequences (repeatable)")
	cmd.Flags().IntVar(&seed, "seed", -1, "random seed for reproducible output (-1 = random)")
	cmd.Flags().StringVar(&systemPrompt, "system-prompt", "", "system instructions prepended to the request")
	cmd.Flags().IntVar(&contextWindow, "context-window", 0, "context window size in tokens (0 = backend default)")
	cmd.Flags().StringSliceVar(&loraEntries, "lora", nil, "LoRA alias or path; bare --lora loads loras.default from config")
	cmd.Flags().Float32Var(&loraWeight, "lora-weight", 0, "default LoRA strength when weight is omitted (0 = use config default_weight or 1.0)")
	registerLoRAFlags(cmd, app, &textLoRAState)
	addTextTrainFlags(cmd, &textTrain, false)

	return cmd
}

func newImageCmd(app *App) *cobra.Command {
	var width, height, steps int
	var negativePrompt, imageModel, imageMode, sampler, initImage string
	var maskImage, controlImage, controlType, styleImage string
	var cfgScale, denoisingStrength, styleWeight float32
	var batchTotal, batchSize, batchCount, imageSeed int
	var frameWidth, frameHeight, columns, rows, frameCount int
	var scaleInput string
	var upscale, edit, sprite bool
	var referenceImage, spriteAction, spriteView string
	var spriteDirections, spritePadding int
	var spriteLoop, spriteTransparent bool
	var loras []string
	var loraWeight float32
	var imageLoRAState loraFlagState
	var imageTrain imageTrainOpts
	var imageControls clix.ImageControlOptions
	cmd := &cobra.Command{
		Use:   "image [prompt]",
		Short: "Generate or transform images",
		Long: `Run image generation and transformation tasks.

The workflow is inferred from your flags (see examples):

  wuji image "sunset" -W 768 -H 512 -s 25
      → generate (text-to-image)

  wuji image "watercolor" --image photo.png --denoising-strength 0.6
      → img2img (transform source image)

  wuji image "boat" --image photo.png --mask mask.png
      → inpaint (fill masked area)

  wuji image "portrait" --style ref.jpg
      → style-transfer (visual style from reference; not face identity)

  wuji image "person" --pose=full --pose pose.png
      → controlnet (structure/pose; --openpose is a synonym)

  wuji image "landscape" --depth photo.jpg
      → depth2img (--depth alone; synonyms: legacy --control-image)

  wuji image "portrait" --image photo.png --pose pose.png --denoising-strength 0.6
      → img2img + control units

  wuji image --image photo.png
      → variation (no prompt; ControlNet/batch/steps supported)

  wuji image --image photo.png --pose pose.png
      → variation + control units (no prompt)

  wuji image "make it winter" --edit --image photo.png
      → edit (falls back to img2img when driver has no edit backend)

  wuji image "orc walk cycle" --sprite --image hero.png --style pixel-ref.png --reference sheet.png --pose skeleton.png -W 64 -H 64 --columns 4 --rows 2 --action walk --view side --mode pixel
      → sprite (character + style + sheet reference + pose control)

  wuji image "grass tile" --mode tile
      → generate with tileset pipeline profile

Short flags: -W width  -H height  -s steps  (-h is help; use --help)

`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if imageTrain.Enabled {
				if err := requirePromptUnlessTrain(true, args); err != nil {
					return err
				}
				if imageTrain.BaseModel == "" {
					imageTrain.BaseModel = imageModel
				}
				if imageTrain.Mode == "" {
					imageTrain.Mode = imageMode
				}
				if imageTrain.Width <= 0 {
					imageTrain.Width = width
				}
				if imageTrain.Height <= 0 {
					imageTrain.Height = height
				}
				if imageTrain.Seed == -1 {
					imageTrain.Seed = imageSeed
				}
				if imageTrain.BatchSize <= 0 && batchSize > 0 {
					imageTrain.BatchSize = batchSize
				}
				if imageTrain.ControlType == "" && controlType != "" {
					imageTrain.ControlType = controlType
				}
				if imageTrain.ControlType == "" {
					units, err := imageControls.BuildUnits(cmd)
					if err != nil {
						return err
					}
					if t := clix.FirstControlType(units); t != "" {
						imageTrain.ControlType = string(t)
					}
				}
				return runImageTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.ImageGeneration), imageTrain, clix.JoinArgs(args), imageModel, cmd)
			}

			loraRefs, err := resolveLoRAs(cmd, app, loras, loraWeight, imageLoRAState)
			if err != nil {
				return err
			}
			upscaleRequested := upscale
			scale, upscaleRequested, err := clix.ParseImageScaleInput(scaleInput, upscaleRequested)
			if err != nil {
				return err
			}
			controlUnits, err := imageControls.BuildUnits(cmd)
			if err != nil {
				return err
			}
			controlMode, err := imageControls.ControlMode()
			if err != nil {
				return err
			}
			vramMB := app.Core.ResourcesConfig().TotalVRAMMB
			req, err := clix.BuildImageRequest(clix.JoinArgs(args), false, clix.ImageRequestFields{
				NegativePrompt:    negativePrompt,
				Model:             imageModel,
				Mode:              imageMode,
				Width:             width,
				Height:            height,
				Steps:             steps,
				Sampler:           sampler,
				CFGScale:          cfgScale,
				BatchTotal:        batchTotal,
				BatchSize:         batchSize,
				BatchCount:        batchCount,
				BatchSizeSet:      cmd.Flags().Changed("batch-size"),
				BatchCountSet:     cmd.Flags().Changed("batch-count"),
				AvailableVRAMMB:   vramMB,
				Seed:              imageSeed,
				DenoisingStrength: denoisingStrength,
				InitImage:         initImage,
				MaskImage:         maskImage,
				ControlImage:      controlImage,
				ControlType:       controlType,
				ControlUnits:      controlUnits,
				ControlMode:       controlMode,
				StyleImage:        styleImage,
				StyleWeight:       styleWeight,
				Scale:             scale,
				UpscaleRequested:  upscaleRequested,
				EditRequested:     edit,
				SpriteRequested:   sprite,
				ReferenceImage:    referenceImage,
				SpriteAction:      spriteAction,
				SpriteView:        spriteView,
				SpriteDirections:  spriteDirections,
				SpriteLoop:        spriteLoop,
				SpritePadding:     spritePadding,
				SpriteTransparent: spriteTransparent,
				FrameWidth:        frameWidth,
				FrameHeight:       frameHeight,
				Columns:           columns,
				Rows:              rows,
				FrameCount:        frameCount,
				LoRAs:             loraRefs,
			})
			if err != nil {
				return err
			}
			resp, err := app.Core.GenerateImage(context.Background(), app.resolveDriver(cmd, capability.ImageGeneration), req)
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
	cmd.Flags().IntVarP(&width, "width", "W", 512, "image width in pixels")
	cmd.Flags().IntVarP(&height, "height", "H", 512, "image height in pixels")
	cmd.Flags().IntVarP(&steps, "steps", "s", 20, "sampling steps (1–150)")
	cmd.Flags().StringVar(&negativePrompt, "negative-prompt", "", "what to avoid in the image")
	cmd.Flags().StringVar(&imageModel, "model", "", "base model (driver-specific, e.g. .safetensors name)")
	cmd.Flags().StringVar(&imageMode, "mode", "", "pipeline profile: prop, pixel, tile, ui, character, texture, …")
	cmd.Flags().StringVar(&sampler, "sampler", "", "sampler algorithm (e.g. euler_a, dpm++_2m_karras)")
	cmd.Flags().Float32Var(&cfgScale, "cfg-scale", 0, "classifier-free guidance (1.0–20.0, 0 = backend default)")
	cmd.Flags().IntVar(&batchTotal, "batch", 0, "target total image count; splits into --batch-size × --batch-count (heuristic when only this flag is set)")
	cmd.Flags().IntVar(&batchSize, "batch-size", 0, "images generated in parallel (0 = backend default; with --batch, the other dimension is computed)")
	cmd.Flags().IntVar(&batchCount, "batch-count", 1, "number of sequential batches (with --batch, the other dimension is computed)")
	cmd.Flags().IntVar(&imageSeed, "seed", -1, "random seed (-1 = random)")
	cmd.Flags().Float32Var(&denoisingStrength, "denoising-strength", 0, "img2img/edit strength (0.0–1.0, 0 = backend default)")
	cmd.Flags().BoolVar(&edit, "edit", false, "instruction edit of --image (uses img2img when driver has no edit backend)")
	cmd.Flags().BoolVar(&sprite, "sprite", false, "generate a sprite sheet atlas")
	cmd.Flags().BoolVar(&sprite, "spritesheet", false, "alias for --sprite")
	cmd.Flags().IntVar(&frameWidth, "frame-width", 0, "sprite frame width (defaults to -W with --sprite)")
	cmd.Flags().IntVar(&frameHeight, "frame-height", 0, "sprite frame height (defaults to -H with --sprite)")
	cmd.Flags().IntVar(&columns, "columns", 0, "sprite sheet columns")
	cmd.Flags().IntVar(&rows, "rows", 0, "sprite sheet rows")
	cmd.Flags().IntVar(&frameCount, "frames", 0, "total animation frames (alternative to --columns/--rows)")
	cmd.Flags().StringVar(&referenceImage, "reference", "", "reference sprite sheet for layout/style consistency")
	cmd.Flags().StringVar(&referenceImage, "reference-sheet", "", "alias for --reference")
	cmd.Flags().StringVar(&spriteAction, "action", "", "animation preset: idle, walk, run, attack, jump, cast, death")
	cmd.Flags().StringVar(&spriteView, "view", "", "camera preset: side, front, back, top, isometric, three-quarter")
	cmd.Flags().IntVar(&spriteDirections, "directions", 0, "directional frame count: 4 or 8")
	cmd.Flags().BoolVar(&spriteLoop, "loop", false, "generate a loopable animation")
	cmd.Flags().IntVar(&spritePadding, "padding", 0, "padding between frames in the atlas (pixels)")
	cmd.Flags().BoolVar(&spriteTransparent, "transparent", false, "use transparent background")
	cmd.Flags().StringVar(&initImage, "init-image", "", "source image (img2img, inpaint, upscale, edit, variation, sprite character reference)")
	cmd.Flags().StringVar(&initImage, "image", "", "alias for --init-image")
	cmd.Flags().StringVar(&maskImage, "mask", "", "inpaint mask image (white = replace)")
	cmd.Flags().StringVar(&controlImage, "control-image", "", "deprecated: use typed control flags (e.g. --depth, --pose)")
	cmd.Flags().StringVar(&controlType, "control-type", "", "deprecated: use typed control flags (e.g. --pose=full)")
	clix.RegisterImageControlFlags(cmd, &imageControls)
	cmd.Flags().StringVar(&styleImage, "style", "", "reference style image (style-transfer, IP-Adapter, or --sprite)")
	cmd.Flags().StringVar(&styleImage, "style-image", "", "alias for --style")
	cmd.Flags().Float32Var(&styleWeight, "style-weight", 0, "style/IP-Adapter strength (0 = backend default, typically 1.0)")
	cmd.Flags().BoolVar(&upscale, "upscale", false, "scale --image (default 4×; combine with --scale)")
	cmd.Flags().StringVar(&scaleInput, "scale", "", "scale factor: 2, 2.5, 50%, 300% (decimal '.', with --image)")
	cmd.Flags().StringSliceVar(&loras, "lora", nil, "LoRA alias or path; bare --lora loads loras.default from config")
	cmd.Flags().Float32Var(&loraWeight, "lora-weight", 0, "default LoRA strength when weight is omitted (0 = use config default_weight or 1.0)")
	registerLoRAFlags(cmd, app, &imageLoRAState)
	addImageTrainFlags(cmd, &imageTrain, false)
	return cmd
}

func newVideoCmd(app *App) *cobra.Command {
	// video
	var duration float32
	var videoFPS, videoFrames, contextLength, videoSeed int
	var scaleInput string
	var motionStrength float32
	var videoNegativePrompt, videoModel, videoSampler, videoScheduler, imagePath, videoPath, cameraControl string
	var videoTrain videoTrainOpts
	cmd := &cobra.Command{
		Use:   "video [prompt]",
		Short: "Generate or transform video",
		Long: `Generate or transform video files.

The workflow is inferred from your flags:

  wuji video "ocean waves at sunset"
      → generate (text-to-video)

  wuji video "gentle zoom" --image frame.png
      → i2v (animate a still image)

  wuji video --video clip.mp4 --fps 60
      → interpolate (raise frame rate to target fps)

  wuji video --video clip.mp4 --scale 4
      → scale (2, 2.5, 50%, 300%, …)

Generation tuning (generate / i2v; 0 or empty = backend default):
  --duration, --fps, --frames       length and timing
  --motion-strength                 how much movement in the scene
  --context-length                  temporal window for longer clips
  --sampler, --scheduler            diffusion sampling (like wuji image)
  --negative-prompt, --model, --seed

  --camera zoom_in|pan_left|…       camera motion preset (driver-dependent; not a prompt rewrite)

`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if videoTrain.Enabled {
				if err := requirePromptUnlessTrain(true, args); err != nil {
					return err
				}
				if videoTrain.BaseModel == "" {
					videoTrain.BaseModel = videoModel
				}
				if videoTrain.Seed == -1 {
					videoTrain.Seed = videoSeed
				}
				if videoTrain.ContextLength <= 0 && contextLength > 0 {
					videoTrain.ContextLength = contextLength
				}
				if videoTrain.FPS <= 0 && videoFPS > 0 {
					videoTrain.FPS = videoFPS
				}
				if videoTrain.Frames <= 0 && videoFrames > 0 {
					videoTrain.Frames = videoFrames
				}
				return runVideoTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.VideoGeneration), videoTrain, clix.JoinArgs(args), videoModel)
			}
			scaleRequested := cmd.Flags().Changed("scale")
			scale, scaleRequested, err := clix.ParseImageScaleInput(scaleInput, scaleRequested)
			if err != nil {
				return err
			}
			req, err := clix.BuildVideoRequest(clix.JoinArgs(args), false, clix.VideoRequestFields{
				Duration: duration, FPS: videoFPS, Frames: videoFrames, FPSSet: cmd.Flags().Changed("fps"),
				MotionStrength: motionStrength, ContextLength: contextLength, Sampler: videoSampler,
				Scheduler: videoScheduler,
				ImagePath: imagePath, VideoPath: videoPath, CameraControl: cameraControl,
				NegativePrompt: videoNegativePrompt, Model: videoModel, Seed: videoSeed,
				Scale: scale, ScaleSet: scaleRequested,
			})
			if err != nil {
				return err
			}
			resp, err := app.Core.GenerateVideo(context.Background(), app.resolveDriver(cmd, capability.VideoGeneration), req)
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
	cmd.Flags().StringVar(&imagePath, "image", "", "source image for image-to-video")
	cmd.Flags().StringVar(&videoPath, "video", "", "input video for interpolation or scaling")
	cmd.Flags().IntVar(&videoFrames, "frames", 0, "total number of frames (0 = derive from duration × fps)")
	cmd.Flags().IntVar(&videoFPS, "fps", 24, "playback frame rate; with --video and no --scale, sets interpolation target fps")
	cmd.Flags().Float32Var(&duration, "duration", 5, "target duration in seconds (used when --frames is 0)")
	cmd.Flags().Float32Var(&motionStrength, "motion-strength", 0, "how much movement in the scene (0 = backend default)")
	cmd.Flags().IntVar(&contextLength, "context-length", 0, "temporal window in frames for long clips (0 = backend default)")
	cmd.Flags().StringVar(&videoSampler, "sampler", "", "noise removal algorithm (e.g. euler, dpm++; 0 = backend default)")
	cmd.Flags().StringVar(&videoScheduler, "scheduler", "", "step schedule for sampling (e.g. karras; backend default if empty)")
	cmd.Flags().StringVar(&cameraControl, "camera", "", "camera motion preset: zoom_in, zoom_out, pan_left, pan_right, pan_up, pan_down")
	cmd.Flags().StringVar(&scaleInput, "scale", "", "scale factor with --video: 2, 2.5, 50%, 300% (decimal '.')")
	cmd.Flags().StringVar(&videoNegativePrompt, "negative-prompt", "", "what to avoid in the video")
	cmd.Flags().StringVar(&videoModel, "model", "", "video model (driver-specific)")
	cmd.Flags().IntVar(&videoSeed, "seed", -1, "random seed (-1 = random)")
	addVideoTrainFlags(cmd, &videoTrain, false)
	return cmd
}

func newAudioCmd(app *App) *cobra.Command {
	// audio
	var audioDuration, overlap, audioTemperature, audioCFGScale, audioTopP float32
	var audioTopK, audioSampleRate, audioSeed int
	var lyrics, audioNegativePrompt, audioModel, referencePath, audioFormat, voiceName, audioLang string
	var audioEmotion, audioStyle string
	var audioSpeed, audioEnergy float32
	var audioPitch int
	var sfx, speech bool
	var audioTrain audioTrainOpts
	cmd := &cobra.Command{
		Use:   "audio [prompt]",
		Short: "Generate audio from a prompt",
		Long: `Generate music or sound effects from a text prompt.

The workflow is inferred from your flags:

  wuji audio "lofi hip hop beat"
      → text-to-music

  wuji audio "epic orchestral theme" --reference melody.mid
      → melody-to-music

  wuji audio "thunder and rain" --sfx
      → text-to-sfx

  wuji audio "Guten Morgen" --voice myvoice
      → text-to-speech (cloned voice profile)

  wuji audio "Hello world" --speech
      → text-to-speech (generic voice)

Generation tuning (0 or empty = backend default):
  --duration, --overlap             total length; overlap stitches 10s chunks for longer music
  --format wav|mp3|flac|ogg|opus    output format (transcoded via ffmpeg when needed)
  --voice, --speech                 speak text (--voice uses a cloned profile)
  --lang, --speed, --pitch          speech language, rate, and pitch (text-to-speech only)
  --emotion, --style, --energy      speech expression (text-to-speech only)
  --lyrics                          song lyrics with tags like [Verse], [Chorus]
  --temperature, --cfg-scale        sampling and guidance (music/sfx)
  --top-p, --top-k                  token sampling (music/sfx)
  --negative-prompt, --model, --seed, --sample-rate

`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if audioTrain.Enabled {
				if err := requirePromptUnlessTrain(true, args); err != nil {
					return err
				}
				if audioTrain.BaseModel == "" {
					audioTrain.BaseModel = audioModel
				}
				if audioTrain.SampleRate <= 0 && audioSampleRate > 0 {
					audioTrain.SampleRate = audioSampleRate
				}
				if audioTrain.Seed == -1 {
					audioTrain.Seed = audioSeed
				}
				return runAudioTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.AudioGeneration), audioTrain, clix.JoinArgs(args), audioModel, sfx, speech, referencePath)
			}
			if len(args) < 1 {
				return fmt.Errorf("prompt required")
			}
			req, err := clix.BuildAudioRequest(clix.JoinArgs(args), clix.AudioRequestFields{
				Lyrics: lyrics, NegativePrompt: audioNegativePrompt, Model: audioModel, Duration: audioDuration, DurationSet: cmd.Flags().Changed("duration"),
				Overlap: overlap, Temperature: audioTemperature, CFGScale: audioCFGScale, TopP: audioTopP,
				TopK: audioTopK, SampleRate: audioSampleRate, ReferencePath: referencePath,
				Voice: voiceName, Language: audioLang, Speed: audioSpeed, Pitch: audioPitch,
				Emotion: audioEmotion, Style: audioStyle, Energy: audioEnergy,
				Format: audioFormat, SFXRequested: sfx, SpeechRequested: speech, Seed: audioSeed,
			})
			if err != nil {
				return err
			}
			resp, err := app.Core.GenerateAudio(context.Background(), app.resolveDriver(cmd, capability.AudioGeneration), req)
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
	cmd.Flags().Float32Var(&audioDuration, "duration", 10, "audio length in seconds")
	cmd.Flags().StringVar(&lyrics, "lyrics", "", "song lyrics with optional tags like [Verse], [Chorus]")
	cmd.Flags().StringVar(&audioNegativePrompt, "negative-prompt", "", "sounds to avoid (e.g. distortion, clipping)")
	cmd.Flags().StringVar(&audioModel, "model", "", "audio model (driver-specific)")
	cmd.Flags().Float32Var(&overlap, "overlap", 0, "crossfade overlap in seconds when --duration exceeds 10s (music tasks only)")
	cmd.Flags().StringVar(&audioFormat, "format", "", "output format: wav, mp3, flac, ogg, opus, m4a (default: driver native; transcode via ffmpeg)")
	cmd.Flags().Float32Var(&audioTemperature, "temperature", 0, "creativity (0.7–1.2 typical, 0 = backend default)")
	cmd.Flags().Float32Var(&audioCFGScale, "cfg-scale", 0, "guidance scale (3.0–15.0, 0 = backend default)")
	cmd.Flags().Float32Var(&audioTopP, "top-p", 0, "nucleus sampling on audio tokens (0 = backend default)")
	cmd.Flags().IntVar(&audioTopK, "top-k", 0, "top-k sampling on audio tokens (0 = disabled)")
	cmd.Flags().IntVar(&audioSampleRate, "sample-rate", 0, "output sample rate in Hz (e.g. 44100, 0 = backend default)")
	cmd.Flags().StringVar(&referencePath, "reference", "", "reference melody/audio/MIDI path (melody-to-music)")
	cmd.Flags().BoolVar(&sfx, "sfx", false, "generate sound effects instead of music")
	cmd.Flags().StringVar(&voiceName, "voice", "", "cloned voice profile name for text-to-speech")
	cmd.Flags().BoolVar(&speech, "speech", false, "speak text with a generic voice (text-to-speech)")
	cmd.Flags().StringVar(&audioLang, "lang", "auto", "speech language (auto, de, en, …; text-to-speech only)")
	cmd.Flags().Float32Var(&audioSpeed, "speed", 0, "speech rate multiplier (0.25–4.0, 0 = backend default)")
	cmd.Flags().IntVar(&audioPitch, "pitch", 0, "pitch shift in semitones (-12 to +12, text-to-speech only)")
	cmd.Flags().StringVar(&audioEmotion, "emotion", "", "speech emotion (e.g. happy, sad, angry, neutral; driver-dependent)")
	cmd.Flags().StringVar(&audioStyle, "style", "", "speech style (e.g. whisper, news, conversational; driver-dependent)")
	cmd.Flags().Float32Var(&audioEnergy, "energy", 0, "speech energy/intensity (0.0–1.0, 0 = backend default)")
	cmd.Flags().IntVar(&audioSeed, "seed", -1, "random seed (-1 = random)")
	addAudioTrainFlags(cmd, &audioTrain, false)
	return cmd
}

func newMeshCmd(app *App) *cobra.Command {
	var format, taskExplicit, meshModel, meshMode, representation, meshPath, highMeshPath, imagePath, videoPath, scaleInput string
	var depthPath, pointCloudPath, splatPath, maskPath, styleImagePath, textureImagePath, animationPath string
	var images []string
	var targetTris, meshSeed int
	var retopo, remesh, repair, smooth, refine, segment, rig, animate, retarget, uv, pbr bool
	var scene, variation, edit, upscale bool
	var meshTrain meshTrainOpts
	cmd := &cobra.Command{
		Use:     "mesh [prompt]",
		Aliases: []string{"3d"},
		Short:   "Generate or transform 3D mesh assets",
		Long: `Generate or transform mesh assets for games and 3D workflows.

The task is inferred from your flags (see driver mesh_tasks for support):

  wuji mesh "low poly treasure chest"
      → generate (text-to-mesh)

  wuji mesh --image concept.png
      → i2m (image-to-mesh)

  wuji mesh --images front.png side.png back.png
      → multiview

  wuji mesh --video scan.mp4
      → v2m

  wuji mesh --depth depth.png
      → depth2m

  wuji mesh --pointcloud cloud.ply
      → pcd2m

  wuji mesh --splat scene.splat
      → splat2m

  wuji mesh --mesh hero.glb -p "rusty metal"
      → texture

  wuji mesh --mesh hero.glb --target-tris 5000
      → decimate

  wuji mesh --mesh scan.glb --repair
      → repair

  wuji mesh "medieval room" --scene
      → scene (multi-object layout)

  wuji mesh --mesh hero.glb --variation
      → variation

  wuji mesh "add horns" --mesh hero.glb --edit
      → edit (semantic mesh change)

  wuji mesh --mesh hero.glb --upscale --scale 2
      → upscale (geometric detail)

  wuji mesh "knight" --mode human
      → generate with humanoid pipeline profile

  wuji mesh "forest scene" --representation splat
      → generate Gaussian splat intermediate (driver-dependent export)

Use --task to force a task when inference would be ambiguous.
`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if meshTrain.Enabled {
				if err := requirePromptUnlessTrain(true, args); err != nil {
					return err
				}
				if meshTrain.BaseModel == "" {
					meshTrain.BaseModel = meshModel
				}
				if meshTrain.Mode == "" {
					meshTrain.Mode = meshMode
				}
				if meshTrain.Seed == -1 {
					meshTrain.Seed = meshSeed
				}
				return runMeshTrain(cmd.Context(), app, app.resolveDriver(cmd, capability.Mesh), meshTrain, clix.JoinArgs(args), meshModel)
			}
			scaleRequested := cmd.Flags().Changed("scale")
			scale, scaleRequested, err := clix.ParseImageScaleInput(scaleInput, upscale || scaleRequested)
			if err != nil {
				return err
			}
			var representationVal driver.MeshRepresentation
			if cmd.Flags().Changed("representation") {
				representationVal, err = driver.ParseMeshRepresentation(representation)
				if err != nil {
					return err
				}
			}
			req, err := clix.BuildMeshRequest(clix.JoinArgs(args), false, clix.MeshRequestFields{
				TaskExplicit:       taskExplicit,
				Model:              meshModel,
				Format:             format,
				Mode:               meshMode,
				Seed:               meshSeed,
				TargetTris:         targetTris,
				TargetTrisSet:      cmd.Flags().Changed("target-tris"),
				Scale:              scale,
				ScaleSet:           scaleRequested,
				UpscaleRequested:   upscale,
				MeshPath:           meshPath,
				HighMeshPath:       highMeshPath,
				ImagePath:          imagePath,
				Images:             images,
				VideoPath:          videoPath,
				DepthPath:          depthPath,
				PointCloudPath:     pointCloudPath,
				SplatPath:          splatPath,
				MaskPath:           maskPath,
				StyleImagePath:     styleImagePath,
				TextureImagePath:   textureImagePath,
				AnimationPath:      animationPath,
				RetopoRequested:    retopo,
				RemeshRequested:    remesh,
				RepairRequested:    repair,
				SmoothRequested:    smooth,
				RefineRequested:    refine,
				SegmentRequested:   segment,
				RigRequested:       rig,
				AnimateRequested:   animate,
				RetargetRequested:  retarget,
				UVRequested:        uv,
				PBRRequested:       pbr,
				SceneRequested:     scene,
				VariationRequested: variation,
				EditRequested:      edit,
				Representation:     representationVal,
			})
			if err != nil {
				return err
			}
			resp, err := app.Core.GenerateMesh(context.Background(), app.resolveDriver(cmd, capability.Mesh), req)
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%s, task=%s, mode=%s, representation=%s)\n",
				resp.Path, resp.Format, resp.Task, resp.ModeOrDefault(), resp.RepresentationOrDefault())
			return nil
		},
	}
	cmd.Flags().StringVar(&taskExplicit, "task", "", "mesh task (generate, i2m, texture, decimate, …)")
	cmd.Flags().StringVar(&format, "format", "glb", "output format (glb, fbx, obj, …)")
	cmd.Flags().StringVar(&meshModel, "model", "", "mesh model (driver-specific)")
	cmd.Flags().StringVar(&meshMode, "mode", "", "pipeline profile: prop, human, hardsurface, terrain, environment, face")
	cmd.Flags().StringVar(&representation, "representation", "", "output representation: mesh, nerf, splat, pointcloud (generation tasks only)")
	cmd.Flags().IntVar(&meshSeed, "seed", -1, "random seed (-1 = random)")
	cmd.Flags().StringVar(&meshPath, "mesh", "", "input or output mesh file")
	cmd.Flags().StringVar(&highMeshPath, "high-mesh", "", "high-poly mesh for baking")
	cmd.Flags().StringVar(&imagePath, "image", "", "source image for image-to-mesh")
	cmd.Flags().StringSliceVar(&images, "images", nil, "multiple views for multiview reconstruction")
	cmd.Flags().StringVar(&videoPath, "video", "", "source video for video-to-mesh")
	cmd.Flags().StringVar(&depthPath, "depth", "", "depth or normal map for depth-to-mesh")
	cmd.Flags().StringVar(&pointCloudPath, "pointcloud", "", "point cloud input (.ply, …)")
	cmd.Flags().StringVar(&splatPath, "splat", "", "Gaussian splat input")
	cmd.Flags().StringVar(&maskPath, "mask", "", "region mask for mesh inpaint")
	cmd.Flags().StringVar(&styleImagePath, "style-image", "", "style reference image")
	cmd.Flags().StringVar(&textureImagePath, "texture-image", "", "reference image for img2tex")
	cmd.Flags().StringVar(&textureImagePath, "reference-image", "", "alias for --texture-image")
	cmd.Flags().StringVar(&animationPath, "animation", "", "animation clip for retargeting")
	cmd.Flags().IntVar(&targetTris, "target-tris", 0, "target triangle count for decimation")
	cmd.Flags().BoolVar(&retopo, "retopo", false, "retopologize mesh topology")
	cmd.Flags().BoolVar(&remesh, "remesh", false, "remesh with uniform topology")
	cmd.Flags().BoolVar(&repair, "repair", false, "repair mesh geometry")
	cmd.Flags().BoolVar(&smooth, "smooth", false, "smooth mesh geometry")
	cmd.Flags().BoolVar(&refine, "refine", false, "add geometric detail")
	cmd.Flags().BoolVar(&segment, "segment", false, "semantic part segmentation")
	cmd.Flags().BoolVar(&rig, "rig", false, "auto-rig mesh")
	cmd.Flags().BoolVar(&animate, "animate", false, "generate animation")
	cmd.Flags().BoolVar(&retarget, "retarget", false, "retarget animation onto mesh")
	cmd.Flags().BoolVar(&uv, "uv", false, "auto UV unwrap")
	cmd.Flags().BoolVar(&pbr, "pbr", false, "generate PBR material maps")
	cmd.Flags().BoolVar(&scene, "scene", false, "generate a multi-object scene from a prompt")
	cmd.Flags().BoolVar(&variation, "variation", false, "create a mesh variant")
	cmd.Flags().BoolVar(&edit, "edit", false, "edit mesh geometry from a prompt")
	cmd.Flags().BoolVar(&upscale, "upscale", false, "increase mesh geometric resolution (default scale 2×; combine with --scale)")
	cmd.Flags().StringVar(&scaleInput, "scale", "", "scale factor for --upscale: 2, 2.5, 200%")
	addMeshTrainFlags(cmd, &meshTrain, false)
	return cmd
}

func newGenerateCmd(app *App) *cobra.Command {
	cmd := &cobra.Command{
		Use:        "generate",
		Short:      "Deprecated alias for wuji text, image, …",
		Deprecated: "use top-level commands (wuji text, wuji image, wuji video, wuji audio, wuji mesh)",
	}
	cmd.AddCommand(newTextCmd(app), newImageCmd(app), newVideoCmd(app), newAudioCmd(app), newMeshCmd(app))
	return cmd
}

func newDatasetCmd(app *App) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "dataset",
		Short: "Manage datasets",
	}

	listCmd := &cobra.Command{
		Use:        "list",
		Short:      "List datasets",
		Deprecated: "use `wuji list datasets` instead",
		RunE: func(cmd *cobra.Command, args []string) error {
			return printDatasetList(app, cmd)
		},
	}

	createCmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path, _ := cmd.Flags().GetString("path")
			desc, _ := cmd.Flags().GetString("description")
			resp, err := app.Core.ManageDataset(context.Background(), app.resolveDriver(cmd, capability.DatasetMgmt), driver.DatasetRequest{
				Task: driver.DatasetTaskCreate, Name: args[0], Path: path, Description: desc,
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Message)
			return nil
		},
	}
	createCmd.Flags().String("path", "", "dataset path (required)")
	createCmd.Flags().String("description", "", "dataset description")
	_ = createCmd.MarkFlagRequired("path")

	deleteCmd := &cobra.Command{
		Use:   "delete [name]",
		Short: "Delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.ManageDataset(context.Background(), app.resolveDriver(cmd, capability.DatasetMgmt), driver.DatasetRequest{
				Task: driver.DatasetTaskDelete, Name: args[0],
			})
			if err != nil {
				return err
			}
			fmt.Fprintln(os.Stdout, resp.Message)
			return nil
		},
	}

	ingestCmd := &cobra.Command{
		Use:   "ingest [name]",
		Short: "Import files into a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			source, _ := cmd.Flags().GetString("source")
			format, _ := cmd.Flags().GetString("format")
			recursive, _ := cmd.Flags().GetBool("recursive")
			resp, err := app.Core.ManageDataset(context.Background(), app.resolveDriver(cmd, capability.DatasetMgmt), driver.DatasetRequest{
				Task: driver.DatasetTaskIngest, Name: args[0], SourcePath: source, Recursive: recursive, Format: format,
			})
			if err != nil {
				return err
			}
			fmt.Fprintf(os.Stdout, "%s (%d files, %d bytes", resp.Message, resp.FilesAdded, resp.BytesAdded)
			if resp.VersionID != "" {
				fmt.Fprintf(os.Stdout, ", version %s", resp.VersionID)
			}
			fmt.Fprintln(os.Stdout, ")")
			return nil
		},
	}
	ingestCmd.Flags().String("source", "", "file or directory to import (required)")
	ingestCmd.Flags().String("format", "", "optional format hint: jsonl, csv, parquet, files")
	ingestCmd.Flags().Bool("recursive", false, "recursively import directories")
	_ = ingestCmd.MarkFlagRequired("source")

	versionCmd := &cobra.Command{
		Use:   "version [name]",
		Short: "Create a dataset snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			tag, _ := cmd.Flags().GetString("tag")
			message, _ := cmd.Flags().GetString("message")
			resp, err := app.Core.ManageDataset(context.Background(), app.resolveDriver(cmd, capability.DatasetMgmt), driver.DatasetRequest{
				Task: driver.DatasetTaskVersion, Name: args[0], Tag: tag, Message: message,
			})
			if err != nil {
				return err
			}
			if len(resp.Versions) > 0 {
				v := resp.Versions[0]
				fmt.Fprintf(os.Stdout, "%s\n%s (%d bytes)\n", resp.Message, v.Path, v.Size)
				return nil
			}
			fmt.Fprintln(os.Stdout, resp.Message)
			return nil
		},
	}
	versionCmd.Flags().String("tag", "", "version tag (default: backend-generated)")
	versionCmd.Flags().String("message", "", "optional version note")

	versionsCmd := &cobra.Command{
		Use:   "versions [name]",
		Short: "List dataset versions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := app.Core.ManageDataset(context.Background(), app.resolveDriver(cmd, capability.DatasetMgmt), driver.DatasetRequest{
				Task: driver.DatasetTaskListVersions, Name: args[0],
			})
			if err != nil {
				return err
			}
			for _, v := range resp.Versions {
				fmt.Fprintf(os.Stdout, "%s\t%s\t%s\t%d bytes\t%d\n", v.ID, v.Tag, v.Path, v.Size, v.CreatedAtUnix)
			}
			return nil
		},
	}

	cmd.AddCommand(listCmd, createCmd, deleteCmd, ingestCmd, versionCmd, versionsCmd)
	return cmd
}
