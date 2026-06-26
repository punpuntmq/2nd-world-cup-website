import { ArrowLeft } from "lucide-react";
import { DetailItem } from "../components/common/DetailItem.jsx";
import { MatchTeams } from "../components/matches/MatchTeams.jsx";
import { ScoreButton } from "../components/matches/ScoreButton.jsx";
import { StandingsPanel } from "../components/standings/StandingsPanel.jsx";
import { TeamsPanel } from "../components/teams/TeamsPanel.jsx";
import { formatDateTime } from "../utils/date.js";
import { findMatch } from "../utils/matches.js";
import { NotFoundPage } from "./NotFoundPage.jsx";

export function MatchPage({ data, matchId, onNavigate }) {
  const match = findMatch(data, matchId);

  if (!match) return <NotFoundPage title="Không tìm thấy trận đấu" onNavigate={onNavigate} />;

  return (
    <main className="detail-layout">
      <TeamsPanel teams={data.teams || []} onNavigate={onNavigate} />
      <section className="detail-page">
        <button className="back-button" type="button" onClick={() => onNavigate("/")}>
          <ArrowLeft size={16} aria-hidden="true" />
          Dashboard
        </button>
        <div className={`match-detail-card ${match.statusGroup}`}>
          <div className="match-topline">
            <div>
              <p className="eyebrow">{match.stage || "World Cup"}</p>
              <h1>{match.group || "Match"}</h1>
            </div>
            <span className={`status-pill ${match.statusGroup}`}>{match.statusLabel}</span>
          </div>
          <MatchTeams match={match} onNavigate={onNavigate} large />
          <ScoreButton match={match} onNavigate={onNavigate} large />
          <div className="hero-meta">
            <span>{formatDateTime(match.kickoffUTC)}</span>
            <span>{match.matchday ? `Matchday ${match.matchday}` : "World Cup"}</span>
            <span>{match.minute}</span>
          </div>
        </div>
        <h2 className="section-title">Thông tin trận đấu</h2>
        <div className="detail-grid">
          <DetailItem label="Status" value={match.statusLabel} />
          <DetailItem label="Stage" value={match.stage || "-"} />
          <DetailItem label="Group" value={match.group || "-"} />
          <DetailItem label="Kickoff" value={formatDateTime(match.kickoffUTC)} />
        </div>
      </section>
      <StandingsPanel groups={data.standings || []} activeGroup={match.group} onGroupChange={() => {}} />
    </main>
  );
}
