export function ConnectionBar({
  connection,
  code,
  role,
}: {
  connection: string
  code?: string
  role?: string
}) {
  const label =
    connection === 'live' ? 'Ao vivo' : connection === 'reconnecting' ? 'Reconectando…' : 'Conectando…'
  return (
    <div className="topbar" role="banner">
      <div className="topbar-left">
        {code ? <span className="room-code">Sala {code}</span> : <span className="brand-mini">Última Pista</span>}
        {role ? <span className="role-pill">{role}</span> : null}
      </div>
      <span className={`conn conn-${connection}`} aria-live="polite">
        {label}
      </span>
    </div>
  )
}
