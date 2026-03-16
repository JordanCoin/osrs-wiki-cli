# osrs-wiki — OSRS Wiki CLI

CLI tool for the [Old School RuneScape Wiki](https://oldschool.runescape.wiki). Look up item images, Grand Exchange prices, training guides, boss strategies, and any wiki page — all from the command line.

Built for [AFK Mod](https://afkmod.app). Agent-friendly with `--json` output.

## Install

```bash
curl -sL https://github.com/JordanCoin/osrs-wiki-cli/releases/latest/download/osrs-wiki-linux-amd64 -o /usr/local/bin/osrs-wiki
chmod +x /usr/local/bin/osrs-wiki
```

## Image Lookup

Get the **correct** wiki image URL for any item. Never guesses filenames — uses the MediaWiki API. Case insensitive.

```bash
osrs-wiki image "Twisted Bow"              # 150px thumbnail
osrs-wiki image "Twisted Bow" --size 300   # larger thumbnail
osrs-wiki image "Twisted Bow" --full       # full resolution (1419x1457)
osrs-wiki image "Twisted Bow" --json       # structured output
```

Works with any casing:
```bash
osrs-wiki image "twisted bow"     # ✅
osrs-wiki image "TWISTED BOW"     # ✅
osrs-wiki image "Torva Platebody" # ✅
```

## Prices

Grand Exchange prices from the [real-time prices API](https://oldschool.runescape.wiki/w/RuneScape:Real-time_Prices) (updated every 5 minutes).

```bash
osrs-wiki price "Twisted Bow"        # Twisted Bow: 1.6B gp
osrs-wiki price "Dragon Claws"       # Dragon Claws: 62.5M gp
osrs-wiki price "Abyssal Whip"       # Abyssal whip: 1.4M gp
```

## Wiki Pages

Read any OSRS Wiki page as clean plaintext. No HTML, no scraping — uses the MediaWiki extracts API.

```bash
osrs-wiki wiki "Twisted Bow"           # Item page
osrs-wiki wiki "Chambers of Xeric"     # Raid info
osrs-wiki wiki "Guardians of the Rift" # Minigame
osrs-wiki wiki "Dragon Slayer II"      # Quest
osrs-wiki wiki "Vorkath/Strategies"    # Boss strategy
```

## Training & Strategy Guides

Shortcuts for common guide pages. Resolves skill names to the correct wiki training guide.

```bash
# Skill training guides
osrs-wiki guide runecraft     # → Pay-to-play Runecraft training
osrs-wiki guide slayer        # → Pay-to-play Slayer training
osrs-wiki guide mining        # → Pay-to-play Mining training
osrs-wiki guide sailing       # → Sailing training

# Boss strategies
osrs-wiki guide vorkath       # → Vorkath/Strategies
osrs-wiki guide zulrah        # → Zulrah/Strategies
osrs-wiki guide nex           # → Nex/Strategies
osrs-wiki guide toa           # → Tombs of Amascut/Strategies

# Other guides
osrs-wiki guide money         # → Money making guide
osrs-wiki guide quests        # → Optimal quest guide
```

## Item Info

Look up item details from the GE mapping (examine text, alch values, members status).

```bash
osrs-wiki item "Twisted Bow"
# Twisted bow (ID: 20997)
#   Examine: A mystical bow carved from the twisted remains of the Great Olm.
#   Members: Yes
#   High Alch: 150K gp
```

## Search

Find items and pages by partial name.

```bash
osrs-wiki search "dragon" --limit 5
# Dragon scimitar
# Dragon claws
# Dragon hunter crossbow
# Dragon warhammer
# Dragon bones
```

## JSON Output

Add `--json` to any command:

```bash
osrs-wiki image "Twisted Bow" --json
# {
#   "title": "Twisted bow",
#   "image_url": "https://oldschool.runescape.wiki/images/thumb/Twisted_bow_detail.png/150px-...",
#   "full_url": "https://oldschool.runescape.wiki/images/Twisted_bow_detail.png",
#   "width": 150,
#   "height": 154
# }
```

## Why This Exists

LLMs guess wiki image URLs and get them wrong — wrong capitalization, missing `_detail`, wrong file extension. This CLI uses the MediaWiki API to always return the correct URL. It also gives the model access to real wiki content instead of relying on stale training data.

## License

MIT
