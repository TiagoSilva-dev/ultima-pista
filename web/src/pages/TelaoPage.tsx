import { useCallback, useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Link, useParams } from 'react-router-dom'
import { CaseCard } from '../components/CaseCard'
import { ConnectionBar } from '../components/ConnectionBar'
import { History } from '../components/History'
import { useRoom } from '../hooks/useRoom'

export function TelaoPage() {
  const { code = '' } = useParams()
  const { room, error, connection } = useRoom({ code, telao: true })
  const [fs, setFs] = useState(false)

  const toggleFs = useCallback(async () => {
    try {
      if (!document.fullscreenElement) {
        await document.documentElement.requestFullscreen()
        setFs(true)
      } else {
        await document.exitFullscreen()
        setFs(false)
      }
    } catch {
      setFs(false)
    }
  }, [])

  const invite = typeof window !== 'undefined' ? `${window.location.origin}/entrar/${code}` : ''

  if (!room && !error) {
    return (
      <main className="telao">
        <ConnectionBar connection={connection} code={code} role="Telão" />
        <p className="loading">Carregando telão…</p>
      </main>
    )
  }

  if (error || !room) {
    return (
      <main className="telao">
        <ConnectionBar connection={connection} role="Telão" />
        <div className="panel">
          <h1>Telão indisponível</h1>
          <p>{error || 'Sala não encontrada'}</p>
          <Link to="/">Início</Link>
        </div>
      </main>
    )
  }

  const round = room.round

  return (
    <main className="telao">
      <ConnectionBar connection={connection} code={room.code} role="Telão" />
      <div className="telao-toolbar">
        <h1 className="brand-title sm">Última Pista</h1>
        <button type="button" className="btn btn-secondary" onClick={() => void toggleFs()}>
          {fs ? 'Sair da tela cheia' : 'Tela cheia'}
        </button>
      </div>

      {!round ? (
        <div className="telao-lobby">
          <p className="lede">Aguardando o início do caso…</p>
          <div className="qr-block">
            <QRCodeSVG value={invite} size={180} bgColor="#EEE7D7" fgColor="#141313" />
            <div>
              <p className="room-code-lg">{room.code}</p>
              <p>Escaneie para entrar como detetive</p>
            </div>
          </div>
          <ul className="people large">
            {room.participants.map((p) => (
              <li key={p.id}>
                <span className="avatar">{p.nickname.slice(0, 1)}</span>
                {p.nickname}
              </li>
            ))}
          </ul>
        </div>
      ) : round.status === 'resolvida' ? (
        <div className="telao-resolved">
          <p className="eyebrow">Caso resolvido</p>
          {round.case ? <CaseCard c={round.case} revealed={round.revealedSolution} flip /> : null}
        </div>
      ) : (
        <div className="telao-grid">
          <div className="telao-case">{round.case ? <CaseCard c={round.case} /> : null}</div>
          <div className="telao-side">
            <h2>Perguntas recentes</h2>
            <History questions={round.questions.slice(-8)} theories={[]} />
            <div className="qr-block mini">
              <QRCodeSVG value={invite} size={96} bgColor="#EEE7D7" fgColor="#141313" />
              <span className="room-code">{room.code}</span>
            </div>
          </div>
        </div>
      )}
    </main>
  )
}
