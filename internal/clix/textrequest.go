package clix

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/coditary/wuji-core/pkg/driver"
)

type TextRequestFields struct {
	UseStdin          bool
	StdinReader       io.Reader // tests only; nil uses os.Stdin
	TextFile          string
	ImagePath         string
	VideoPath         string
	AudioPath         string
	DocumentPath      string
	Translate         bool
	TargetLang        string
	Language          string
	Quality           string
	TrimSilence       bool
	SilenceLevel      float32
	Model             string
	MessagesFile      string
	SystemPrompt      string
	MaxTokens         int
	Temperature       float32
	TopP              float32
	TopK              int
	MinP              float32
	FrequencyPenalty  float32
	PresencePenalty   float32
	RepetitionPenalty float32
	StopSequences     []string
	Seed              int
	ContextWindow     int
	LoRAs             []driver.LoRARef
}

func BuildTextRequest(args []string, trainEnabled bool, fields TextRequestFields) (driver.TextRequest, error) {
	if trainEnabled {
		return driver.TextRequest{
			Prompt: JoinArgs(args),
			Model:  fields.Model,
		}, nil
	}

	mediaFlags := []struct {
		flag string
		path string
		mode driver.TextInputMode
	}{
		{"--image", fields.ImagePath, driver.TextInputImage},
		{"--video", fields.VideoPath, driver.TextInputVideo},
		{"--audio", fields.AudioPath, driver.TextInputAudio},
		{"--document", fields.DocumentPath, driver.TextInputDocument},
	}

	var mode driver.TextInputMode
	var mediaPath string
	mediaCount := 0
	for _, mf := range mediaFlags {
		if strings.TrimSpace(mf.path) == "" {
			continue
		}
		mediaCount++
		if mediaCount > 1 {
			return driver.TextRequest{}, fmt.Errorf("only one media input allowed: use one of --image, --video, --audio, or --document")
		}
		mode = mf.mode
		mediaPath = strings.TrimSpace(mf.path)
	}

	messages, err := resolveChatMessages(fields)
	if err != nil {
		return driver.TextRequest{}, err
	}

	prompt, err := resolveTextPrompt(args, fields, mediaCount == 0 && len(messages) == 0)
	if err != nil {
		return driver.TextRequest{}, err
	}

	if mediaCount == 0 {
		mode = driver.TextInputPrompt
	}

	if fields.Translate && mode != driver.TextInputPrompt {
		return driver.TextRequest{}, fmt.Errorf("--translate is only valid with text input (prompt, --text, or stdin)")
	}

	if mode == driver.TextInputPrompt && strings.TrimSpace(prompt) == "" && len(messages) == 0 {
		return driver.TextRequest{}, fmt.Errorf("text input required: provide a prompt, --messages <file>, --text <file>, stdin (pipe or --stdin), or a media flag (--image, --video, --audio, --document)")
	}

	if audioInferenceFlagsSet(fields) && mode != driver.TextInputAudio {
		return driver.TextRequest{}, fmt.Errorf("audio transcription flags (--quality, --trim-silence, --silence-level) are only valid with --audio")
	}

	beamSize, err := ParseAudioQuality(fields.Quality)
	if err != nil {
		return driver.TextRequest{}, err
	}
	if err := ValidateSilenceLevel(fields.SilenceLevel); err != nil {
		return driver.TextRequest{}, err
	}
	trimSilence := fields.TrimSilence || fields.SilenceLevel > 0

	req := driver.TextRequest{
		InputMode:         mode,
		Messages:          messages,
		Prompt:            prompt,
		MediaPath:         mediaPath,
		Translate:         fields.Translate,
		TargetLang:        fields.TargetLang,
		Language:          fields.Language,
		BeamSize:          beamSize,
		VADEnabled:        trimSilence,
		VADThreshold:      fields.SilenceLevel,
		Model:             fields.Model,
		SystemPrompt:      fields.SystemPrompt,
		MaxTokens:         fields.MaxTokens,
		Temperature:       fields.Temperature,
		TopP:              fields.TopP,
		TopK:              fields.TopK,
		MinP:              fields.MinP,
		FrequencyPenalty:  fields.FrequencyPenalty,
		PresencePenalty:   fields.PresencePenalty,
		RepetitionPenalty: fields.RepetitionPenalty,
		StopSequences:     fields.StopSequences,
		ContextWindow:     fields.ContextWindow,
		LoRAs:             fields.LoRAs,
	}
	if fields.Seed >= 0 {
		req.Seed = &fields.Seed
	}
	return req, nil
}

type AudioInferenceFields struct {
	Quality      string
	TrimSilence  bool
	SilenceLevel float32
}

func AddAudioInferenceFlags(cmd *cobra.Command, f *AudioInferenceFields) {
	cmd.Flags().StringVar(&f.Quality, "quality", "", "transcription quality: fast, balanced, careful, or beam width 1–5 (default: backend default)")
	cmd.Flags().BoolVar(&f.TrimSilence, "trim-silence", false, "remove long silent passages before transcribing")
	cmd.Flags().Float32Var(&f.SilenceLevel, "silence-level", 0, "silence trimming strength 0.0–1.0 (0 = backend default; implies --trim-silence when > 0)")
}

func audioInferenceFlagsSet(fields TextRequestFields) bool {
	return fields.Quality != "" || fields.TrimSilence || fields.SilenceLevel > 0
}

func ParseAudioQuality(raw string) (int, error) {
	normalized := strings.TrimSpace(strings.ToLower(raw))
	if normalized == "" {
		return 0, nil
	}
	switch normalized {
	case "fast":
		return 1, nil
	case "balanced":
		return 3, nil
	case "careful":
		return 5, nil
	default:
		n, err := strconv.Atoi(normalized)
		if err != nil {
			return 0, fmt.Errorf("--quality must be fast, balanced, careful, or a number from 1 to 5")
		}
		if n < 1 || n > 5 {
			return 0, fmt.Errorf("--quality number must be between 1 and 5")
		}
		return n, nil
	}
}

func ValidateSilenceLevel(level float32) error {
	if level == 0 {
		return nil
	}
	if level < 0 || level > 1 {
		return fmt.Errorf("--silence-level must be between 0.0 and 1.0")
	}
	return nil
}

func resolveChatMessages(fields TextRequestFields) ([]driver.ChatMessage, error) {
	path := strings.TrimSpace(fields.MessagesFile)
	if path == "" {
		return nil, nil
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read --messages file: %w", err)
	}
	var messages []driver.ChatMessage
	if err := json.Unmarshal(raw, &messages); err != nil {
		return nil, fmt.Errorf("parse --messages file: %w", err)
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("--messages file must contain at least one message")
	}
	return messages, nil
}

func resolveTextPrompt(args []string, fields TextRequestFields, allowStdin bool) (string, error) {
	hasMessages := strings.TrimSpace(fields.MessagesFile) != ""
	hasArgs := len(args) > 0
	textFile := strings.TrimSpace(fields.TextFile)
	hasTextFile := textFile != ""

	if hasMessages && (hasArgs || hasTextFile || fields.UseStdin) {
		return "", fmt.Errorf("provide either --messages <file> or a prompt/--text/--stdin input, not both")
	}
	if hasArgs && hasTextFile {
		return "", fmt.Errorf("provide either a positional prompt or --text <file>, not both")
	}
	if hasArgs && fields.UseStdin {
		return "", fmt.Errorf("provide either a positional prompt or --stdin, not both")
	}
	if hasTextFile && fields.UseStdin {
		return "", fmt.Errorf("provide either --text <file> or --stdin, not both")
	}

	if hasArgs {
		return JoinArgs(args), nil
	}
	if hasTextFile {
		content, err := os.ReadFile(textFile)
		if err != nil {
			return "", fmt.Errorf("read --text file: %w", err)
		}
		return string(content), nil
	}
	if !allowStdin {
		return "", nil
	}
	if fields.UseStdin || stdinIsPiped(fields.StdinReader) {
		return readStdin(fields.StdinReader)
	}
	return "", nil
}

func readStdin(r io.Reader) (string, error) {
	if r == nil {
		r = os.Stdin
	}
	data, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read stdin: %w", err)
	}
	return string(data), nil
}

func stdinIsPiped(testReader io.Reader) bool {
	if testReader != nil {
		return true
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice == 0
}
