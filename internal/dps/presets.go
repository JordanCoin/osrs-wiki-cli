package dps

// GearSet defines equipment for all slots.
type GearSet struct {
	Name    string
	Style   string // melee, ranged, magic
	Head    string
	Cape    string
	Neck    string
	Ammo    string
	Weapon  string
	Body    string
	Shield  string
	Legs    string
	Hands   string
	Feet    string
	Ring    string
	UseSpec bool // whether to calculate using special attack
}

// GearSlots returns a map of slot name to item name for iteration.
func (g GearSet) GearSlots() map[string]string {
	return map[string]string{
		"head":   g.Head,
		"cape":   g.Cape,
		"neck":   g.Neck,
		"ammo":   g.Ammo,
		"weapon": g.Weapon,
		"body":   g.Body,
		"shield": g.Shield,
		"legs":   g.Legs,
		"hands":  g.Hands,
		"feet":   g.Feet,
		"ring":   g.Ring,
	}
}

// Presets maps preset names to BiS gear sets.
// Based on https://oldschool.runescape.wiki/w/Armour/Highest_bonuses
var Presets = map[string]GearSet{
	"max-melee": {
		Name:   "Max Melee",
		Style:  "melee",
		Head:   "Torva full helm",
		Cape:   "Infernal cape",
		Neck:   "Amulet of torture",
		Weapon: "Scythe of vitur",
		Body:   "Torva platebody",
		Shield: "", // scythe is 2h
		Legs:   "Torva platelegs",
		Hands:  "Ferocious gloves",
		Feet:   "Primordial boots",
		Ring:   "Ultor ring",
	},
	"max-melee-1h": {
		Name:   "Max Melee (1H)",
		Style:  "melee",
		Head:   "Torva full helm",
		Cape:   "Infernal cape",
		Neck:   "Amulet of torture",
		Weapon: "Ghrazi rapier",
		Body:   "Torva platebody",
		Shield: "Avernic defender",
		Legs:   "Torva platelegs",
		Hands:  "Ferocious gloves",
		Feet:   "Primordial boots",
		Ring:   "Ultor ring",
	},
	"max-range": {
		Name:   "Max Range",
		Style:  "ranged",
		Head:   "Masori mask (f)",
		Cape:   "Blessed dizana's quiver",
		Neck:   "Necklace of anguish",
		Ammo:   "Dragon arrow",
		Weapon: "Twisted bow",
		Body:   "Masori body (f)",
		Shield: "", // tbow is 2h
		Legs:   "Masori chaps (f)",
		Hands:  "Zaryte vambraces",
		Feet:   "Pegasian boots",
		Ring:   "Venator ring",
	},
	"max-range-crossbow": {
		Name:   "Max Range (Crossbow)",
		Style:  "ranged",
		Head:   "Masori mask (f)",
		Cape:   "Blessed dizana's quiver",
		Neck:   "Necklace of anguish",
		Ammo:   "Ruby dragon bolts (e)",
		Weapon: "Zaryte crossbow",
		Body:   "Masori body (f)",
		Shield: "Twisted buckler",
		Legs:   "Masori chaps (f)",
		Hands:  "Zaryte vambraces",
		Feet:   "Pegasian boots",
		Ring:   "Venator ring",
	},
	"max-mage": {
		Name:   "Max Mage",
		Style:  "magic",
		Head:   "Ancestral hat",
		Cape:   "Imbued guthix cape",
		Neck:   "Occult necklace",
		Weapon: "Tumeken's shadow",
		Body:   "Ancestral robe top",
		Shield: "", // shadow is 2h
		Legs:   "Ancestral robe bottom",
		Hands:  "Tormented bracelet",
		Feet:   "Eternal boots",
		Ring:   "Magus ring",
	},
	"max-mage-sang": {
		Name:   "Max Mage (Sanguinesti)",
		Style:  "magic",
		Head:   "Ancestral hat",
		Cape:   "Imbued guthix cape",
		Neck:   "Occult necklace",
		Weapon: "Sanguinesti staff",
		Body:   "Ancestral robe top",
		Shield: "Arcane spirit shield",
		Legs:   "Ancestral robe bottom",
		Hands:  "Tormented bracelet",
		Feet:   "Eternal boots",
		Ring:   "Magus ring",
	},
	"fang": {
		Name:   "Fang Setup",
		Style:  "melee",
		Head:   "Torva full helm",
		Cape:   "Infernal cape",
		Neck:   "Amulet of torture",
		Weapon: "Osmumten's fang",
		Body:   "Torva platebody",
		Shield: "Avernic defender",
		Legs:   "Torva platelegs",
		Hands:  "Ferocious gloves",
		Feet:   "Primordial boots",
		Ring:   "Ultor ring",
	},
	"soulreaper": {
		Name:   "Soulreaper Axe",
		Style:  "melee",
		Head:   "Torva full helm",
		Cape:   "Infernal cape",
		Neck:   "Amulet of torture",
		Weapon: "Soulreaper axe",
		Body:   "Torva platebody",
		Shield: "", // 2h
		Legs:   "Torva platelegs",
		Hands:  "Ferocious gloves",
		Feet:   "Primordial boots",
		Ring:   "Ultor ring",
	},
	"void-range": {
		Name:   "Elite Void Range",
		Style:  "ranged",
		Head:   "Void ranger helm",
		Cape:   "Blessed dizana's quiver",
		Neck:   "Necklace of anguish",
		Ammo:   "Dragon arrow",
		Weapon: "Twisted bow",
		Body:   "Elite void top",
		Shield: "",
		Legs:   "Elite void robe",
		Hands:  "Void knight gloves",
		Feet:   "Pegasian boots",
		Ring:   "Venator ring",
	},
	"dharok": {
		Name:   "Dharok's (1 HP)",
		Style:  "melee",
		Head:   "Dharok's helm",
		Cape:   "Infernal cape",
		Neck:   "Amulet of torture",
		Weapon: "Dharok's greataxe",
		Body:   "Dharok's platebody",
		Shield: "",
		Legs:   "Dharok's platelegs",
		Hands:  "Ferocious gloves",
		Feet:   "Primordial boots",
		Ring:   "Ultor ring",
	},
	"slayer-melee": {
		Name:   "Slayer Melee",
		Style:  "melee",
		Head:   "Slayer helmet (i)",
		Cape:   "Infernal cape",
		Neck:   "Amulet of torture",
		Weapon: "Scythe of vitur",
		Body:   "Torva platebody",
		Shield: "",
		Legs:   "Torva platelegs",
		Hands:  "Ferocious gloves",
		Feet:   "Primordial boots",
		Ring:   "Ultor ring",
	},
	"slayer-range": {
		Name:   "Slayer Range",
		Style:  "ranged",
		Head:   "Slayer helmet (i)",
		Cape:   "Blessed dizana's quiver",
		Neck:   "Necklace of anguish",
		Ammo:   "Dragon arrow",
		Weapon: "Twisted bow",
		Body:   "Masori body (f)",
		Shield: "",
		Legs:   "Masori chaps (f)",
		Hands:  "Zaryte vambraces",
		Feet:   "Pegasian boots",
		Ring:   "Venator ring",
	},
}

// PresetNames returns sorted preset names.
func PresetNames() []string {
	names := make([]string, 0, len(Presets))
	for name := range Presets {
		names = append(names, name)
	}
	return names
}
