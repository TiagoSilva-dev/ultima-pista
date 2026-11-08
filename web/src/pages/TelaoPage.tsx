import { useCallback, useEffect, useState } from 'react'
import { QRCodeSVG } from 'qrcode.react'
import { Link, useParams } from 'react-router-dom'
import { CaseCard } from '../components/CaseCard'
import { Participants } from '../components/Participants'
import { QuestionList } from '../components/QuestionList'
import { RoomHeader } from '../components/RoomHeader'
import { Sheet } from '../components/Sheet'
import { useRoom } from '../hooks/useRoom'

const QR_BG = '#efe6d4'
const QR_FG = '#1c1816'

export function TelaoPage() {
  const { code = '' } = useParams()
  const { room, error, connection } = useRoom({ code, telao: true })
  const [fs, setFs] = useState(false)

  useEffect(() => {
    const onChange = () => setFs(!!document.fullscreenElement)
    document.addEventListener('fullscreenchange', onChange)
    return () => document.removeEventListener('fullscreenchange', onChange)
  }, [])

  const toggleFs = useCallback(async () => {
    try {
      if (!document.fullscreenElement) await document.documentElement.requestFullscreen()
      else await document.exitFullscreen()
    } catch {
      setFs(false)
    }
  }, [])

  const invite = typeof window !== 'undefined' ? `${window.location.origin}/entrar/${code}` : ''

  const header = (
    <RoomHeader
      code={room?.code || code}
      role="Telão"
      connection={connection}
      actions={
        <button type="button" className="btn btn-invite" onClick={() => void toggleFs()}>
          {fs ? 'Sair da tela cheia' : 'Tela cheia'}
        </button>
      }
    />
  )

  if (error || !room) {
    return (
      <Sheet header={header} className="sheet-narrow">
        {error ? (
          <div className="notice">
            <h1 className="notice-title">Telão indisponível</h1>
            <p>{error}</p>
            <Link className="btn btn-primary" to="/">
              Início
            </Link>
          </div>
        ) : (
          <p className="loading">Carregando telão…</p>
        )}
      </Sheet>
    )
  }

  const round = room.round

  if (!round) {
    return (
      <Sheet header={header} className="sheet-telao">
        <div className="telao-lobby">
          <div className="qr-card">
            <QRCodeSVG value={invite} size={220} bgColor={QR_BG} fgColor={QR_FG} />
          </div>
          <div>
            <span className="tag">Entre como detetive</span>
            <p className="lobby-code xl">{room.code}</p>
            <p className="lobby-lede">Escaneie o código ou acesse o convite. O caso começa em instantes…</p>
            <Participants people={room.participants} className="room-people lg" />
          </div>
        </div>
      </Sheet>
    )
  }

  const resolved = round.status === 'resolvida'

  return (
    <Sheet header={header} className="sheet-telao">
      <div className="room-grid">
        <div className="room-case">
          {round.case ? (
            <CaseCard
              c={round.case}
              revealed={resolved ? round.revealedSolution : undefined}
              label={resolved ? 'Caso resolvido' : 'Caso'}
              flip={resolved}
            />
          ) : null}
        </div>
        <section className="room-side">
          <h2 className="side-title">Perguntas da sala</h2>
          <QuestionList questions={round.questions.slice(-8)} />
          <div className="telao-foot">
            <Participants people={room.participants} className="room-people" />
            <div className="qr-mini">
              <QRCodeSVG value={invite} size={84} bgColor={QR_BG} fgColor={QR_FG} />
            </div>
          </div>
        </section>
      </div>
    </Sheet>
  )
}
