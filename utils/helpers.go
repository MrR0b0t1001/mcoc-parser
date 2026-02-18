package utils

import (
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/bwmarrin/discordgo"
)

func ParseBotCommand(content string) (int, bool) {
	content = strings.TrimSpace(content)

	if !strings.HasPrefix(content, "/") {
		return 0, false
	}

	champIDStr := strings.TrimSpace(strings.TrimPrefix(content, "/"))
	if champIDStr == "" {
		return 0, false
	}

	// convert champID to int
	champID, err := strconv.Atoi(champIDStr)
	if err != nil {
		return 0, false
	}

	return champID, true
}

func SendBatched(s *discordgo.Session, channelID, msg string, maxLen int) error {
	if maxLen <= 0 || maxLen > 2000 {
		maxLen = 1900
	}

	// Work in runes so we don't cut UTF-8 in the middle.
	for len(msg) > 0 {
		if utf8.RuneCountInString(msg) <= maxLen {
			_, err := s.ChannelMessageSend(channelID, msg)
			return err
		}

		// Take a window of maxLen runes
		cut := cutIndexByRunes(msg, maxLen)
		chunk := msg[:cut]

		// Prefer cutting at last newline or space inside the chunk
		if i := lastIndexAny(chunk, "\n"); i > 0 {
			chunk = msg[:i]
			cut = i
		} else if i := lastIndexAny(chunk, " "); i > 0 {
			chunk = msg[:i]
			cut = i
		}

		chunk = strings.TrimSpace(chunk)
		if chunk != "" {
			if _, err := s.ChannelMessageSend(channelID, chunk); err != nil {
				return err
			}
			// Small delay helps avoid hitting burst rate limits
			time.Sleep(200 * time.Millisecond)
		}

		msg = strings.TrimSpace(msg[cut:])
	}
	return nil
}

//------------------------- HELPERS ----------------------//

func cutIndexByRunes(s string, runeLimit int) int {
	if runeLimit <= 0 {
		return 0
	}
	i := 0
	for idx := range s {
		if i == runeLimit {
			return idx
		}
		i++
	}
	return len(s)
}

func lastIndexAny(s, chars string) int {
	return strings.LastIndexAny(s, chars)
}

//--------------------------- CHAMPION MAP HOLY SHIT ITS BIG-----------------------//

var ChampionList = map[int]string{
	1:   "Abomination",
	2:   "Abomination (Immortal)",
	3:   "Absorbing Man",
	4:   "Adam Warlock",
	5:   "Aegon",
	6:   "Agent Venom",
	7:   "Air-Walker",
	8:   "America Chavez",
	9:   "Angela",
	10:  "Annihilus",
	11:  "Ant-Man",
	12:  "Ant-Man (Future)",
	13:  "Anti-Venom",
	14:  "Apocalypse",
	15:  "Arcade",
	16:  "Archangel",
	17:  "Arnim Zola",
	18:  "Attuma",
	19:  "Baron Zemo",
	20:  "Bastion",
	21:  "Beast",
	22:  "Beta Ray Bill",
	23:  "Bishop",
	24:  "Black Bolt",
	25:  "Black Cat",
	26:  "Black Panther",
	27:  "Black Panther (Civil War)",
	28:  "Black Widow",
	29:  "Black Widow (Claire Voyant)",
	30:  "Black Widow (Deadly Origin)",
	31:  "Blade",
	32:  "Bullseye",
	33:  "Cable",
	34:  "Captain America",
	35:  "Captain America (Infinity War)",
	36:  "Captain America (Sam Wilson)",
	37:  "Captain America (WWII)",
	38:  "Captain Britain",
	39:  "Captain Marvel (Classic)",
	40:  "Captain Marvel (Movie)",
	41:  "Carnage",
	42:  "Cassandra Nova",
	43:  "Cassie Lang",
	44:  "Cheeilth",
	45:  "Civil Warrior",
	46:  "Colossus",
	47:  "Corvus Glaive",
	48:  "Cosmic Ghost Rider",
	49:  "Count Nefaria",
	50:  "Crossbones",
	51:  "Cull Obsidian",
	52:  "Cyclops (Blue Team)",
	53:  "Cyclops (New Xavier School)",
	54:  "Dani Moonstar",
	55:  "Daredevil (Classic)",
	56:  "Daredevil (Hells Kitchen)",
	57:  "Dark Phoenix",
	58:  "Darkhawk",
	59:  "Dazzler",
	60:  "Deadpool",
	61:  "Deadpool (X-Force)",
	62:  "Diablo",
	63:  "Doctor Doom",
	64:  "Doctor Octopus",
	65:  "Doctor Strange",
	66:  "Doctor Voodoo",
	67:  "Domino",
	68:  "Dormammu",
	69:  "Dracula",
	70:  "Dragon Man",
	71:  "Drax",
	72:  "Dust",
	73:  "Ebony Maw",
	74:  "Electro",
	75:  "Elektra",
	76:  "Elsa Bloodstone",
	77:  "Emma Frost",
	78:  "Enchantress",
	79:  "Falcon",
	80:  "Falcon (Joaquin Torres)",
	81:  "Franken-Castle",
	82:  "Galan",
	83:  "Gambit",
	84:  "Gamora",
	85:  "Gentle",
	86:  "Ghost",
	87:  "Ghost Rider",
	88:  "Gladiator",
	89:  "Goldpool",
	90:  "Gorr",
	91:  "Green Goblin",
	92:  "Groot",
	93:  "Guardian",
	94:  "Guillotine",
	95:  "Guillotine (Deathless)",
	96:  "Guillotine 2099",
	97:  "Gwenpool",
	98:  "Havok",
	99:  "Hawkeye",
	100: "Heimdall",
	101: "Hela",
	102: "Hercules",
	103: "High Evolutionary",
	104: "Hit-Monkey",
	105: "Howard the Duck",
	106: "Hulk",
	107: "Hulk (Immortal)",
	108: "Hulk (Ragnarok)",
	109: "Hulkbuster",
	110: "Hulkling",
	111: "Human Torch",
	112: "Hyperion",
	113: "Iceman",
	114: "Ikaris",
	115: "Imperiosa",
	116: "Invisible Woman",
	117: "Ironheart",
	118: "Iron Fist",
	119: "Iron Fist (Immortal)",
	120: "Iron Man",
	121: "Iron Man (Infamous)",
	122: "Iron Man (Infinity War)",
	123: "Iron Patriot",
	124: "Isophyene",
	125: "Jabari Panther",
	126: "Jack O Lantern",
	127: "Jean Grey",
	128: "Jessica Jones",
	129: "Joe Fixit",
	130: "Jubilee",
	131: "Juggernaut",
	132: "Kang",
	133: "Karolina Dean",
	134: "Karnak",
	135: "Kate Bishop",
	136: "Killmonger",
	137: "Kindred",
	138: "King Groot",
	139: "King Groot (Deathless)",
	140: "Kingpin",
	141: "Kitty Pryde",
	142: "Knull",
	143: "Korg",
	144: "Kraven",
	145: "Kushala",
	146: "Lady Deathstrike",
	147: "Lizard",
	148: "Loki",
	149: "Longshot",
	150: "Lumatrix",
	151: "Luke Cage",
	152: "M.O.D.O.K.",
	153: "Maestro",
	154: "Magik",
	155: "Magneto",
	156: "Magneto (House of X)",
	157: "Man-Thing",
	158: "Mangog",
	159: "Mantis",
	160: "Masacre",
	161: "Medusa",
	162: "Mephisto",
	163: "Mister Fantastic",
	164: "Mister Negative",
	165: "Mister Sinister",
	166: "Misty Knight",
	167: "Mojo",
	168: "Mole Man",
	169: "Moon Knight",
	170: "Moondragon",
	171: "Morbius",
	172: "Mordo",
	173: "Morningstar",
	174: "Mr. Knight",
	175: "Ms. Marvel",
	176: "Ms. Marvel (Kamala Khan)",
	177: "Mysterio",
	178: "Namor",
	179: "Nebula",
	180: "Negasonic Teenage Warhead",
	181: "Nick Fury",
	182: "Nico Minoru",
	183: "Night Thrasher",
	184: "Nightcrawler",
	185: "Nimrod",
	186: "Northstar",
	187: "Nova",
	188: "Odin",
	189: "Okoye",
	190: "Old Man Logan",
	191: "Omega Red",
	192: "Omega Sentinel",
	193: "Onslaught",
	194: "Patriot",
	195: "Peni Parker",
	196: "Phoenix",
	197: "Photon",
	198: "Platinumpool",
	199: "Professor X",
	200: "Prowler",
	201: "Proxima Midnight",
	202: "Psycho-Man",
	203: "Psylocke",
	204: "Punisher",
	205: "Punisher 2099",
	206: "Purgatory",
	207: "Quake",
	208: "Quicksilver",
	209: "Red Goblin",
	210: "Red Guardian",
	211: "Red Hulk",
	212: "Red Skull",
	213: "Rhino",
	214: "Rintrah",
	215: "Rocket Raccoon",
	216: "Rogue",
	217: "Ronan",
	218: "Ronin",
	219: "Sabretooth",
	220: "Sandman",
	221: "Sasquatch",
	222: "Sauron",
	223: "Scarlet Witch (Classic)",
	224: "Scarlet Witch (Sigil)",
	225: "Scorpion",
	226: "Scream",
	227: "Sentinel",
	228: "Sentry",
	229: "Sersi",
	230: "Shang-Chi",
	231: "Shathra",
	232: "She-Hulk",
	233: "She-Hulk (Deathless)",
	234: "Shocker",
	235: "Shuri",
	236: "Silver Samurai",
	237: "Silk",
	238: "Silver Centurion",
	239: "Silver Sable",
	240: "Silver Samurai",
	241: "Silver Surfer",
	242: "Solvarch",
	243: "Sorceror Supreme",
	244: "Spider-Gwen",
	245: "Spider-Ham",
	246: "Spider-Man (Classic)",
	247: "Spider-Man (Miles Morales)",
	248: "Spider-Man (Pavitr Prabhakar)",
	249: "Spider-Man (Stark Enhanced)",
	250: "Spider-Man (Stealth Suit)",
	251: "Spider-Man (Supreme)",
	252: "Spider-Man (Symbiote)",
	253: "Spider-Man 2099",
	254: "Spider-Punk",
	255: "Spider-Woman",
	256: "Spiral",
	257: "Spot",
	258: "Squirrel Girl",
	259: "Star-Lord",
	260: "Star-Lord (Stellar-Forged)",
	261: "Storm",
	262: "Storm (Pyramid X)",
	263: "Stryfe",
	264: "Sunspot",
	265: "Super-Skrull",
	266: "Superior Iron Man",
	267: "Symbiote Supreme",
	268: "Taskmaster",
	269: "Terrax",
	270: "Thanos",
	271: "Thanos (Deathless)",
	272: "The Champion",
	273: "The Destroyer",
	274: "The Hood",
	275: "The Leader",
	276: "The Maker",
	277: "The Overseer",
	278: "The Serpent",
	279: "Thing",
	280: "Thor",
	281: "Thor (Jane Foster)",
	282: "Thor (Ragnarok)",
	283: "Tigra",
	284: "Titania",
	285: "Toad",
	286: "Ultron",
	287: "Ultron (Lab)",
	288: "Unstoppable Colossus",
	289: "Valkyrie",
	290: "Venom",
	291: "Venom the Duck",
	292: "Venompool",
	293: "Vision",
	294: "Vision (Aarkus)",
	295: "Vision (Age of Ultron)",
	296: "Vision (Deathless)",
	297: "Viv Vision",
	298: "Void",
	299: "Vox",
	300: "Vulture",
	301: "War Machine",
	302: "Warlock",
	303: "Wasp",
	304: "Werewolf By Night",
	305: "White Tiger",
	306: "Wiccan",
	307: "Winter Soldier",
	308: "Wolverine",
	309: "Wolverine (Weapon X)",
	310: "Wolverine (X-23)",
	311: "Wong",
	312: "Yellowjacket",
	313: "Yelena Belova",
	314: "Yondu",
}
