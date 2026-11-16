package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

const robotsFile = `# comments are ignored
User-agent: *
Disallow: /private/
Allow: /private/jobs/
Disallow: /*.pdf$

User-agent: OtherBot
User-agent: AutoReachBot
Disallow: /internal/
`

func TestRobotsRulesFollowTheMostSpecificGroupAndLongestMatch(t *testing.T) {
	ours := utils.ParseRobots([]byte(robotsFile), "AutoReachBot/1.0")
	anyone := utils.ParseRobots([]byte(robotsFile), "SomeCrawler/2.0")

	assert.False(t, ours.Allows("/internal/jobs"), "our own group applies")
	assert.True(t, ours.Allows("/private/page"), "the * group is ignored when a group names us")

	assert.False(t, anyone.Allows("/private/page"))
	assert.True(t, anyone.Allows("/private/jobs/1"), "the longer Allow wins")
	assert.False(t, anyone.Allows("/files/offer.pdf"))
	assert.True(t, anyone.Allows("/files/offer.pdf.html"), "$ anchors the end")
	assert.True(t, anyone.Allows("/jobs/1"))

	assert.True(t, utils.ParseRobots(nil, "AutoReachBot/1.0").Allows("/anything"))
	assert.False(t, utils.DisallowAllRobots().Allows("/jobs/1"))
}
