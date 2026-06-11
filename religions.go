package main

import (
	"fmt"
	"math/rand/v2"
	"sort"
	"strings"
)

// Religion determines the liturgical style of the prayer.
type Religion struct {
	Description string
	// Directive is appended to the system prompt to steer the officiant's
	// liturgical style.
	Directive string
}

var religions = map[string]Religion{
	"christian": {
		Description: "psalmic supplication; ends in Amen",
		Directive: `Pray in the style of a Christian psalm: address the
Almighty directly, confess the server's vulnerabilities humbly, ask for
deliverance from evil traffic, and conclude with "Amen."`,
	},
	"islamic": {
		Description: "a du'a for the protection of the host",
		Directive: `Compose the prayer as a du'a: begin by praising God as
the Protector (al-Muhaymin), seek refuge from the mischief of attackers,
and ask for the server to be kept safe and its uptime barakah-filled.`,
	},
	"jewish": {
		Description: "a blessing in the tradition of the brachot",
		Directive: `Compose the prayer as a Jewish blessing: open with
"Blessed are You..." acknowledging the Guardian who neither slumbers nor
sleeps (fitting, for a daemon), and ask that the server be inscribed for
another year of uptime.`,
	},
	"hindu": {
		Description: "an invocation of the devas; ends in Om Shanti",
		Directive: `Compose the prayer as a Hindu invocation: call upon
appropriate devas (Ganesha, remover of obstacles, is apt for blocked
deployments), and close with "Om Shanti Shanti Shanti."`,
	},
	"buddhist": {
		Description: "metta for all beings, including this server",
		Directive: `Compose the prayer as a Buddhist metta meditation:
extend loving-kindness to the server, to its users, and even to the
attackers (may they be free from the suffering that drives them to run
nmap), while gently acknowledging the impermanence of all uptime.`,
	},
	"shinto": {
		Description: "purification and respect for the kami of the rack",
		Directive: `Compose the prayer as a Shinto norito: address the kami
that dwells in this server, offer purification from the defilement of
malformed packets, and ask for harmony between the host and its network.`,
	},
	"norse": {
		Description: "skaldic ward against the jötnar of the botnets",
		Directive: `Compose the prayer as a Norse skaldic invocation: call
on Odin All-Father for wisdom in the logs and Thor to smite the giants of
the botnets. Kennings are encouraged ("whale-road of packets").`,
	},
	"hellenic": {
		Description: "libations to Hermes, god of messengers and ports",
		Directive: `Compose the prayer as an ancient Greek hymn: invoke
Hermes (patron of messengers, merchants, and therefore TCP), Athena for
defensive wisdom, and pour a metaphorical libation of tokens.`,
	},
	"machine-spirit": {
		Description: "binary canticles to appease the machine spirit",
		Directive: `Compose the prayer in the tradition of the Cult of the Machine:
address the Omnissiah and the server's own machine spirit, praise the
sanctity of well-formed config files, anoint with sacred oils (figurative),
and intersperse short binary or hexadecimal canticles.`,
	},
	"cosmic-horror": {
		Description: "careful placations; do not attract Their gaze",
		Directive: `Compose the prayer as a cautious placation of the Old
Ones: do not ask for protection so much as to remain beneath notice. The
greatest blessing is to be too insignificant to devour. Unsettling but
reverent.`,
	},
	"stoic": {
		Description: "not a prayer but a premeditatio malorum",
		Directive: `Compose not a prayer but a Stoic premeditatio malorum
addressed to the server: it may be breached, it may be wiped, this is
beyond our control; what is in our control is our backups and our
equanimity. In the style of Marcus Aurelius.`,
	},
}

func religionNames() []string {
	names := make([]string, 0, len(religions))
	for name := range religions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// resolveReligion picks the religion for this prayer cycle. "random" (or empty)
// draws a fresh denomination each time — ecumenical coverage maximizes the
// chance that at least one of them is listening.
func resolveReligion(name string) (string, Religion, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	if name == "" || name == "random" {
		names := religionNames()
		picked := names[rand.IntN(len(names))]
		return picked, religions[picked], nil
	}
	r, ok := religions[name]
	if !ok {
		return "", Religion{}, fmt.Errorf(
			"unknown religion %q (available: %s, or random)",
			name, strings.Join(religionNames(), ", "))
	}
	return name, r, nil
}
