package cli

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func playAudioFile(path string) error {
	if path == "" {
		return fmt.Errorf("empty audio path")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("audio file: %w", err)
	}

	var lastErr error
	for _, player := range audioPlayers(path) {
		if _, err := exec.LookPath(player.bin); err != nil {
			continue
		}
		cmd := exec.Command(player.bin, player.args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr != nil {
		return fmt.Errorf("could not play audio: %w", lastErr)
	}
	return fmt.Errorf("no audio player found (install paplay, ffplay, or mpv)")
}

type playerSpec struct {
	bin  string
	args []string
}

func audioPlayers(path string) []playerSpec {
	switch runtime.GOOS {
	case "darwin":
		return []playerSpec{
			{bin: "afplay", args: []string{path}},
			{bin: "ffplay", args: []string{"-nodisp", "-autoexit", "-loglevel", "quiet", path}},
			{bin: "mpv", args: []string{"--no-video", "--really-quiet", path}},
		}
	default:
		return []playerSpec{
			{bin: "paplay", args: []string{path}},
			{bin: "pw-play", args: []string{path}},
			{bin: "aplay", args: []string{path}},
			{bin: "ffplay", args: []string{"-nodisp", "-autoexit", "-loglevel", "quiet", path}},
			{bin: "mpv", args: []string{"--no-video", "--really-quiet", path}},
		}
	}
}
