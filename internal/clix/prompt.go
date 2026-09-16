package clix

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

func ReadLine(prompt string, in io.Reader) (string, error) {
	if in == nil {
		in = os.Stdin
	}
	fmt.Fprint(os.Stdout, prompt)
	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", err
		}
		return "", fmt.Errorf("input required")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

func ReadPassword(prompt string) (string, error) {
	fmt.Fprint(os.Stdout, prompt)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return ReadLine("", os.Stdin)
	}
	b, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stdout)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(b)), nil
}

func SelectOption(title string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, fmt.Errorf("no options")
	}
	fmt.Fprintf(os.Stdout, "\n%s\n", title)
	for i, opt := range options {
		fmt.Fprintf(os.Stdout, "  %d) %s\n", i+1, opt)
	}
	fmt.Fprint(os.Stdout, "  o) Other\n")
	fmt.Fprint(os.Stdout, "\nChoice [1]: ")
	line, err := ReadLine("", os.Stdin)
	if err != nil {
		return -1, err
	}
	if line == "" {
		return 0, nil
	}
	if strings.EqualFold(line, "o") || strings.EqualFold(line, "other") {
		return len(options), nil
	}
	var n int
	if _, err := fmt.Sscanf(line, "%d", &n); err != nil || n < 1 || n > len(options) {
		return -1, fmt.Errorf("invalid choice %q", line)
	}
	return n - 1, nil
}
