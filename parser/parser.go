package parser

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MrR0b0t1001/mcoc-parser/pkg/types"
	"github.com/PuerkitoBio/goquery"
)

func FetchChampionHTML(name string) (title, html string, err error) {
	slug := "champion-spotlight-" + normalizeChampionName(name)
	fmt.Println(slug)

	// craft the api endpoint url we will need to hit
	// url encode the slug so it can be safely added
	endpointURL := fmt.Sprintf(
		"https://playcontestofchampions.com/wp-json/wp/v2/posts?slug=%s",
		url.QueryEscape(slug),
	)

	req, err := http.NewRequest("GET", endpointURL, nil)
	if err != nil {
		return "", "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", "", fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", err
	}

	posts := []types.ChampionPost{}

	if err := json.Unmarshal(body, &posts); err != nil {
		return "", "", err
	}

	if len(posts) == 0 {
		return "", "", fmt.Errorf("no post found for slug %q", slug)
	}

	return posts[0].Title.Rendered, posts[0].Content.Rendered, nil
}

func ParseChampion(titleRendered, contentRenderedHTML string) (*types.Champion, error) {
	title := strings.TrimSpace(html.UnescapeString((titleRendered)))
	if title == "" {
		title = "Unknown Champion"
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(contentRenderedHTML))
	if err != nil {
		return &types.Champion{}, err
	}

	ch := &types.Champion{Name: title}

	// 1) Character Class + Basic Abilities
	// Typically found in a <p> containing "Character Class:" and "Basic Abilities:"
	extractClassAndBasics(doc, ch)

	// 2) Strengths / Weaknesses
	extractStrengthsWeaknesses(doc, ch)

	// 3) Abilities + Notes
	extractAbilities(doc, ch)

	return ch, nil
}

//-------------------------------------------------------------------------------------------
// HELPERS

func extractClassAndBasics(doc *goquery.Document, ch *types.Champion) {
	doc.Find("p").Each(func(_ int, p *goquery.Selection) {
		txt := cleanText(p.Text())
		if !strings.Contains(txt, "Character Class:") &&
			!strings.Contains(txt, "Basic Abilities:") {
			return
		}

		// Example:
		// "Character Class: Skill Basic Abilities: Cruelty, Precision"
		cls := findAfterLabel(txt, "Character Class:")
		if cls != "" && ch.Class == "" {
			// stop at "Basic Abilities:" if present
			cls = strings.Split(cls, "Basic Abilities:")[0]
			ch.Class = strings.TrimSpace(cls)
		}

		basics := findAfterLabel(txt, "Basic Abilities:")
		if basics != "" && len(ch.BasicAbilities) == 0 {
			// split by comma
			parts := splitCSVLike(basics)
			ch.BasicAbilities = parts
		}
	})
}

func extractStrengthsWeaknesses(doc *goquery.Document, ch *types.Champion) {
	// The section is usually:
	// <h2 id="strengths">Strengths and Weaknesses</h2>
	// <h3>Strengths</h3> <ul>...
	// <h3>Weaknesses</h3> <ul>...
	//
	// But ids vary ("Strengths", "strengths"), so we search by heading text too.

	// Find the "Strengths" heading (<h3>) and read the next <ul> list items.
	doc.Find("h3").Each(func(_ int, h3 *goquery.Selection) {
		header := cleanText(h3.Text())

		switch header {
		case "Strengths":
			ul := nextMeaningfulList(h3)
			if ul != nil {
				ul.Find("li").Each(func(_ int, li *goquery.Selection) {
					t := cleanText(li.Text())
					if t != "" {
						ch.Strengths = appendUnique(ch.Strengths, t)
					}
				})
			}

		case "Weaknesses":
			ul := nextMeaningfulList(h3)
			if ul != nil {
				ul.Find("li").Each(func(_ int, li *goquery.Selection) {
					t := cleanText(li.Text())
					if t != "" {
						ch.Weaknesses = appendUnique(ch.Weaknesses, t)
					}
				})
			}
		}
	})
}

func extractAbilities(doc *goquery.Document, ch *types.Champion) {
	// Abilities section heading is often:
	// <h2 id="abilities">Abilities</h2>  (Black Widow)
	// or <h2 id="Ability">Abilities</h2> (Kraven)
	//
	// We'll locate an <h2> whose text is "Abilities" OR whose id matches abilities/Ability,
	// then iterate siblings until we hit the next big section (Synergy Bonuses / Recommended Masteries).

	start := findAbilitiesH2(doc)
	if start == nil {
		return
	}

	var current *types.Ability

	// Walk following siblings after the Abilities <h2>.
	for n := start.Next(); n != nil && n.Length() > 0; n = n.Next() {
		tag := strings.ToLower(goquery.NodeName(n))
		txt := cleanText(n.Text())

		// Stop conditions: next major section
		if tag == "h2" && isStopHeading(txt) {
			break
		}

		// Ability name headings are typically <h3>.
		if tag == "h3" {
			// flush previous
			if current != nil &&
				(current.Name != "" && (len(current.Effects) > 0 || len(current.DevNotes) > 0)) {
				ch.Abilities = append(ch.Abilities, *current)
			}
			current = &types.Ability{Name: txt}
			continue
		}

		if current == nil {
			continue
		}

		// Effects are usually <ul><li>...</li></ul>
		if tag == "ul" || tag == "ol" {
			n.Find("li").Each(func(_ int, li *goquery.Selection) {
				line := cleanText(li.Text())
				if line != "" {
					current.Effects = append(current.Effects, line)
				}
			})
			continue
		}

		// Sometimes there are effect lines in <p> (ex: "Passive" under Signature Ability)
		if tag == "p" {
			line := cleanText(n.Text())
			if line != "" && !looksLikeMetaNote(line) {
				current.Effects = append(current.Effects, line)
			}
			continue
		}

		// Notes are often:
		// <figure class="wp-block-pullquote"><blockquote><p>...</p><cite>Dev Notes</cite></blockquote></figure>
		if tag == "figure" && strings.Contains(n.AttrOr("class", ""), "wp-block-pullquote") {
			// Only treat as note if cite contains "Dev Notes" or "Expert Player Notes" etc.
			cite := cleanText(n.Find("cite").First().Text())
			noteText := cleanText(n.Find("blockquote p").First().Text())
			if cite != "" && noteText != "" {
				current.DevNotes = append(current.DevNotes, fmt.Sprintf("%s: %s", cite, noteText))
			}
			continue
		}
	}

	// flush last
	if current != nil &&
		(current.Name != "" && (len(current.Effects) > 0 || len(current.DevNotes) > 0)) {
		ch.Abilities = append(ch.Abilities, *current)
	}
}

/* ---------------- small utilities ---------------- */

func findAbilitiesH2(doc *goquery.Document) *goquery.Selection {
	var found *goquery.Selection

	doc.Find("h2").EachWithBreak(func(_ int, h2 *goquery.Selection) bool {
		id := strings.TrimSpace(h2.AttrOr("id", ""))
		txt := cleanText(h2.Text())

		// Match ONLY the actual "Abilities" section:
		// - Some pages: <h2 id="abilities">Abilities</h2>
		// - Kraven pages: <h2 id="Ability">Abilities</h2>
		if strings.EqualFold(txt, "Abilities") &&
			(strings.EqualFold(id, "abilities") || strings.EqualFold(id, "ability") || id == "") {
			found = h2
			return false
		}

		// Extra: explicitly allow Kraven's id="Ability" even if text weirdly changes
		if strings.EqualFold(id, "Ability") && strings.EqualFold(txt, "Abilities") {
			found = h2
			return false
		}

		return true
	})

	return found
}

func isStopHeading(h2Text string) bool {
	// Add more stop headings as you encounter them.
	stop := []string{
		"Synergy Bonuses",
		"Recommended Masteries",
		"Masteries",
		"Synergies",
	}
	for _, s := range stop {
		if strings.EqualFold(h2Text, s) {
			return true
		}
	}
	return false
}

func nextMeaningfulList(h *goquery.Selection) *goquery.Selection {
	// Typically the next sibling is <ul>, but sometimes whitespace nodes exist.
	for n := h.Next(); n != nil && n.Length() > 0; n = n.Next() {
		tag := strings.ToLower(goquery.NodeName(n))
		if tag == "ul" || tag == "ol" {
			return n
		}
		// if we hit another heading first, abort
		if tag == "h2" || tag == "h3" {
			return nil
		}
	}
	return nil
}

func FormatForDiscord(ch *types.Champion) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("**%s**\n", ch.Name))
	b.WriteString(fmt.Sprintf("*Class:* %s\n\n", ch.Class))

	b.WriteString("__Basic Abilities__\n")
	for _, a := range ch.BasicAbilities {
		b.WriteString(fmt.Sprintf("• %s\n", a))
	}

	b.WriteString("\n__Strengths__\n")
	for _, s := range ch.Strengths {
		b.WriteString(fmt.Sprintf("• %s\n", s))
	}

	b.WriteString("\n__Weaknesses__\n")
	for _, w := range ch.Weaknesses {
		b.WriteString(fmt.Sprintf("• %s\n", w))
	}

	b.WriteString("\n__Abilities__\n")
	for _, ab := range ch.Abilities {
		b.WriteString(fmt.Sprintf("\n**%s**\n", ab.Name))

		for _, e := range ab.Effects {
			b.WriteString(fmt.Sprintf("  - %s\n", e))
		}

		for _, note := range ab.DevNotes {
			b.WriteString(fmt.Sprintf("  > %s\n", note))
		}
	}

	return b.String()
}
