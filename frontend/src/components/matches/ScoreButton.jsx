export function ScoreButton({ match, onNavigate, large }) {
  return (
    <button
      type="button"
      className={large ? "score-button large" : "score-button"}
      onClick={() => onNavigate(`/match/${match.id}`)}
      title="Xem chi tiết trận đấu"
    >
      <span>{match.homeScore}</span>
      <span className="score-separator">-</span>
      <span>{match.awayScore}</span>
    </button>
  );
}
