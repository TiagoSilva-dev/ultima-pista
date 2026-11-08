import type { Participant } from '../types'
import { PersonIcon } from './Icons'

export function Participants({
  people,
  showRole = false,
  className = '',
}: {
  people: Participant[]
  showRole?: boolean
  className?: string
}) {
  if (people.length === 0) return null
  return (
    <ul className={`participants ${className}`} aria-label="Participantes">
      {people.map((p) => (
        <li key={p.id}>
          <PersonIcon className="avatar" />
          <span>{p.nickname}</span>
          {showRole ? <span className="participant-role">{p.role}</span> : null}
        </li>
      ))}
    </ul>
  )
}
