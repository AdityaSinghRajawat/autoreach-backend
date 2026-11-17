package jobBoards

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/AdityaSinghRajawat/autoreach-backend/config"
	jobTypes "github.com/AdityaSinghRajawat/autoreach-backend/pkg/customtypes/jobManagement"
	modelHelpers "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/helpers"
	jobModel "github.com/AdityaSinghRajawat/autoreach-backend/pkg/models/jobManagement"
	"github.com/AdityaSinghRajawat/autoreach-backend/pkg/utils"
)

type careerPageBoard struct {
	fetcher *utils.WebFetcher
}

func (b *careerPageBoard) BoardType() string { return string(modelHelpers.BoardTypeCareerPage) }

// Reads schema.org JobPosting data from a careers page or a sitemap's pages.
func (b *careerPageBoard) ListJobs(
	ctx context.Context,
	source *jobModel.JobSource,
	knownExternalIDs map[string]struct{},
) (jobTypes.BoardListing, error) {
	robots := newRobotsCache(b.fetcher)

	body, err := b.fetchAllowed(ctx, robots, source.BoardToken)
	if err != nil {
		return jobTypes.BoardListing{}, err
	}

	pages, sitemaps, isSitemap := utils.ParseSitemap(body)
	if !isSitemap {
		return jobTypes.BoardListing{
			Jobs:     jobPostingsOnPage(source.BoardToken, body),
			Complete: true,
		}, nil
	}

	for index, sitemap := range sitemaps {
		if index == config.GetCareerPageMaxSitemaps() {
			break
		}

		childBody, err := b.fetchAllowed(ctx, robots, sitemap)
		if err != nil {
			return jobTypes.BoardListing{}, err
		}

		childPages, _, _ := utils.ParseSitemap(childBody)
		pages = append(pages, childPages...)
	}

	summaries := []jobTypes.BoardJob{}

	for _, page := range pages {
		if len(page) > config.GetMaxExternalJobIDLength() || !robots.allows(ctx, page) ||
			(source.SearchQuery != nil && !strings.Contains(page, *source.SearchQuery)) {
			continue
		}

		summaries = append(summaries, jobTypes.BoardJob{ExternalID: page, URL: page})
	}

	jobs, err := readNewJobs(ctx, summaries, knownExternalIDs, 1,
		func(ctx context.Context, summary jobTypes.BoardJob) (*jobTypes.BoardJob, error) {
			page, err := b.fetcher.Fetch(ctx, summary.URL)
			if err != nil {
				return nil, err
			}

			postings := jobPostingsOnPage(summary.URL, page)
			if len(postings) == 0 {
				return nil, nil
			}

			postings[0].ExternalID = summary.ExternalID

			return &postings[0], nil
		},
	)

	return jobTypes.BoardListing{Jobs: jobs, Complete: true}, err
}

func (b *careerPageBoard) fetchAllowed(
	ctx context.Context,
	robots *robotsCache,
	pageURL string,
) ([]byte, error) {
	if !robots.allows(ctx, pageURL) {
		return nil, fmt.Errorf("read %s: %w", pageURL, config.ErrRobotsDisallowed)
	}

	return b.fetcher.Fetch(ctx, pageURL)
}

type robotsCache struct {
	fetcher *utils.WebFetcher
	rules   map[string]utils.RobotsRules
}

func newRobotsCache(fetcher *utils.WebFetcher) *robotsCache {
	return &robotsCache{fetcher: fetcher, rules: map[string]utils.RobotsRules{}}
}

// A missing robots.txt allows everything; one that cannot be read allows nothing (RFC 9309).
func (c *robotsCache) allows(ctx context.Context, pageURL string) bool {
	parsed, err := url.Parse(pageURL)
	if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") {
		return false
	}

	origin := parsed.Scheme + "://" + parsed.Host

	rules, cached := c.rules[origin]
	if !cached {
		body, err := c.fetcher.Fetch(ctx, origin+config.GetCareerPageRobotsPath())

		switch {
		case err == nil:
			rules = utils.ParseRobots(body, c.fetcher.UserAgent())
		case utils.ClassifyUpstreamError(err) == config.GetErrCodeUpstreamRejected():
			rules = utils.RobotsRules{}
		default:
			rules = utils.DisallowAllRobots()
		}

		c.rules[origin] = rules
	}

	return rules.Allows(parsed.EscapedPath())
}
