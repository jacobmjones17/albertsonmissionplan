// Package challenges holds the original twelve Monthly Challenges (seeded into the database on
// first run, and used by "Reset to original") and the artwork that shipped with them.
package challenges

import (
	"fmt"
	"io/fs"

	"github.com/albertson/albertsonmissionplan/web"
)

// Bullet is one "who: what" idea under a challenge (e.g. Families: Hold a weekly …).
// Text may use the site's simple inline markup: **bold** and [label](https://link).
type Bullet struct {
	Who  string `json:"who"`
	Text string `json:"text"`
}

// Default is one month's original challenge.
type Default struct {
	Month   int // 1 = January … 12 = December
	Slug    string
	Heading string // "Scripture Study Challenge" (month tiles drop the trailing "Challenge")
	Goal    string
	Bullets []Bullet
}

// Defaults mirrors frontend/src/monthlyData.ts (TestDefaultsMatchFrontend keeps them in sync).
var Defaults = []Default{
	{1, "scripture-study", "Scripture Study Challenge", "Encourage daily scripture study", []Bullet{
		{"Individuals", "Read a chapter from the Book of Mormon each day."},
		{"Families", "Hold a weekly family scripture study and discussion."},
		{"Ward", "Share insights and favorite verses in a weekly ward email or social media group."},
	}},
	{2, "act-of-kindness", "Act of Kindness Challenge", "Promote service and kindness", []Bullet{
		{"Individuals", "Perform one act of kindness each day (for example, compliment someone or help a neighbor)."},
		{"Families", "Plan and carry out a family service project (for example, clean the chapel or donate to those in need)."},
		{"Ward", "Organize a ward-wide service day (for example, a donation drive)."},
	}},
	{3, "temple-family-history", "Temple & Family History Challenge", "Foster connection with ancestors and temple work", []Bullet{
		{"Individuals", "Spend time each week on [FamilySearch.org](https://www.familysearch.org/en/)."},
		{"Families", "Share family stories and work on a family tree together."},
		{"Ward", "Host a family history workshop or temple trip."},
	}},
	{4, "missionary-work", "Missionary Work Challenge", "Increase missionary efforts and outreach", []Bullet{
		{"Individuals", "Share a gospel message or invitation to a church activity with a friend."},
		{"Families", "Invite neighbors to family home evening or a church event."},
		{"Ward", `Organize a "bring a friend" Sunday.`},
	}},
	{5, "health-wellness", "Health & Wellness Challenge", "Encourage physical, mental, and spiritual well-being", []Bullet{
		{"Individuals", "Set personal health goals (exercise, healthy eating, mindful rest)."},
		{"Families", "Plan and participate in a family fitness activity each week."},
		{"Ward", "Host a health and wellness fair."},
	}},
	{6, "faith-testimony", "Faith & Testimony Challenge", "Strengthen faith and personal testimonies", []Bullet{
		{"Individuals", "Write in a journal about personal faith experiences."},
		{"Families", "Share testimonies during family home evening."},
		{"Ward", "Hold a testimony meeting or fireside."},
	}},
	{7, "gratitude", "Gratitude Challenge", "Cultivate an attitude of gratitude", []Bullet{
		{"Individuals", "Keep a daily gratitude journal."},
		{"Families", "Create a gratitude tree or jar and add to it each day."},
		{"Ward", "Share expressions of gratitude in sacrament meeting or the ward group chat."},
	}},
	{8, "prayer", "Prayer Challenge", "Improve and deepen prayer habits", []Bullet{
		{"Individuals", "Commit to praying morning and night."},
		{"Families", "Hold family prayer daily."},
		{"Ward", "Organize a ward prayer chain."},
	}},
	{9, "sabbath-day", "Sabbath Day Challenge", "Honor and keep the Sabbath day holy", []Bullet{
		{"Individuals", "Study Sabbath-related scriptures and talks."},
		{"Families", "Plan meaningful Sabbath activities."},
		{"Ward", "Share Sabbath experiences in ward meetings."},
	}},
	{10, "conference", "General Conference Challenge", "Study and apply general conference", []Bullet{
		{"Individuals", "Study talks from previous conferences."},
		{"Families", "Discuss conference messages and set goals."},
		{"Ward", "Host a conference recap and goal-setting session."},
	}},
	{11, "service", "Service Challenge", "Serve others in meaningful ways", []Bullet{
		{"Individuals", "Volunteer at local charities or in the community."},
		{"Families", "Plan and carry out a family service project."},
		{"Ward", "Organize a ward-wide service initiative."},
	}},
	{12, "christmas", "Christmas Challenge", "Focus on the true meaning of Christmas", []Bullet{
		{"Individuals", "Study the life of Christ and His teachings."},
		{"Families", `Participate in "Light the World"–style activities.`},
		{"Ward", "Host a nativity event or Christmas program."},
	}},
}

// ForMonth returns the original challenge for month 1–12.
func ForMonth(month int) (Default, bool) {
	if month < 1 || month > len(Defaults) {
		return Default{}, false
	}
	return Defaults[month-1], true
}

// Image sizes for challenge art (3:2). The display image is used on wide screens; the thumb on
// phones, the printable sheet, and the offline download.
const (
	ImageWidth  = 960
	ImageHeight = 640
	ThumbWidth  = 480
	ThumbHeight = 320
)

// Art is one month's stored artwork.
type Art struct {
	ContentType  string
	Image, Thumb []byte
}

// DefaultArt returns the original WebP artwork for a month, read from the embedded frontend
// build (backend/web/dist/site/monthly/<slug>-{960,480}.webp).
func DefaultArt(month int) (Art, error) {
	d, ok := ForMonth(month)
	if !ok {
		return Art{}, fmt.Errorf("challenges: no month %d", month)
	}
	img, err := fs.ReadFile(web.Dist, "dist/site/monthly/"+d.Slug+"-960.webp")
	if err != nil {
		return Art{}, fmt.Errorf("challenges: default art for %s: %w", d.Slug, err)
	}
	thumb, err := fs.ReadFile(web.Dist, "dist/site/monthly/"+d.Slug+"-480.webp")
	if err != nil {
		return Art{}, fmt.Errorf("challenges: default art for %s: %w", d.Slug, err)
	}
	return Art{ContentType: "image/webp", Image: img, Thumb: thumb}, nil
}
