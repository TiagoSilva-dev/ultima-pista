import type { CasePublic } from '../types'
import { CaseArt } from './CaseArt'

/**
 * Case card: on wide screens a lamp-lit illustration with the case written on a
 * torn sheet of paper; on phones only the paper sheet is shown.
 */
export function CaseCard({
  c,
  revealed,
  label = 'Caso',
  flip,
  showMeta = false,
}: {
  c: CasePublic
  revealed?: string
  label?: string
  flip?: boolean
  showMeta?: boolean
}) {
  return (
    <article className={`case-card ${flip ? 'case-card-flip' : ''} ${revealed ? 'is-revealed' : ''}`}>
      <div className="case-art">
        <CaseArt kind={c.illustration || 'glass'} />
      </div>
      {/* the wrapper carries the drop shadow, since clip-path would cut it off */}
      <div className="case-paper-wrap">
        <div className="case-paper">
          <span className="tag tag-stamp">{label}</span>
          <h2 className="case-title">{c.title}</h2>
          {showMeta ? (
            <p className="case-meta">
              {c.theme} · {c.difficulty}
            </p>
          ) : null}
          <p className="case-text">{revealed || c.publicText}</p>
        </div>
      </div>
    </article>
  )
}
