package store

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"worldcup-realtime/internal/football"
)

type FootballClient interface {
	Fetch(context.Context) (football.RawState, error)
}

type MemoryStore struct {
	mu              sync.RWMutex
	client          FootballClient
	refreshInterval time.Duration
	raw             football.RawState
	view            ViewState
	lastError       string
}

func NewMemoryStore(client FootballClient, refreshInterval time.Duration) *MemoryStore {
	return &MemoryStore{
		client:          client,
		refreshInterval: refreshInterval,
	}
}

func (s *MemoryStore) Run(ctx context.Context) {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			_ = s.Refresh(ctx)
		}
	}
}

func (s *MemoryStore) Refresh(ctx context.Context) error {
	raw, err := s.client.Fetch(ctx)
	lastError := ""
	if err != nil {
		lastError = err.Error()
		if len(raw.Matches.Matches) == 0 && len(raw.Teams.Teams) == 0 && len(raw.Standings.Standings) == 0 {
			return err
		}
	}

	view := buildView(raw, s.refreshInterval, lastError)

	s.mu.Lock()
	s.raw = raw
	s.view = view
	s.lastError = lastError
	s.mu.Unlock()
	return err
}

func (s *MemoryStore) Snapshot() ViewState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.view
}

type standingSummary struct {
	Group    string
	Rank     int
	Played   int
	Points   int
	GoalDiff int
}

func buildView(raw football.RawState, refreshInterval time.Duration, lastError string) ViewState {
	competition := firstCompetition(raw)
	standingByTeam := make(map[int]standingSummary)
	teamByID := make(map[int]TeamView)

	standings := buildStandings(raw.Standings, raw.Teams.Teams, raw.Matches.Matches, standingByTeam)
	for _, team := range raw.Teams.Teams {
		view := teamViewFromTeam(team)
		if summary, ok := standingByTeam[team.ID]; ok {
			view.Group = summary.Group
			view.Rank = summary.Rank
			view.Played = summary.Played
			view.Points = summary.Points
			view.GoalDiff = summary.GoalDiff
		}
		teamByID[team.ID] = view
	}

	for _, match := range raw.Matches.Matches {
		ensureTeam(teamByID, match.HomeTeam)
		ensureTeam(teamByID, match.AwayTeam)
	}
	for _, group := range standings {
		for _, row := range group.Rows {
			ensureTeam(teamByID, football.TeamRef{
				ID:        row.Team.ID,
				Name:      row.Team.Name,
				ShortName: row.Team.ShortName,
				TLA:       row.Team.TLA,
				Crest:     row.Team.Crest,
			})
		}
	}

	teams := make([]TeamView, 0, len(teamByID))
	for _, team := range teamByID {
		teams = append(teams, team)
	}
	sort.SliceStable(teams, func(i, j int) bool {
		if teams[i].Group != teams[j].Group {
			return teams[i].Group < teams[j].Group
		}
		if teams[i].Rank != teams[j].Rank {
			if teams[i].Rank == 0 {
				return false
			}
			if teams[j].Rank == 0 {
				return true
			}
			return teams[i].Rank < teams[j].Rank
		}
		return teams[i].Name < teams[j].Name
	})

	liveMatches, upcomingMatches, finishedMatches, specialMatches := classifyMatches(raw.Matches.Matches)
	var current *MatchView
	if len(liveMatches) > 0 {
		current = cloneMatch(liveMatches[0])
	} else if len(upcomingMatches) > 0 {
		current = cloneMatch(upcomingMatches[0])
	}

	var next *MatchView
	for _, match := range upcomingMatches {
		if current == nil || match.ID != current.ID {
			next = cloneMatch(match)
			break
		}
	}
	if next == nil && current != nil && current.StatusGroup == "upcoming" {
		next = cloneMatch(*current)
	}

	return ViewState{
		Meta: MetaView{
			CompetitionCode: competition.Code,
			CompetitionName: competition.Name,
			Source:          raw.Source,
			FetchedAt:       raw.FetchedAt.Format(time.RFC3339),
			LastError:       lastError,
			RefreshSeconds:  int(refreshInterval.Seconds()),
			LiveCount:       len(liveMatches),
			UpcomingCount:   len(upcomingMatches),
			FinishedCount:   len(finishedMatches),
			SpecialCount:    len(specialMatches),
		},
		Teams:           teams,
		CurrentMatch:    current,
		NextMatch:       next,
		Highlight:       buildHighlight(current, next, len(liveMatches), len(upcomingMatches), len(finishedMatches)),
		UpcomingMatches: upcomingMatches,
		FinishedMatches: finishedMatches,
		SpecialMatches:  specialMatches,
		Standings:       standings,
	}
}

func firstCompetition(raw football.RawState) football.Competition {
	for _, competition := range []football.Competition{raw.Matches.Competition, raw.Teams.Competition, raw.Standings.Competition} {
		if competition.Code != "" || competition.Name != "" {
			if competition.Code == "" {
				competition.Code = "WC"
			}
			if competition.Name == "" {
				competition.Name = "FIFA World Cup"
			}
			return competition
		}
	}
	return football.Competition{Code: "WC", Name: "FIFA World Cup"}
}

func buildStandings(raw football.StandingsResponse, teams []football.Team, matches []football.Match, standingByTeam map[int]standingSummary) []StandingGroupView {
	groupOrder := worldCupGroupNames()
	groupRows := make(map[string][]StandingRowView, len(groupOrder))
	for _, groupName := range groupOrder {
		groupRows[groupName] = []StandingRowView{}
	}

	groupByTeam := groupsFromMatches(matches)
	usedTeam := make(map[int]bool)

	fallbackGroupIndex := 0
	for _, group := range raw.Standings {
		rawGroupName := canonicalGroupName(group.Group)
		blockGroupCount := 1
		if rawGroupName == "" && len(group.Table) > 4 {
			blockGroupCount = (len(group.Table) + 3) / 4
		}
		for entryIndex, entry := range group.Table {
			groupName := rawGroupName
			if groupName == "" {
				groupName = groupByTeam[entry.Team.ID]
			}
			if groupName == "" && len(group.Table) > 4 {
				groupName = groupNameByIndex(fallbackGroupIndex + entryIndex/4)
			}
			if groupName == "" {
				groupName = groupNameByIndex(fallbackGroupIndex)
			}

			row := rowFromEntry(entry, len(groupRows[groupName])+1)
			groupRows[groupName] = append(groupRows[groupName], row)
			usedTeam[entry.Team.ID] = true
		}
		if rawGroupName == "" {
			fallbackGroupIndex += blockGroupCount
		}
	}

	for _, team := range teams {
		if team.ID == 0 || usedTeam[team.ID] {
			continue
		}
		groupName := groupByTeam[team.ID]
		if groupName == "" {
			groupName = firstGroupWithSpace(groupRows, groupOrder)
		}
		row := rowFromTeam(team, len(groupRows[groupName])+1)
		groupRows[groupName] = append(groupRows[groupName], row)
		usedTeam[team.ID] = true
	}

	groups := make([]StandingGroupView, 0, len(groupOrder))
	for _, groupName := range groupOrder {
		rows := groupRows[groupName]
		sort.SliceStable(rows, func(i, j int) bool {
			if rows[i].Points != rows[j].Points {
				return rows[i].Points > rows[j].Points
			}
			if rows[i].GoalDifference != rows[j].GoalDifference {
				return rows[i].GoalDifference > rows[j].GoalDifference
			}
			if rows[i].GoalsFor != rows[j].GoalsFor {
				return rows[i].GoalsFor > rows[j].GoalsFor
			}
			return rows[i].Position < rows[j].Position
		})

		for index := range rows {
			rows[index].Position = index + 1
			standingByTeam[rows[index].Team.ID] = standingSummary{
				Group:    groupName,
				Rank:     rows[index].Position,
				Played:   rows[index].PlayedGames,
				Points:   rows[index].Points,
				GoalDiff: rows[index].GoalDifference,
			}
		}

		groups = append(groups, StandingGroupView{
			Group:       groupName,
			MemberCount: len(rows),
			Rows:        rows,
		})
	}
	return groups
}

func classifyMatches(matches []football.Match) ([]MatchView, []MatchView, []MatchView, []MatchView) {
	live := []MatchView{}
	upcoming := []MatchView{}
	finished := []MatchView{}
	special := []MatchView{}

	for _, match := range matches {
		view := matchViewFromMatch(match)
		switch view.StatusGroup {
		case "live":
			live = append(live, view)
		case "finished":
			finished = append(finished, view)
		case "special":
			special = append(special, view)
		default:
			upcoming = append(upcoming, view)
		}
	}

	sort.SliceStable(live, func(i, j int) bool {
		return live[i].SortTimeUnix < live[j].SortTimeUnix
	})
	sort.SliceStable(upcoming, func(i, j int) bool {
		return upcoming[i].SortTimeUnix < upcoming[j].SortTimeUnix
	})
	sort.SliceStable(finished, func(i, j int) bool {
		return finished[i].SortTimeUnix > finished[j].SortTimeUnix
	})
	sort.SliceStable(special, func(i, j int) bool {
		return special[i].SortTimeUnix < special[j].SortTimeUnix
	})
	return live, upcoming, finished, special
}

func matchViewFromMatch(match football.Match) MatchView {
	kickoff, ok := parseAPITime(match.UTCDate)
	sortUnix := int64(0)
	kickoffUTC := match.UTCDate
	if ok {
		sortUnix = kickoff.Unix()
		kickoffUTC = kickoff.Format(time.RFC3339)
	}
	statusGroup := statusGroup(match.Status)

	return MatchView{
		ID:           match.ID,
		HomeTeam:     miniFromRef(match.HomeTeam),
		AwayTeam:     miniFromRef(match.AwayTeam),
		HomeScore:    scoreString(match.Score.FullTime.Home),
		AwayScore:    scoreString(match.Score.FullTime.Away),
		Status:       strings.ToUpper(match.Status),
		StatusLabel:  statusLabel(match.Status),
		StatusGroup:  statusGroup,
		Minute:       minuteString(match.Minute, match.InjuryTime),
		KickoffUTC:   kickoffUTC,
		Venue:        fallback(match.Venue, "Sân vận động đang cập nhật"),
		Stage:        displayStage(match.Stage),
		Group:        displayGroup(ptrString(match.Group)),
		Matchday:     matchdayString(match.Matchday),
		SortTimeUnix: sortUnix,
	}
}

func teamViewFromTeam(team football.Team) TeamView {
	squad := make([]PlayerView, 0, len(team.Squad))
	for _, player := range team.Squad {
		squad = append(squad, PlayerView{
			ID:          player.ID,
			Name:        player.Name,
			Position:    fallback(player.Position, "Đang cập nhật"),
			Nationality: player.Nationality,
			ShirtNumber: shirtNumber(player.ShirtNumber),
		})
	}

	coach := ""
	if team.Coach != nil {
		coach = team.Coach.Name
	}
	return TeamView{
		ID:        team.ID,
		Name:      fallback(team.Name, team.ShortName),
		ShortName: fallback(team.ShortName, team.Name),
		TLA:       team.TLA,
		Crest:     team.Crest,
		Area:      team.Area.Name,
		Coach:     fallback(coach, "Đang cập nhật"),
		Venue:     fallback(team.Venue, "Đang cập nhật"),
		Squad:     squad,
	}
}

func ensureTeam(teamByID map[int]TeamView, ref football.TeamRef) {
	if ref.ID == 0 {
		return
	}
	if _, ok := teamByID[ref.ID]; ok {
		return
	}
	teamByID[ref.ID] = TeamView{
		ID:        ref.ID,
		Name:      fallback(ref.Name, ref.ShortName),
		ShortName: fallback(ref.ShortName, ref.Name),
		TLA:       ref.TLA,
		Crest:     ref.Crest,
		Coach:     "Đang cập nhật",
		Venue:     "Đang cập nhật",
	}
}

func miniFromRef(ref football.TeamRef) TeamMiniView {
	return TeamMiniView{
		ID:        ref.ID,
		Name:      fallback(ref.Name, ref.ShortName),
		ShortName: fallback(ref.ShortName, ref.Name),
		TLA:       ref.TLA,
		Crest:     ref.Crest,
	}
}

func miniFromTeam(team football.Team) TeamMiniView {
	return TeamMiniView{
		ID:        team.ID,
		Name:      fallback(team.Name, team.ShortName),
		ShortName: fallback(team.ShortName, team.Name),
		TLA:       team.TLA,
		Crest:     team.Crest,
	}
}

func rowFromEntry(entry football.StandingEntry, position int) StandingRowView {
	return StandingRowView{
		Position:       position,
		Team:           miniFromRef(entry.Team),
		PlayedGames:    entry.PlayedGames,
		Won:            entry.Won,
		Draw:           entry.Draw,
		Lost:           entry.Lost,
		GoalsFor:       entry.GoalsFor,
		GoalsAgainst:   entry.GoalsAgainst,
		GoalDifference: entry.GoalDifference,
		Points:         entry.Points,
		Form:           entry.Form,
	}
}

func rowFromTeam(team football.Team, position int) StandingRowView {
	return StandingRowView{
		Position: position,
		Team:     miniFromTeam(team),
	}
}

func groupsFromMatches(matches []football.Match) map[int]string {
	groupByTeam := make(map[int]string)
	for _, match := range matches {
		groupName := canonicalGroupName(ptrString(match.Group))
		if groupName == "" {
			continue
		}
		if match.HomeTeam.ID != 0 {
			groupByTeam[match.HomeTeam.ID] = groupName
		}
		if match.AwayTeam.ID != 0 {
			groupByTeam[match.AwayTeam.ID] = groupName
		}
	}
	return groupByTeam
}

func firstGroupWithSpace(groupRows map[string][]StandingRowView, groupOrder []string) string {
	for _, groupName := range groupOrder {
		if len(groupRows[groupName]) < 4 {
			return groupName
		}
	}
	return groupOrder[len(groupOrder)-1]
}

func worldCupGroupNames() []string {
	return []string{
		"Group A",
		"Group B",
		"Group C",
		"Group D",
		"Group E",
		"Group F",
		"Group G",
		"Group H",
		"Group I",
		"Group J",
		"Group K",
		"Group L",
	}
}

func groupNameByIndex(index int) string {
	groups := worldCupGroupNames()
	if index < 0 {
		return groups[0]
	}
	if index >= len(groups) {
		return groups[len(groups)-1]
	}
	return groups[index]
}

func cloneMatch(match MatchView) *MatchView {
	return &match
}

func buildHighlight(current, next *MatchView, liveCount, upcomingCount, finishedCount int) HighlightView {
	if current == nil {
		return HighlightView{
			Title:   "Đang chờ dữ liệu",
			Message: "Server đã sẵn sàng nhận dữ liệu từ football-data hoặc fake-data.",
			Detail:  "Chưa có trận đấu nào trong bộ nhớ RAM.",
		}
	}

	if current.StatusGroup == "live" {
		return HighlightView{
			Title:   "Trận đang diễn ra",
			Message: fmt.Sprintf("%s %s-%s %s", current.HomeTeam.ShortName, current.HomeScore, current.AwayScore, current.AwayTeam.ShortName),
			Detail:  fmt.Sprintf("%s tại %s. Sắp tới: %s trận, đã kết thúc: %s trận.", current.Minute, current.Venue, strconv.Itoa(upcomingCount), strconv.Itoa(finishedCount)),
		}
	}

	if next != nil {
		return HighlightView{
			Title:   "Trận tiếp theo",
			Message: fmt.Sprintf("%s vs %s", next.HomeTeam.ShortName, next.AwayTeam.ShortName),
			Detail:  fmt.Sprintf("%s tại %s. Live hiện có: %s trận.", next.StatusLabel, next.Venue, strconv.Itoa(liveCount)),
		}
	}

	return HighlightView{
		Title:   current.StatusLabel,
		Message: fmt.Sprintf("%s vs %s", current.HomeTeam.ShortName, current.AwayTeam.ShortName),
		Detail:  current.Venue,
	}
}

func statusGroup(status string) string {
	switch strings.ToUpper(status) {
	case "IN_PLAY", "EXTRA_TIME", "PENALTY_SHOOTOUT":
		return "live"
	case "FINISHED", "AWARDED":
		return "finished"
	case "PAUSED", "SUSPENDED", "POSTPONED", "CANCELLED", "CANCELED":
		return "special"
	default:
		return "upcoming"
	}
}

func statusLabel(status string) string {
	switch strings.ToUpper(status) {
	case "SCHEDULED":
		return "Đã lên lịch"
	case "TIMED":
		return "Sắp diễn ra"
	case "IN_PLAY":
		return "Đang đá"
	case "EXTRA_TIME":
		return "Hiệp phụ"
	case "PENALTY_SHOOTOUT":
		return "Luân lưu"
	case "PAUSED":
		return "Tạm dừng"
	case "FINISHED":
		return "Kết thúc"
	case "SUSPENDED":
		return "Bị tạm hoãn"
	case "POSTPONED":
		return "Hoãn"
	case "CANCELLED", "CANCELED":
		return "Hủy"
	case "AWARDED":
		return "Xử thắng"
	default:
		return fallback(status, "Đang cập nhật")
	}
}

func displayStage(stage string) string {
	stage = strings.TrimSpace(stage)
	if stage == "" {
		return ""
	}
	stage = strings.ReplaceAll(stage, "_", " ")
	return strings.Title(strings.ToLower(stage))
}

func displayGroup(group string) string {
	return canonicalGroupName(group)
}

func canonicalGroupName(group string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return ""
	}

	normalized := strings.ToUpper(group)
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	normalized = strings.Trim(normalized, "_")
	normalized = strings.TrimPrefix(normalized, "GROUP_")
	normalized = strings.TrimPrefix(normalized, "GROUP")
	normalized = strings.Trim(normalized, "_")

	if len(normalized) == 1 && normalized[0] >= 'A' && normalized[0] <= 'L' {
		return "Group " + normalized
	}
	return ""
}

func parseAPITime(value string) (time.Time, bool) {
	if value == "" {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, false
	}
	return parsed, true
}

func scoreString(value *int) string {
	if value == nil {
		return "-"
	}
	return strconv.Itoa(*value)
}

func minuteString(minute, injuryTime *int) string {
	if minute == nil {
		return "Chưa bắt đầu"
	}
	if injuryTime != nil && *injuryTime > 0 {
		return fmt.Sprintf("%d+%d'", *minute, *injuryTime)
	}
	return fmt.Sprintf("%d'", *minute)
}

func matchdayString(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func shirtNumber(value *int) string {
	if value == nil {
		return "-"
	}
	return strconv.Itoa(*value)
}

func ptrString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func fallback(value, fallbackValue string) string {
	if strings.TrimSpace(value) == "" {
		return fallbackValue
	}
	return value
}
