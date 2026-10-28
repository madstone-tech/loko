#!/usr/bin/env bash
# Regenerates the site's bundled fonts (feature 021, research R6). Maintainers
# run it by hand; `loko build` never does. It downloads pinned OFL sources,
# checks their SHA-256, instances the variable fonts at the weights the theme
# uses and subsets them to Latin WOFF2. Commit what it writes.
#
# Needs: python3 with fontTools and brotli (pip install fonttools brotli);
# set PYTHON to choose the interpreter.
set -euo pipefail

repo_root=$(cd "$(dirname "$0")/../.." && pwd)
out="$repo_root/internal/adapters/html/theme/fonts"
branding="${MDSTN_BRANDING:-$repo_root/../mdstn-branding}"
py="${PYTHON:-python3}"

# google/fonts at a fixed commit, so the inputs never change under us.
gf="https://raw.githubusercontent.com/google/fonts/9862abc066c9b83bf84c3e8ed1fc2d8d48f9bbf0/ofl"

unicodes="U+0000-00FF,U+0131,U+0152-0153,U+02C6,U+02DA,U+02DC,U+2000-206F,U+20AC,U+2122,U+2190-2193,U+2212,U+25B8,U+25BE"

work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT

fetch() { # url dest sha256
	"$py" -I -c 'import sys, urllib.request
with urllib.request.urlopen(sys.argv[1]) as r, open(sys.argv[2], "wb") as f:
    f.write(r.read())' "$1" "$2"
	echo "$3  $2" | shasum -a 256 -c --quiet -
}

fetch "$gf/orbitron/Orbitron%5Bwght%5D.ttf" "$work/orbitron.ttf" f42db2dd16e642258e35782916eceb1dcdbea06fb958d77ad71dc5963587e8fd
fetch "$gf/orbitron/OFL.txt" "$work/OFL-orbitron.txt" ab609b0e110d622435ff337cdf233288556e011bbf9bd0550be98846c0630819
fetch "$gf/sourcesans3/SourceSans3%5Bwght%5D.ttf" "$work/sourcesans3.ttf" 042fe2cc0b933e328410d7acbd0aa6a1873dca5aef81875f4bc214b08825c7b9
fetch "$gf/sourcesans3/SourceSans3-Italic%5Bwght%5D.ttf" "$work/sourcesans3-italic.ttf" 39e3ab05ccd7cb94907c31005bb5bec1d5432f0b096a2b782976e217a540eb6c
fetch "$gf/sourcesans3/OFL.txt" "$work/OFL-source-sans-3.txt" 09746787287a289323b0ec3cff4d1a4a801331b82b7207c1e186f5d26619a392
fetch "$gf/jetbrainsmono/JetBrainsMono%5Bwght%5D.ttf" "$work/jetbrainsmono.ttf" 48715a42ec242c21e9f02692891e147d022299a52e48d5e413e1a942193ffeda
fetch "$gf/jetbrainsmono/OFL.txt" "$work/OFL-jetbrains-mono.txt" b2fe5e8987594e9ffd1d2ca52a2f5d73eb8335243893c5d6254b5ad69269591d
cp "$branding/fonts/Audiowide-Regular.ttf" "$work/audiowide.ttf"
cp "$branding/fonts/OFL.txt" "$work/OFL-audiowide.txt"

instance() { # src weight dest
	"$py" -m fontTools.varLib.instancer --quiet --no-recalc-timestamp -o "$3" "$1" "wght=$2"
}

subset() { # src dest
	"$py" -m fontTools.subset "$1" --output-file="$2" --flavor=woff2 \
		--layout-features='kern,liga' --unicodes="$unicodes" --no-hinting --desubroutinize
}

instance "$work/orbitron.ttf" 600 "$work/orbitron-600.ttf"
instance "$work/orbitron.ttf" 700 "$work/orbitron-700.ttf"
instance "$work/sourcesans3.ttf" 400 "$work/source-sans-3-400.ttf"
instance "$work/sourcesans3.ttf" 600 "$work/source-sans-3-600.ttf"
instance "$work/sourcesans3-italic.ttf" 400 "$work/source-sans-3-400-italic.ttf"
instance "$work/jetbrainsmono.ttf" 400 "$work/jetbrains-mono-400.ttf"
instance "$work/jetbrainsmono.ttf" 600 "$work/jetbrains-mono-600.ttf"
cp "$work/audiowide.ttf" "$work/audiowide-400.ttf"

mkdir -p "$out"
rm -f "$out"/*.woff2 "$out"/OFL-*.txt
for name in orbitron-600 orbitron-700 source-sans-3-400 source-sans-3-400-italic source-sans-3-600 \
	jetbrains-mono-400 jetbrains-mono-600 audiowide-400; do
	subset "$work/$name.ttf" "$out/$name.woff2"
done
cp "$work"/OFL-*.txt "$out/"

total=0
for f in "$out"/*.woff2; do
	size=$(wc -c <"$f" | tr -d ' ')
	total=$((total + size))
	printf '%8d  %s\n' "$size" "$(basename "$f")"
done
printf '%8d  total (%d KB)\n' "$total" "$((total / 1024))"
