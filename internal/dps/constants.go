package dps

// Monster ID constants from weirdgloop/osrs-dps-calc constants.ts
// Every ID list is ported 1-to-1 from the TypeScript source.

// ── Tombs of Amascut ───────────────────────────────────────────────

var AkkhaIDs = []int{11789, 11790, 11791, 11792, 11793, 11794, 11795, 11796}
var AkkhaShadowIDs = []int{11797, 11798, 11799}
var BabaIDs = []int{11778, 11779, 11780}
var KephriShieldedIDs = []int{11719}
var KephriUnshieldedIDs = []int{11721}
var KephriOverlordIDs = []int{11724, 11725, 11726}
var ZebakIDs = []int{11730, 11732, 11733}
var ToAObeliskIDs = []int{11751, 11750, 11752}
var P2WardenIDs = []int{11753, 11754, 11756, 11757}
var P3WardenIDs = []int{11761, 11763, 11762, 11764}
var ToAWardenCoreEjectedIDs = []int{11755, 11758}

var ToAPathMonsterIDs = concatIDs(AkkhaIDs, AkkhaShadowIDs, BabaIDs,
	KephriShieldedIDs, KephriUnshieldedIDs, KephriOverlordIDs, ZebakIDs)

var ToAMonsterIDs = concatIDs(ToAPathMonsterIDs, ToAObeliskIDs, P2WardenIDs,
	ToAWardenCoreEjectedIDs, P3WardenIDs)

// ── Theatre of Blood ───────────────────────────────────────────────

var VerzikP1IDs = []int{
	10830, 10831, 10832, // entry mode
	8369, 8370, 8371, // normal
	10847, 10848, 10849, // hard
}
var VerzikIDs = concatIDs(VerzikP1IDs, []int{
	10833, 10834, 10835, // em
	8372, 8373, 8374, // normal
	10850, 10851, 10852, // hard
})

var SotetsegIDs = []int{8387, 8388, 10867, 10868}

// ── Chambers of Xeric ──────────────────────────────────────────────

var TektonIDs = []int{7540, 7543, 7544, 7545}
var GuardianIDs = []int{7569, 7571, 7570, 7572}
var OlmHeadIDs = []int{7551, 7554}
var OlmMeleeHandIDs = []int{7552, 7555}
var OlmMageHandIDs = []int{7550, 7553}
var OlmIDs = concatIDs(OlmHeadIDs, OlmMeleeHandIDs, OlmMageHandIDs)
var ScavengerBeastIDs = []int{7548, 7549}
var AbyssalPortalIDs = []int{7533}
var GlowingCrystalIDs = []int{7568}
var IceDemonIDs = []int{7584, 7585}
var VespulaIDs = []int{7530, 7531, 7532}
var VespineSoldierIDs = []int{7538, 7539}
var DeathlyRangerIDs = []int{7559}

// ── Other bosses ───────────────────────────────────────────────────

var FragmentOfSerenIDs = []int{8917, 8918, 8919, 8920}
var NightmareIDs = []int{
	378, 9425, 9426, 9427, 9428, 9429, 9430, 9431, 9432, 9433, 9460,
	377, 9423, 9416, 9417, 9418, 9419, 9420, 9421, 9422, 9424, 11153, 11154, 11155,
}
var NightmareTotemIDs = []int{9434, 9437, 9440, 9443, 9435, 9438, 9441, 9444}
var NexIDs = []int{11278, 11279, 11280, 11281, 11282}
var ZulrahIDs = []int{2042, 2043, 2044}
var VardorvisIDs = []int{12223, 12224, 12228, 12425, 12426, 13656}

// ── Newer content ──────────────────────────────────────────────────

var TitanBossIDs = []int{12596, 14147}
var TitanElementalIDs = []int{14150, 14151}
var HueycoatlHeadIDs = []int{14009, 14010, 14013}
var HueycoatlBodyIDs = []int{14017}
var HueycoatlTailIDs = []int{14014}
var HueycoatlPhaseIDs = concatIDs(HueycoatlHeadIDs, HueycoatlTailIDs)
var DoomOfMokhaiotlIDs = []int{14707}
var EclipseMoonIDs = []int{13012}
var AbyssalSireTransitionIDs = []int{5886, 5889, 5891}
var YamaIDs_const = []int{14176}
var YamaVoidFlareIDs = []int{14179}
var TormentedDemonIDs = []int{13599, 13600, 13601, 13602, 13603, 13604, 13605, 13606}
var AraxxorIDs = []int{13668}

// ── Immunity lists ─────────────────────────────────────────────────

var DuskIDs = []int{7851, 7854, 7855, 7882, 7883, 7886, 7887, 7888, 7889}
var WarriorsGuildCyclopes = []int{2463, 2465, 2467, 2464, 2466, 2468, 2137, 2138, 2139, 2140, 2141, 2142}

var ImmuneToMeleeDamageIDs = concatIDs(
	[]int{494}, // Kraken
	AbyssalPortalIDs,
	[]int{7706, 7708},          // Zuk, Jal-MejJak
	[]int{12214, 12215, 12219}, // Leviathan
	ZulrahIDs,
)

var ImmuneToNonSalamanderMeleeIDs = []int{
	3169, 3170, 3171, 3172, 3173, 3174, 3175, 3176, 3177, 3178, 3179, 3180, 3181, 3182, 3183, // Aviansie
	7037, // Reanimated aviansie
}

var ImmuneToRangedDamageIDs = concatIDs(TektonIDs, DuskIDs, GlowingCrystalIDs, WarriorsGuildCyclopes)
var ImmuneToMagicDamageIDs = concatIDs(DuskIDs, WarriorsGuildCyclopes)

// ── Special behavior lists ─────────────────────────────────────────

var UsesDefLevelForMagicDefIDs = concatIDs(
	IceDemonIDs, VerzikIDs, FragmentOfSerenIDs,
	[]int{11709, 11712}, // Baboon Brawler
	[]int{9118},         // Rabbit (Prifddinas)
)

var OneHitMonsters = []int{7223, 8584, 11193}

var GuaranteedAccuracyMonsters = []int{5916} // Spawn (Abyssal Sire)

var BAAttackerMonsters = []int{
	1667, 5739, 5740, 5741, 5742, 5743, 5744, 5745, 5746, 5747, // fighters
	1668, 5757, 5758, 5759, 5760, 5761, 5762, 5763, 5764, 5765, // rangers
}

var UnderwaterMonsters = []int{7796}

var InfiniteHealthMonsters = []int{14779} // Gemstone crab

// AlwaysMaxHitMonsters: style-specific guaranteed max hit.
var AlwaysMaxHitMelee = concatIDs(
	[]int{11710, 11713}, // Baboon thrower
	[]int{12814},        // Frem warband archer
	ToAWardenCoreEjectedIDs,
	YamaVoidFlareIDs,
)
var AlwaysMaxHitRanged = concatIDs(
	[]int{11711, 11714}, // Baboon mage
	[]int{12815},        // Frem warband seer
	[]int{11717},        // Cursed baboon
	[]int{11715},        // Baboon shaman
	YamaVoidFlareIDs,
)
var AlwaysMaxHitMagic = concatIDs(
	[]int{11709, 11712}, // Baboon brawler
	[]int{12816},        // Frem warband berserker
	TitanElementalIDs,
	YamaVoidFlareIDs,
)

// ── Weapon spec data ───────────────────────────────────────────────

// WeaponSpecCosts maps weapon name to spec energy cost (out of 100).
var WeaponSpecCosts = map[string]int{
	"Armadyl godsword":          50,
	"Bandos godsword":           50,
	"Saradomin godsword":        50,
	"Zamorak godsword":          50,
	"Ancient godsword":          50,
	"Dragon dagger":             25,
	"Dragon claws":              50,
	"Dragon warhammer":          50,
	"Dragon halberd":            30,
	"Dragon mace":               25,
	"Dragon sword":              40,
	"Dragon longsword":          25,
	"Dragon scimitar":           55,
	"Abyssal dagger":            50,
	"Abyssal bludgeon":          50,
	"Abyssal whip":              50,
	"Granite hammer":            50,
	"Elder maul":                50,
	"Barrelchest anchor":        50,
	"Soulreaper axe":            0, // no cost, uses stacks
	"Toxic blowpipe":            50,
	"Zaryte crossbow":           75,
	"Heavy ballista":            65,
	"Light ballista":            65,
	"Dark bow":                  55,
	"Magic shortbow":            55,
	"Magic shortbow (i)":        50,
	"Webweaver bow":             50,
	"Volatile nightmare staff":  55,
	"Eldritch nightmare staff":  75,
	"Accursed sceptre":          50,
	"Accursed sceptre (a)":      50,
	"Voidwaker":                 50,
	"Crystal halberd":           30,
	"Saradomin sword":           100,
	"Saradomin's blessed sword": 65,
	"Burning claws":             50,
	"Bone claws":                50,
	"Tonalztics of ralos":       50,
	"Arkan blade":               50,
	"Osmumten's fang":           25,
	"Osmumten's fang (or)":      25,
	"Rosewood blowpipe":         50,
	"Brine sabre":               75,
	"Seercull":                  100,
	"Eye of ayak":               50,
}

// ── Helper ──────────────────────────────────────────────────────────

// concatIDs merges multiple int slices.
func concatIDs(slices ...[]int) []int {
	total := 0
	for _, s := range slices {
		total += len(s)
	}
	result := make([]int, 0, total)
	for _, s := range slices {
		result = append(result, s...)
	}
	return result
}

// ContainsID checks if an ID is in a list.
func ContainsID(ids []int, id int) bool {
	for _, i := range ids {
		if i == id {
			return true
		}
	}
	return false
}
