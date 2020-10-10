#!/bin/bash

set -euo pipefail

# The directory of the script
script_dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

declare -A url_title_map=(
    # Scryfall
    ["https://scryfall.com/@Rallemis/decks/9a1b2295-67cb-4a81-b900-2e2b1c1b6740"]="Rhys, Eater of Points;Rhys, Eater of Points - Tokens"
    # deckstats.net
    ["https://deckstats.net/decks/161156/1769395-memnarch"]="Memnarch (EDH - Commander);Memnarch (EDH - Commander) - Tokens"
    ["https://deckstats.net/decks/161156/1763076-kaalia-stax"]="Kaalia Stax (EDH - Commander);Kaalia Stax (EDH - Commander) - Tokens"
    # tappedout standard deck with Commander
    ["https://tappedout.net/mtg-decks/mogis-a-very-silly-commander/"]="Mogis- A Very Silly Commander;Mogis- A Very Silly Commander - Maybeboard;Mogis- A Very Silly Commander - Tokens"
    # tappedout cube
    ["https://tappedout.net/mtg-cube-drafts/12-05-20-pauper-cube/"]="Pauper Cube"
    # Deckbox
    ["https://deckbox.org/sets/2768129"]="Instant-Sorcery;Instant-Sorcery - Tokens"
    # mtggoldfish.com
    ["https://www.mtggoldfish.com/deck/3435521#paper"]="James - S3W8;James - S3W8 - Sideboard;James - S3W8 - Tokens"
    # Moxfield
    ["https://www.moxfield.com/decks/yA6HKgjJS0O1J1X17skNqQ"]="Simic Swarm;Simic Swarm - Tokens"
    # Manastack
    ["https://manastack.com/deck/ultra-competitive-vintage-beatdown-cu-deck"]="Ultra-Competitive Vintage Beatdown CU Deck;Ultra-Competitive Vintage Beatdown CU Deck - Sideboard"
    # Archidekt with 2 commanders
    ["https://archidekt.com/decks/552612#Look,_Ma!_Almost_No_Lands!"]="Look, Ma! Almost No Lands!;Look, Ma! Almost No Lands! - Tokens"
    # Aetherhub
    ["https://aetherhub.com/Deck/lurrus-rakdos-sacrifice-kroxa-bo1-cgb"]="Lurrus Rakdos Sacrifice Kroxa BO1 CGB;Lurrus Rakdos Sacrifice Kroxa BO1 CGB - Sideboard;Lurrus Rakdos Sacrifice Kroxa BO1 CGB - Tokens"
    ["https://aetherhub.com/Metagame/Standard-BO1/Deck/dimir-360243"]="Dimir Flash;Dimir Flash - Sideboard"
    # Frogtown
    ["https://www.frogtown.me/deckViewer/5f81151b581362577cef78b1/edit.html"]="Black Smart Life;Black Smart Life - Sideboard"
    # Cubetutor
    ["https://www.cubetutor.com/viewcube/14381"]="Hypercube;Hypercube - Tokens"
    ["https://www.cubetutor.com/cubedeck/605254"]="Power Dimir;Power Dimir - Tokens"
    # Cubecobra
    ["https://cubecobra.com/cube/overview/modovintage"]="Modo Vintage Cube;Modo Vintage Cube - Tokens"
    # mtg.wtf
    ["https://mtg.wtf/deck/znc/lands-wrath"]="Land's Wrath;Land's Wrath - Tokens"
    # YGOProDeck
    ["https://ygoprodeck.com/salamangreat-1st-place-locals-2020/"]="Salamangreat 1st Place Locals 2020;Salamangreat 1st Place Locals 2020 - Extra;Salamangreat 1st Place Locals 2020 - Side"
    # Yu-Gi-Oh Top Decks
    ["https://yugiohtopdecks.com/deck/8672"]="Adamancipator;Adamancipator - Extra"
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
(cd "${script_dir}" && go build -ldflags="-s -w" -o "${tmp_dir}/tts-deckconverter" ./cmd/tts-deckconverter)

errors=()

set +e

for url in "${!url_title_map[@]}"; do
    IFS=";" read -ra titles <<< "${url_title_map[$url]}"

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
    if [[ $# -gt 0 && "$1" == "--github-action" ]]; then
        # See https://docs.github.com/en/free-pro-team@latest/actions/reference/workflow-commands-for-github-actions#setting-an-error-message
        echo "::error file=$(basename "${BASH_SOURCE[0]}")::${error}"
    else
        >&2 echo "${error}"
    fi
done

if (( ${#errors[@]} != 0 )); then
    exit 1
fi
