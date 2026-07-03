package football

import (
	"net/url"
	"path"
)

type CompetitionResource string

const (
	MatchesResource CompetitionResource = "matches"
	TeamsResource   CompetitionResource = "teams"
)

type Endpoint interface {
	URL(*Client) string
}

type CompetitionEndpoint struct {
	Resource CompetitionResource
}

func (e CompetitionEndpoint) URL(
	c *Client,
) string {

	return c.competitionEndpoint(
		string(e.Resource),
	)
}

func (c *Client) competitionEndpoint(resource string) string {
	endpoint := path.Join("competitions", c.competitionCode, resource)
	if c.season == "" {
		return endpoint
	}
	values := url.Values{}
	values.Set("season", c.season)
	return endpoint + "?" + values.Encode()
}


