import type { Question, Theory } from '../types'
import { AnswerBadge } from './AnswerBadge'

export function History({
  questions,
  theories,
}: {
  questions: Question[]
  theories: Theory[]
}) {
  const answered = questions.filter((q) => q.status === 'respondida')
  const pendingMine = questions.filter((q) => q.status === 'pendente' && q.mine)
  const theoryItems = theories.filter((t) => t.status !== 'pendente' || t.mine)

  if (answered.length === 0 && pendingMine.length === 0 && theoryItems.length === 0) {
    return <p className="empty">Nenhuma pergunta ainda. A investigação começa aqui.</p>
  }

  return (
    <ul className="history">
      {answered.map((q) => (
        <li key={q.id} className="history-item">
          <div>
            <p className="history-meta">{q.askerNickname}{q.kind === 'oral' ? ' · oral' : ''}</p>
            <p>{q.text}</p>
          </div>
          <AnswerBadge answer={q.answer} />
        </li>
      ))}
      {pendingMine.map((q) => (
        <li key={q.id} className="history-item pending">
          <div>
            <p className="history-meta">Sua pergunta · aguardando</p>
            <p>{q.text}</p>
          </div>
          <AnswerBadge />
        </li>
      ))}
      {theoryItems.map((t) => (
        <li key={t.id} className="history-item theory">
          <div>
            <p className="history-meta">
              Teoria de {t.authorNickname}
              {t.status === 'continuada' ? ' · continuar' : t.status === 'aceita' ? ' · aceita' : ' · pendente'}
            </p>
            <p>{t.text}</p>
            {t.guidance ? <p className="guidance">Orientação: {t.guidance}</p> : null}
          </div>
        </li>
      ))}
    </ul>
  )
}
