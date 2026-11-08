import type { Question, Theory } from '../types'
import { AnswerBadge } from './AnswerBadge'

const THEORY_STATUS: Record<string, string> = {
  continuada: 'continuar investigando',
  aceita: 'aceita',
  pendente: 'aguardando o narrador',
}

/** The "Perguntas da sala" ledger: answered questions, my pending ones and theories. */
export function QuestionList({
  questions,
  theories = [],
  emptyText = 'Nenhuma pergunta ainda. A investigação começa aqui.',
}: {
  questions: Question[]
  theories?: Theory[]
  emptyText?: string
}) {
  const answered = questions.filter((q) => q.status === 'respondida')
  const pendingMine = questions.filter((q) => q.status === 'pendente' && q.mine)
  const theoryItems = theories.filter((t) => t.status !== 'pendente' || t.mine)

  if (answered.length === 0 && pendingMine.length === 0 && theoryItems.length === 0) {
    return <p className="empty">{emptyText}</p>
  }

  return (
    <ul className="qlist">
      {answered.map((q) => (
        <li key={q.id} className="qrow" title={`Pergunta de ${q.askerNickname}`}>
          <p className="qrow-text">
            {q.text}
            {q.kind === 'oral' ? <span className="qrow-meta"> · oral</span> : null}
          </p>
          <AnswerBadge answer={q.answer} />
        </li>
      ))}
      {pendingMine.map((q) => (
        <li key={q.id} className="qrow is-pending">
          <p className="qrow-text">{q.text}</p>
          <AnswerBadge />
        </li>
      ))}
      {theoryItems.map((t) => (
        <li key={t.id} className={`qrow is-theory theory-${t.status}`}>
          <div className="qrow-text">
            <span className="tag tag-sm">Teoria de {t.authorNickname}</span>
            <p>{t.text}</p>
            <p className="qrow-meta">{THEORY_STATUS[t.status] ?? t.status}</p>
            {t.guidance ? <p className="guidance">Orientação: {t.guidance}</p> : null}
          </div>
        </li>
      ))}
    </ul>
  )
}
