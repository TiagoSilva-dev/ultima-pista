import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { CopyIcon } from './Icons'

type Connection = 'connecting' | 'live' | 'reconnecting'

const CONN_LABEL: Record<Connection, string> = {
  live: 'Ao vivo',
  connecting: 'Conectando…',
  reconnecting: 'Reconectando…',
}

/**
 * Burgundy bar that tops every room screen. On wide screens it shows the brand,
 * a divider and the room code; on phones it collapses to "Sala XXXX | Papel".
 */
export function RoomHeader({
  code,
  role,
  connection,
  onCopy,
  copied,
  actions,
}: {
  code?: string
  role?: string
  connection?: Connection
  onCopy?: () => void
  copied?: boolean
  actions?: ReactNode
}) {
  return (
    <header className="room-header">
      <Link to="/" className="room-header-brand">
        Última Pista
      </Link>
      {code ? (
        <div className="room-header-code">
          <span>Sala {code}</span>
          {connection ? (
            <span
              className={`conn-dot conn-${connection}`}
              role="status"
              aria-label={CONN_LABEL[connection]}
              title={CONN_LABEL[connection]}
            />
          ) : null}
        </div>
      ) : null}
      <div className="room-header-end">
        {role ? <span className="tag tag-role">{role}</span> : null}
        {actions}
        {onCopy ? (
          <button type="button" className="btn btn-invite" onClick={onCopy}>
            <CopyIcon className="icon" />
            <span>{copied ? 'Convite copiado' : 'Copiar convite'}</span>
          </button>
        ) : null}
      </div>
    </header>
  )
}
