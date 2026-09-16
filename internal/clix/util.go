package clix

import "strings"

func JoinArgs(args []string) string {
	return strings.Join(args, " ")
}
