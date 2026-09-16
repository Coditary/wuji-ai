#!/usr/bin/env bash
set -euo pipefail

min="${1:-80}"
shift

if [[ $# -eq 0 ]]; then
  echo "usage: $0 <min-percent> <package>..." >&2
  exit 2
fi

failed=0
for pkg in "$@"; do
  cover_out="$(mktemp)"
  if ! go test "$pkg" -coverprofile="$cover_out" -covermode=atomic >/dev/null; then
    echo "FAIL $pkg: tests failed" >&2
    failed=1
    rm -f "$cover_out"
    continue
  fi
  pct="$(go tool cover -func="$cover_out" | awk '/^total:/ { sub(/%/,"",$3); print $3 }')"
  rm -f "$cover_out"
  echo "$pkg: ${pct}%"
  if awk -v p="$pct" -v m="$min" 'BEGIN { exit (p+0 >= m+0) ? 0 : 1 }'; then
    :
  else
    echo "FAIL $pkg below ${min}% (got ${pct}%)" >&2
    failed=1
  fi
done

exit "$failed"
