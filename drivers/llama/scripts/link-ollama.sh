#!/usr/bin/env bash
set -euo pipefail

# Creates symlinks from Ollama models into drivers/llama/models/ollama/
# so Wuji can use them via: wuji generate text "..." --driver llama --model ollama/<name>.gguf

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
LINK_DIR="$ROOT/models/ollama"
OLLAMA_MODELS="${OLLAMA_MODELS:-/usr/share/ollama/.ollama/models}"

sanitize() {
	echo "$1" | tr '/:' '__'
}

fix_permissions() {
	if ! test -r "$OLLAMA_MODELS/blobs" 2>/dev/null; then
		echo "Ollama blobs not readable — requesting sudo to fix group permissions..."
		if sudo chmod -R g+rX /usr/share/ollama/.ollama; then
			echo "Permissions fixed."
		else
			echo "Could not fix permissions. Run manually:"
			echo "  sudo chmod -R g+rX /usr/share/ollama/.ollama"
			return 1
		fi
	fi
}

mkdir -p "$LINK_DIR"
fix_permissions || echo "Continuing without permission fix..."

manifest="$LINK_DIR/manifest.json"
echo "{" > "$manifest"
first=true

count=0
while IFS= read -r name; do
	[[ -z "$name" ]] && continue

	blob=$(ollama show "$name" --modelfile 2>/dev/null | awk '/^FROM / { print $2; exit }')
	if [[ -z "$blob" ]]; then
		echo "skip: $name (no blob path)"
		continue
	fi

	link_name="$(sanitize "$name").gguf"
	link="$LINK_DIR/$link_name"
	ln -sf "$blob" "$link"

	rel="ollama/$link_name"
	if $first; then first=false; else echo "," >> "$manifest"; fi
	printf '  "%s": "%s"' "$rel" "$name" >> "$manifest"

	if [[ -r "$blob" ]]; then
		echo "linked: $name"
	else
		echo "linked: $name (via Ollama API)"
	fi
	((count++)) || true
done < <(ollama list 2>/dev/null | tail -n +2 | awk '{print $1}')

echo "" >> "$manifest"
echo "}" >> "$manifest"

echo ""
echo "Done: $count symlinks in $LINK_DIR"
echo ""
echo "Usage:"
echo "  ./bin/wuji-driver-llama"
echo "  ./bin/wuji generate text \"Hello\" --driver llama --model ollama/<name>.gguf"
