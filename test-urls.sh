#!/bin/bash

set -uo pipefail

# The directory of the script
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

declare -A URL_TITLE_MAP=(
    # Archidekt with 2 commanders
    ["https://archidekt.com/decks/552612#Look,_Ma!_Almost_No_Lands!"]="Look, Ma! Almost No Lands!;Look, Ma! Almost No Lands! - Tokens"
    # tappedout standard deck with Commander
    ["https://tappedout.net/mtg-decks/mogis-a-very-silly-commander/"]="Mogis- A Very Silly Commander;Mogis- A Very Silly Commander - Maybeboard;Mogis- A Very Silly Commander - Tokens"
    # tappedout cube
    ["https://tappedout.net/mtg-cube-drafts/12-05-20-pauper-cube/"]="Pauper Cube"
    # deckstats.net
    ["https://deckstats.net/decks/161156/1769395-memnarch"]="Memnarch (EDH - Commander);Memnarch (EDH - Commander) - Tokens"
    ["https://deckstats.net/decks/161156/1763076-kaalia-stax"]="Kaalia Stax (EDH - Commander);Kaalia Stax (EDH - Commander) - Tokens"
    # mtggoldfish.com
    ["https://www.mtggoldfish.com/deck/3435521#paper"]="James - S3W8;James - S3W8 - Sideboard;James - S3W8 - Tokens"
    # Moxfield
    ["https://www.moxfield.com/decks/yA6HKgjJS0O1J1X17skNqQ"]="Simic Swarm;Simic Swarm - Sideboard;Simic Swarm - Tokens"
    # Japanese Vanguard recipes
    ["https://cf-vanguard.com/deckrecipe/detail/wgp2017_t3_nagoya_4th"]="キム  さん;耀  さん;辰巳  さん"
    # English Vanguard recipes
    ["https://en.cf-vanguard.com/deckrecipe/detail/BSF2019_MY_VGS_3"]="Salty 5.0 - Neo Nectar;LBG - Shadow Paladin;SMOOTH - Oracle Think Tank"
)

tmp_dir="$(mktemp -d)"

# Deletes the temporary directory
function cleanup {
    rm -rf "${tmp_dir}"
    echo "Deleted temporary directory ${tmp_dir}"
}

# Register the cleanup function to be called on the EXIT signal
trap cleanup EXIT

# Build the CLI
(cd "${DIR}" && go build -ldflags="-s -w" -o "${tmp_dir}/tts-deckconverter" ./cmd/tts-deckconverter)

errors=()

for url in "${!URL_TITLE_MAP[@]}"; do
    IFS=";" read -ra titles <<< "${URL_TITLE_MAP[$url]}"

    echo "Checking ${url}…"
    echo

    # Import from URL
    "${tmp_dir}/tts-deckconverter" -output "${tmp_dir}" "${url}"

    echo

    # Check the generated files
    for title in "${titles[@]}"; do
        for file in "${tmp_dir}/${title}.json" "${tmp_dir}/${title}.png"; do
            echo "Checking if ${file} exists…"
            if [[ ! -f "${file}" ]]; then
                errors+=("File ${file} not found")
            fi
        done
    done

    echo

    sleep 5
done

for error in "${errors[@]}"; do
    >&2 echo "${error}"
done

if (( ${#errors[@]} != 0 )); then
    exit 1
fi
