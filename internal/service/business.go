package service

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"worldcup-realtime/internal/football"
)

const (
	defaultCompetitionName = "FIFA World Cup"
	liveWindow             = 3*time.Hour + 30*time.Minute
)

func BuildViewState(raw football.RawState, refreshInterval time.Duration) ViewState {
	now := time.Now().UTC()
	matches := football.SortedMatches(raw.Matches)
	teamByID := indexTeams(raw.Teams)
	summaryByID := collectTeamSummaries(matches, teamByID)
	groupByTeamID := inferTeamGroups(matches)
	standings, standingByTeamID := buildStandings(matches, summaryByID, groupByTeamID)
	teams := buildTeams(raw.Teams, groupByTeamID, standingByTeamID)
	liveMatches, upcomingMatches, finishedMatches, specialMatches := buildMatches(matches, teamByID, now)

	var currentMatch *MatchView
	var nextMatch *MatchView
	if len(upcomingMatches) > 0 {
		nextMatch = copyMatchView(upcomingMatches[0])
	}
	switch {
	case len(liveMatches) > 0:
		currentMatch = copyMatchView(liveMatches[0])
	case nextMatch != nil:
		currentMatch = copyMatchView(*nextMatch)
	case len(finishedMatches) > 0:
		currentMatch = copyMatchView(finishedMatches[0])
	}

	return ViewState{
		Meta: MetaView{
			CompetitionName:   competitionName(raw.Competition),
			CompetitionEmblem: raw.Competition.Emblem,
			Source:            sourceLabel(raw.Source),
			FetchedAt:         formatTime(raw.FetchedAt),
			RefreshSeconds:    refreshSeconds(refreshInterval),
			LiveCount:         len(liveMatches),
		},
		Teams:           teams,
		CurrentMatch:    currentMatch,
		NextMatch:       nextMatch,
		Highlight:       buildHighlight(currentMatch, nextMatch, len(liveMatches)),
		LiveMatches:     liveMatches,
		UpcomingMatches: upcomingMatches,
		FinishedMatches: finishedMatches,
		SpecialMatches:  specialMatches,
		Standings:       standings,
	}
}



func indexTeams(teams []football.Team) map[int]football.Team {
	byID := make(map[int]football.Team, len(teams))
	for _, team := range teams {
		if team.ID != 0 {
			byID[team.ID] = team
		}
	}
	return byID
}

func collectTeamSummaries(matches []football.Match, teamByID map[int]football.Team) map[int]football.TeamSummary {
	summaries := make(map[int]football.TeamSummary, len(teamByID))
	for id, team := range teamByID {
		summaries[id] = team.TeamSummary
	}
	for _, match := range matches {
		if match.HomeTeam.ID != 0 {
			summaries[match.HomeTeam.ID] = mergeSummary(summaries[match.HomeTeam.ID], match.HomeTeam)
		}
		if match.AwayTeam.ID != 0 {
			summaries[match.AwayTeam.ID] = mergeSummary(summaries[match.AwayTeam.ID], match.AwayTeam)
		}
	}
	return summaries
}

func mergeSummary(current, next football.TeamSummary) football.TeamSummary {
	if current.ID == 0 {
		current.ID = next.ID
	}
	if current.Name == "" {
		current.Name = next.Name
	}
	if current.ShortName == "" {
		current.ShortName = next.ShortName
	}
	if current.TLA == "" {
		current.TLA = next.TLA
	}
	if current.Crest == "" {
		current.Crest = next.Crest
	}
	return current
}

func inferTeamGroups(matches []football.Match) map[int]string {
	groups := make(map[int]string)
	for _, match := range matches {
		if match.Stage != "GROUP_STAGE" || match.Group == nil {
			continue
		}
		group := formatGroup(*match.Group)
		if match.HomeTeam.ID != 0 {
			groups[match.HomeTeam.ID] = group
		}
		if match.AwayTeam.ID != 0 {
			groups[match.AwayTeam.ID] = group
		}
	}
	return groups
}

type standingAccumulator struct {
	team           TeamMiniView
	group          string
	played         int
	won            int
	draw           int
	lost           int
	goalsFor       int
	goalsAgainst   int
	goalDifference int
	points         int
	form           []string
}

func buildStandings(matches []football.Match, summaries map[int]football.TeamSummary, groupByTeamID map[int]string) ([]StandingGroupView, map[int]StandingRowView) {
	groups := make(map[string]map[int]*standingAccumulator)
	ensureRow := func(group string, summary football.TeamSummary) *standingAccumulator {
		if group == "" || summary.ID == 0 {
			return nil
		}
		table, ok := groups[group]
		if !ok {
			table = make(map[int]*standingAccumulator)
			groups[group] = table
		}
		row, ok := table[summary.ID]
		if !ok {
			row = &standingAccumulator{
				team:  miniFromSummary(summary),
				group: group,
			}
			table[summary.ID] = row
		}
		return row
	}

	for id, group := range groupByTeamID {
		ensureRow(group, summaries[id])
	}

	for _, match := range matches {
		if match.Stage != "GROUP_STAGE" || match.Group == nil {
			continue
		}
		group := formatGroup(*match.Group)
		home := ensureRow(group, mergeSummary(summaries[match.HomeTeam.ID], match.HomeTeam))
		away := ensureRow(group, mergeSummary(summaries[match.AwayTeam.ID], match.AwayTeam))
		if home == nil || away == nil || !isFinishedStatus(match.Status) {
			continue
		}
		if match.Score.FullTime.Home == nil || match.Score.FullTime.Away == nil {
			continue
		}
		applyResult(home, away, *match.Score.FullTime.Home, *match.Score.FullTime.Away)
	}

	groupNames := make([]string, 0, len(groups))
	for group := range groups {
		groupNames = append(groupNames, group)
	}
	sort.SliceStable(groupNames, func(i, j int) bool {
		return groupLess(groupNames[i], groupNames[j])
	})

	standings := make([]StandingGroupView, 0, len(groupNames))
	standingByTeamID := make(map[int]StandingRowView)
	for _, groupName := range groupNames {
		rows := make([]*standingAccumulator, 0, len(groups[groupName]))
		for _, row := range groups[groupName] {
			rows = append(rows, row)
		}
		sort.SliceStable(rows, func(i, j int) bool {
			a := rows[i]
			b := rows[j]
			if a.points != b.points {
				return a.points > b.points
			}
			if a.goalDifference != b.goalDifference {
				return a.goalDifference > b.goalDifference
			}
			if a.goalsFor != b.goalsFor {
				return a.goalsFor > b.goalsFor
			}
			return displayName(a.team) < displayName(b.team)
		})

		table := make([]StandingRowView, 0, len(rows))
		for i, row := range rows {
			view := row.view(i + 1)
			table = append(table, view)
			standingByTeamID[view.Team.ID] = view
		}
		standings = append(standings, StandingGroupView{
			Group:       groupName,
			MemberCount: len(table),
			Rows:        table,
		})
	}
	return standings, standingByTeamID
}

func applyResult(home, away *standingAccumulator, homeGoals, awayGoals int) {
	home.played++
	away.played++
	home.goalsFor += homeGoals
	home.goalsAgainst += awayGoals
	away.goalsFor += awayGoals
	away.goalsAgainst += homeGoals

	switch {
	case homeGoals > awayGoals:
		home.won++
		home.points += 3
		home.form = append(home.form, "W")
		away.lost++
		away.form = append(away.form, "L")
	case homeGoals < awayGoals:
		away.won++
		away.points += 3
		away.form = append(away.form, "W")
		home.lost++
		home.form = append(home.form, "L")
	default:
		home.draw++
		away.draw++
		home.points++
		away.points++
		home.form = append(home.form, "D")
		away.form = append(away.form, "D")
	}

	home.goalDifference = home.goalsFor - home.goalsAgainst
	away.goalDifference = away.goalsFor - away.goalsAgainst
}

func (row *standingAccumulator) view(position int) StandingRowView {
	return StandingRowView{
		Position:       position,
		Team:           row.team,
		PlayedGames:    row.played,
		Won:            row.won,
		Draw:           row.draw,
		Lost:           row.lost,
		GoalsFor:       row.goalsFor,
		GoalsAgainst:   row.goalsAgainst,
		GoalDifference: row.goalDifference,
		Points:         row.points,
		Form:           strings.Join(lastItems(row.form, 5), ""),
	}
}

func buildTeams(teams []football.Team, groupByTeamID map[int]string, standingByTeamID map[int]StandingRowView) []TeamView {
	views := make([]TeamView, 0, len(teams))
	for _, team := range teams {
		standing := standingByTeamID[team.ID]
		views = append(views, TeamView{
			ID:        team.ID,
			Name:      team.Name,
			ShortName: fallback(team.ShortName, team.Name),
			TLA:       team.TLA,
			Crest:     team.Crest,
			Coach:     coachName(team.Coach),
			Group:     groupByTeamID[team.ID],
			Rank:      standing.Position,
			Played:    standing.PlayedGames,
			Points:    standing.Points,
			GoalDiff:  standing.GoalDifference,
			Squad:     mapPlayers(team.Squad),
		})
	}

	sort.SliceStable(views, func(i, j int) bool {
		a := views[i]
		b := views[j]
		if a.Group != b.Group {
			return groupLess(a.Group, b.Group)
		}
		if a.Rank != b.Rank {
			if a.Rank == 0 {
				return false
			}
			if b.Rank == 0 {
				return true
			}
			return a.Rank < b.Rank
		}
		return a.Name < b.Name
	})
	return views
}

func buildMatches(matches []football.Match, teamByID map[int]football.Team, now time.Time) ([]MatchView, []MatchView, []MatchView, []MatchView) {
	liveMatches := make([]MatchView, 0)
	upcomingMatches := make([]MatchView, 0, len(matches))
	finishedMatches := make([]MatchView, 0)
	specialMatches := make([]MatchView, 0)

	for _, match := range matches {
		view := mapMatch(match, teamByID, now)
		switch view.StatusGroup {
		case "live":
			liveMatches = append(liveMatches, view)
		case "finished":
			finishedMatches = append(finishedMatches, view)
		case "special":
			specialMatches = append(specialMatches, view)
		default:
			upcomingMatches = append(upcomingMatches, view)
		}
	}

	sort.SliceStable(liveMatches, func(i, j int) bool {
		return liveMatches[i].SortTimeUnix < liveMatches[j].SortTimeUnix
	})
	sortUpcoming(upcomingMatches, now)
	sort.SliceStable(finishedMatches, func(i, j int) bool {
		return finishedMatches[i].SortTimeUnix > finishedMatches[j].SortTimeUnix
	})
	sort.SliceStable(specialMatches, func(i, j int) bool {
		return specialMatches[i].SortTimeUnix < specialMatches[j].SortTimeUnix
	})
	return liveMatches, upcomingMatches, finishedMatches, specialMatches
}

func mapMatch(match football.Match, teamByID map[int]football.Team, now time.Time) MatchView {
	kickoff := kickoffTime(match)
	statusGroup := statusGroup(match, now)
	return MatchView{
		ID:           match.ID,
		HomeTeam:     miniFromMatchTeam(match.HomeTeam, teamByID),
		AwayTeam:     miniFromMatchTeam(match.AwayTeam, teamByID),
		HomeScore:    scoreText(match.Score.FullTime.Home, statusGroup),
		AwayScore:    scoreText(match.Score.FullTime.Away, statusGroup),
		Status:       strings.ToUpper(match.Status),
		StatusLabel:  statusLabel(match, statusGroup, now),
		StatusGroup:  statusGroup,
		Minute:       minuteLabel(match, statusGroup, now),
		KickoffUTC:   formatTime(kickoff),
		Stage:        formatStage(match.Stage),
		Group:        groupLabel(match),
		Matchday:     matchdayLabel(match.Matchday),
		SortTimeUnix: kickoff.Unix(),
	}
}

func buildHighlight(currentMatch, nextMatch *MatchView, liveCount int) HighlightView {
	if currentMatch != nil && currentMatch.StatusGroup == "live" {
		return HighlightView{
			Title:   "Trận hiện tại",
			Message: currentMatch.HomeTeam.ShortName + " vs " + currentMatch.AwayTeam.ShortName,
			Detail:  currentMatch.StatusLabel + " · " + currentMatch.Minute,
		}
	}
	if nextMatch != nil {
		return HighlightView{
			Title:   "Trận kế tiếp",
			Message: nextMatch.HomeTeam.ShortName + " vs " + nextMatch.AwayTeam.ShortName,
			Detail:  nextMatch.StatusLabel + " · " + nextMatch.KickoffUTC,
		}
	}
	if currentMatch != nil {
		return HighlightView{
			Title:   "Trận gần nhất",
			Message: currentMatch.HomeTeam.ShortName + " vs " + currentMatch.AwayTeam.ShortName,
			Detail:  currentMatch.StatusLabel,
		}
	}
	if liveCount > 0 {
		return HighlightView{
			Title:   "Đang cập nhật",
			Message: "Có trận đấu trong khung realtime",
			Detail:  "Server đang đồng bộ từ live endpoint.",
		}
	}
	return HighlightView{
		Title:   "Trạng thái",
		Message: "Đang chờ dữ liệu World Cup",
		Detail:  "Server sẽ hiển thị trận tiếp theo sau khi fetch thành công.",
	}
}

func mapPlayers(players []football.Player) []PlayerView {
	views := make([]PlayerView, 0, len(players))
	for _, player := range players {
		views = append(views, PlayerView{
			ID:          player.ID,
			Name:        player.Name,
			Position:    player.Position,
			Nationality: player.Nationality,
		})
	}
	return views
}

func miniFromMatchTeam(summary football.TeamSummary, teamByID map[int]football.Team) TeamMiniView {
	if summary.ID != 0 {
		if team, ok := teamByID[summary.ID]; ok {
			summary = mergeSummary(team.TeamSummary, summary)
		}
	}
	return miniFromSummary(summary)
}

func miniFromSummary(summary football.TeamSummary) TeamMiniView {
	name := fallback(summary.Name, "TBD")
	shortName := fallback(summary.ShortName, name)
	tla := fallback(summary.TLA, shortCode(shortName))
	return TeamMiniView{
		ID:        summary.ID,
		Name:      name,
		ShortName: shortName,
		TLA:       tla,
		Crest:     summary.Crest,
	}
}

func statusGroup(match football.Match, now time.Time) string {
	status := strings.ToUpper(match.Status)
	switch {
	case isLiveStatus(status) || isMatchTimeActive(match, now):
		return "live"
	case isFinishedStatus(status):
		return "finished"
	case isSpecialStatus(status):
		return "special"
	default:
		return "upcoming"
	}
}

func statusLabel(match football.Match, statusGroup string, now time.Time) string {
	status := strings.ToUpper(match.Status)
	if statusGroup == "live" && !isLiveStatus(status) {
		return "Đang cập nhật"
	}
	switch status {
	case "IN_PLAY":
		return "Đang đá"
	case "LIVE":
		return "Đang đá"
	case "PAUSED":
		return "Tạm dừng"
	case "BREAK":
		return "Nghỉ giữa hiệp"
	case "EXTRA_TIME":
		return "Hiệp phụ"
	case "PENALTY_SHOOTOUT":
		return "Luân lưu"
	case "FINISHED":
		return "Đã kết thúc"
	case "AWARDED":
		return "Xử thắng"
	case "SUSPENDED":
		return "Treo trận"
	case "POSTPONED":
		return "Hoãn"
	case "CANCELLED":
		return "Hủy"
	case "TIMED", "SCHEDULED":
		if kickoff := kickoffTime(match); !kickoff.IsZero() && kickoff.Before(now) {
			return "Chờ cập nhật"
		}
		return "Sắp diễn ra"
	default:
		return fallback(status, "Chưa rõ")
	}
}

func minuteLabel(match football.Match, statusGroup string, now time.Time) string {
	if statusGroup == "finished" {
		return "FT"
	}
	if statusGroup == "special" {
		return statusLabel(match, statusGroup, now)
	}
	if statusGroup != "live" {
		return "Kickoff"
	}

	kickoff := kickoffTime(match)
	if kickoff.IsZero() {
		return "LIVE"
	}
	minutes := int(now.Sub(kickoff).Minutes())
	if minutes < 1 {
		return "LIVE"
	}
	if minutes > 130 {
		return "120+"
	}
	return strconv.Itoa(minutes) + "'"
}

func scoreText(score *int, statusGroup string) string {
	if score == nil {
		return "-"
	}
	return strconv.Itoa(*score)
}

func isLiveStatus(status string) bool {
	switch strings.ToUpper(status) {
	case "IN_PLAY", "LIVE", "PAUSED", "BREAK", "EXTRA_TIME", "PENALTY_SHOOTOUT":
		return true
	default:
		return false
	}
}

func isFinishedStatus(status string) bool {
	switch strings.ToUpper(status) {
	case "FINISHED", "AWARDED":
		return true
	default:
		return false
	}
}

func isSpecialStatus(status string) bool {
	switch strings.ToUpper(status) {
	case "SUSPENDED", "POSTPONED", "CANCELLED":
		return true
	default:
		return false
	}
}

func isMatchTimeActive(match football.Match, now time.Time) bool {
	status := strings.ToUpper(match.Status)
	if isFinishedStatus(status) || isSpecialStatus(status) {
		return false
	}
	kickoff := kickoffTime(match)
	if kickoff.IsZero() {
		return false
	}
	return !now.Before(kickoff) && now.Before(kickoff.Add(liveWindow))
}

func kickoffTime(match football.Match) time.Time {
	if strings.TrimSpace(match.UTCDate) == "" {
		return time.Time{}
	}
	kickoff, err := time.Parse(time.RFC3339, match.UTCDate)
	if err != nil {
		return time.Time{}
	}
	return kickoff.UTC()
}

func sortUpcoming(matches []MatchView, now time.Time) {
	nowUnix := now.Unix()
	sort.SliceStable(matches, func(i, j int) bool {
		aFuture := matches[i].SortTimeUnix >= nowUnix
		bFuture := matches[j].SortTimeUnix >= nowUnix
		if aFuture != bFuture {
			return aFuture
		}
		if aFuture {
			return matches[i].SortTimeUnix < matches[j].SortTimeUnix
		}
		return matches[i].SortTimeUnix > matches[j].SortTimeUnix
	})
}

func copyMatchView(match MatchView) *MatchView {
	value := match
	return &value
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func refreshSeconds(interval time.Duration) int {
	seconds := int(interval / time.Second)
	if seconds <= 0 {
		return 45
	}
	return seconds
}

func competitionName(competition football.Competition) string {
	return fallback(competition.Name, defaultCompetitionName)
}

func sourceLabel(source string) string {
	source = strings.TrimSpace(source)
	if source != "" {
		return source
	}
	return "football-data"
}

func coachName(coach *football.Coach) string {
	if coach == nil {
		return "-"
	}
	return fallback(coach.Name, "-")
}

func groupLabel(match football.Match) string {
	if match.Group != nil && strings.TrimSpace(*match.Group) != "" {
		return formatGroup(*match.Group)
	}
	return ""
}

func matchdayLabel(matchday *int) string {
	if matchday == nil {
		return ""
	}
	return strconv.Itoa(*matchday)
}

func formatGroup(group string) string {
	group = strings.TrimSpace(group)
	if group == "" {
		return ""
	}
	upper := strings.ToUpper(group)
	if strings.HasPrefix(upper, "GROUP_") {
		return "Group " + strings.TrimPrefix(upper, "GROUP_")
	}
	if strings.HasPrefix(upper, "GROUP ") {
		return "Group " + strings.TrimSpace(group[6:])
	}
	return group
}

func formatStage(stage string) string {
	stage = strings.TrimSpace(stage)
	if stage == "" {
		return "World Cup"
	}
	parts := strings.Split(strings.ToLower(stage), "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func groupLess(a, b string) bool {
	aRank := groupRank(a)
	bRank := groupRank(b)
	if aRank != bRank {
		return aRank < bRank
	}
	return a < b
}

func groupRank(group string) int {
	upper := strings.ToUpper(strings.TrimSpace(group))
	upper = strings.TrimPrefix(upper, "GROUP_")
	upper = strings.TrimPrefix(upper, "GROUP ")
	if len(upper) == 1 && upper[0] >= 'A' && upper[0] <= 'Z' {
		return int(upper[0] - 'A')
	}
	return 999
}

func displayName(team TeamMiniView) string {
	return fallback(team.ShortName, fallback(team.Name, team.TLA))
}

func shortCode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "TBD"
	}
	letters := make([]rune, 0, 3)
	for _, r := range strings.ToUpper(value) {
		if r >= 'A' && r <= 'Z' {
			letters = append(letters, r)
		}
		if len(letters) == 3 {
			break
		}
	}
	if len(letters) == 0 {
		return "TBD"
	}
	return string(letters)
}

func lastItems(values []string, limit int) []string {
	if len(values) <= limit {
		return values
	}
	return values[len(values)-limit:]
}

func fallback(value, fallbackValue string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallbackValue
	}
	return value
}
