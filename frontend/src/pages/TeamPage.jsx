import { ArrowLeft } from "lucide-react";
import { DetailItem } from "../components/common/DetailItem.jsx";
import { MatchRow } from "../components/matches/MatchRow.jsx";
import { StandingsPanel } from "../components/standings/StandingsPanel.jsx";
import { TeamsPanel } from "../components/teams/TeamsPanel.jsx";
import { Crest } from "../components/common/Crest.jsx";
import { getAllMatches } from "../utils/matches.js";
import { formatGoalDiff, shortPosition } from "../utils/team.js";
import { NotFoundPage } from "./NotFoundPage.jsx";

export function TeamPage({ data, teamId, onNavigate }) {
  const team = (data.teams || []).find((item) => item.id === teamId);
  const matches = getAllMatches(data).filter((match) => match.homeTeam.id === teamId || match.awayTeam.id === teamId);

  if (!team) return <NotFoundPage title="Không tìm thấy đội" onNavigate={onNavigate} />;

  return (
    <main className="detail-layout">
      <TeamsPanel teams={data.teams || []} onNavigate={onNavigate} />
      <section className="detail-page">
        <button className="back-button" type="button" onClick={() => onNavigate("/")}>
          <ArrowLeft size={16} aria-hidden="true" />
          Dashboard
        </button>
        <div className="detail-hero">
          <Crest src={team.crest} fallback={team.tla} />
          <div>
            <p className="eyebrow">{team.group || "World Cup"}</p>
            <h1>{team.name}</h1>
            <p className="detail-muted">{team.coach ? `HLV: ${team.coach}` : "Chưa có thông tin HLV"}</p>
          </div>
        </div>
        <div className="detail-grid">
          <DetailItem label="Rank" value={team.rank ? `#${team.rank}` : "-"} />
          <DetailItem label="Points" value={team.points || 0} />
          <DetailItem label="Played" value={team.played || 0} />
          <DetailItem label="Goal diff" value={formatGoalDiff(team.goalDiff || 0)} />
        </div>
        <h2 className="section-title">Trận đấu của đội</h2>
        <div className="match-list detail-matches">
          {matches.length ? (
            matches.map((match) => <MatchRow key={match.id} match={match} onNavigate={onNavigate} />)
          ) : (
            <div className="empty-state">Chưa có trận đấu.</div>
          )}
        </div>
        <h2 className="section-title">Cầu thủ</h2>
        <div className="squad-grid">
          {(team.squad || []).length ? (
            team.squad.map((player) => (
              <div key={player.id} className="player-card">
                <strong>{player.name}</strong>
                <span>{shortPosition(player.position)}</span>
                <small>{player.nationality}</small>
              </div>
            ))
          ) : (
            <div className="empty-state">Chưa có dữ liệu cầu thủ.</div>
          )}
        </div>
      </section>
      <StandingsPanel groups={data.standings || []} activeGroup={team.group} onGroupChange={() => {}} />
    </main>
  );
}
