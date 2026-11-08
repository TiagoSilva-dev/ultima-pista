import type { CasePublic } from '../types'
import { CaseArt } from './CaseArt'

export function CaseCard({
  c,
  revealed,
  flip,
}: {
  c: CasePublic
  revealed?: string
  flip?: boolean
}) {
  return (
    <article className={`case-card ${flip ? 'case-card-flip' : ''}`}>
      <header className="case-card-head">
        <p className="eyebrow">Caso</p>
        <h2>{c.title}</h2>
        <p className="meta">
          {c.theme} · {c.difficulty}
        </p>
      </header>
      <div className="case-card-art" aria-hidden="true">
        <CaseArt kind={c.illustration || 'library'} />
      </div>
      <p className="case-card-text">{revealed || c.publicText}</p>
    </article>
  )
}
