package dps

// GearSet defines equipment for all slots.
type GearSet struct {
	Name   string
	Style  string // melee, ranged, magic
	Head   string
	Cape   string
	Neck   string
	Ammo   string
	Weapon string
	Body   string
	Shield string
	Legs   string
	Hands  string
	Feet   string
	Ring   string
}

// Presets maps preset names to their gear sets.
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
		Ring:   "Berserker ring (i)",
	},
	"max-range": {
		Name:   "Max Range",
		Style:  "ranged",
		Head:   "Masori mask (f)",
		Cape:   "Ava's assembler",
		Neck:   "Necklace of anguish",
		Ammo:   "Dragon arrow",
		Weapon: "Twisted bow",
		Body:   "Masori body (f)",
		Shield: "", // tbow is 2h
		Legs:   "Masori chaps (f)",
		Hands:  "Zaryte vambraces",
		Feet:   "Pegasian boots",
		Ring:   "Archers ring (i)",
	},
	"max-mage": {
		Name:   "Max Mage",
		Style:  "magic",
		Head:   "Ancestral hat",
		Cape:   "Imbued god cape",
		Neck:   "Occult necklace",
		Weapon: "Sanguinesti staff",
		Body:   "Ancestral robe top",
		Shield: "", // sang is 2h
		Legs:   "Ancestral robe bottom",
		Hands:  "Tormented bracelet",
		Feet:   "Eternal boots",
		Ring:   "Magus ring",
	},
	"mid-melee": {
		Name:   "Mid Melee",
		Style:  "melee",
		Head:   "Neitiznot faceguard",
		Cape:   "Fire cape",
		Neck:   "Amulet of fury",
		Weapon: "Abyssal whip",
		Body:   "Bandos chestplate",
		Shield: "Dragon defender",
		Legs:   "Bandos tassets",
		Hands:  "Barrows gloves",
		Feet:   "Dragon boots",
		Ring:   "Berserker ring (i)",
	},
	"mid-range": {
		Name:   "Mid Range",
		Style:  "ranged",
		Head:   "Blessed coif",
		Cape:   "Ava's assembler",
		Neck:   "Necklace of anguish",
		Ammo:   "Amethyst broad bolts",
		Weapon: "Armadyl crossbow",
		Body:   "Blessed body",
		Shield: "", // ACB is not 2h, but we'll leave shield empty for simplicity
		Legs:   "Blessed chaps",
		Hands:  "Barrows gloves",
		Feet:   "Blessed boots",
		Ring:   "Archers ring (i)",
	},
	"void-range": {
		Name:   "Void Range",
		Style:  "ranged",
		Head:   "Void ranger helm",
		Cape:   "Ava's assembler",
		Neck:   "Necklace of anguish",
		Ammo:   "Dragon arrow",
		Weapon: "Twisted bow",
		Body:   "Elite void top",
		Shield: "",
		Legs:   "Elite void robe",
		Hands:  "Void knight gloves",
		Feet:   "Pegasian boots",
		Ring:   "Archers ring (i)",
	},
}

// PresetNames returns all available preset names.
func PresetNames() []string {
	names := make([]string, 0, len(Presets))
	for k := range Presets {
		names = append(names, k)
	}
	return names
}

// GearSlots returns the non-empty gear items as slot->name pairs.
func (g *GearSet) GearSlots() map[string]string {
	slots := make(map[string]string)
	if g.Head != "" {
		slots["head"] = g.Head
	}
	if g.Cape != "" {
		slots["cape"] = g.Cape
	}
	if g.Neck != "" {
		slots["neck"] = g.Neck
	}
	if g.Ammo != "" {
		slots["ammo"] = g.Ammo
	}
	if g.Weapon != "" {
		slots["weapon"] = g.Weapon
	}
	if g.Body != "" {
		slots["body"] = g.Body
	}
	if g.Shield != "" {
		slots["shield"] = g.Shield
	}
	if g.Legs != "" {
		slots["legs"] = g.Legs
	}
	if g.Hands != "" {
		slots["hands"] = g.Hands
	}
	if g.Feet != "" {
		slots["feet"] = g.Feet
	}
	if g.Ring != "" {
		slots["ring"] = g.Ring
	}
	return slots
}
