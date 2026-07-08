import { ArrowLeft } from "lucide-react";

export function NotFoundPage({ title, onNavigate }) {
  return (
    <main className="detail-layout single">
      <section className="detail-page">
        <button className="back-button" type="button" onClick={() => onNavigate("/")}>
          <ArrowLeft size={16} aria-hidden="true" />
          Dashboard
        </button>
        <div className="empty-state">{title}</div>
      </section>
    </main>
  );
}
