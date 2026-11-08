import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  answerQuestion,
  confirmSpoken,
  confirmTheory,
  continueTheory,
  getPrivate,
  idemKey,
  listCases,
  oralQuestion,
  askQuestion,
  proposeTheory,
  returnLobby,
  startRound,
} from '../api'
import { AnswerBadge } from '../components/AnswerBadge'
import { CaseCard } from '../components/CaseCard'
import { ConnectionBar } from '../components/ConnectionBar'
import { History } from '../components/History'
import { useRoom } from '../hooks/useRoom'
import type { CasePublic, PrivatePayload, Question, RoomSnapshot, Theory } from '../types'

export function RoomPage() {
  const { code = '' } = useParams()
  const { room, error, connection, refresh, setRoom } = useRoom({ code })
  const [cases, setCases] = useState<CasePublic[]>([])
  const [priv, setPriv] = useState<PrivatePayload | null>(null)
  const [showSolution, setShowSolution] = useState(false)
  const [copied, setCopied] = useState(false)
  const [question, setQuestion] = useState('')
  const [theoryOpen, setTheoryOpen] = useState(false)
  const [theoryText, setTheoryText] = useState('')
  const [oralText, setOralText] = useState('')
  const [oralAnswer, setOralAnswer] = useState('sim')
  const [guidance, setGuidance] = useState('')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [spokenSummary, setSpokenSummary] = useState('')
  const [msg, setMsg] = useState<string | null>(null)
  const [sending, setSending] = useState(false)

  const isNarrator = !!room?.you?.isNarrator
  const inRound = !!room?.round
  const resolved = room?.round?.status === 'resolvida'
  const pendingQ = useMemo(
    () => room?.round?.questions.filter((q) => q.status === 'pendente') || [],
    [room],
  )
  const pendingT = useMemo(
    () => room?.round?.theories.filter((t) => t.status === 'pendente') || [],
    [room],
  )

  useEffect(() => {
    void listCases().then(setCases).catch(() => undefined)
  }, [])

  useEffect(() => {
    if (!isNarrator || !inRound || resolved) {
      return
    }
    let cancelled = false
    void getPrivate(code)
      .then((p) => {
        if (!cancelled) {
          setPriv(p)
          setShowSolution(false)
        }
      })
      .catch(() => {
        if (!cancelled) setPriv(null)
      })
    return () => {
      cancelled = true
    }
  }, [isNarrator, inRound, resolved, code, room?.round?.id])

  async function copyInvite() {
    const url = `${window.location.origin}/entrar/${code}`
    try {
      await navigator.clipboard.writeText(url)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      setMsg(url)
    }
  }

  async function onStart(caseId: string) {
    setSending(true)
    setMsg(null)
    try {
      const snap = await startRound(code, caseId)
      setRoom(snap)
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Falha ao iniciar')
    } finally {
      setSending(false)
    }
  }

  async function onAsk(e: FormEvent) {
    e.preventDefault()
    const text = question
    const key = idemKey()
    setSending(true)
    setMsg(null)
    try {
      await askQuestion(code, text, key)
      setQuestion('')
      setMsg('Pergunta enviada.')
      await refresh('ask')
    } catch (err) {
      setQuestion(text)
      setMsg(err instanceof Error ? err.message : 'Falha no envio')
    } finally {
      setSending(false)
    }
  }

  async function onTheory(e: FormEvent) {
    e.preventDefault()
    const text = theoryText
    const key = idemKey()
    setSending(true)
    try {
      await proposeTheory(code, text, key)
      setTheoryText('')
      setTheoryOpen(false)
      setMsg('Teoria enviada ao narrador.')
      await refresh('theory')
    } catch (err) {
      setTheoryText(text)
      setMsg(err instanceof Error ? err.message : 'Falha ao enviar teoria')
    } finally {
      setSending(false)
    }
  }

  if (!room && !error) {
    return (
      <main className="page">
        <ConnectionBar connection={connection} code={code} />
        <p className="loading">Abrindo a sala…</p>
      </main>
    )
  }

  if (error && !room) {
    return (
      <main className="page">
        <ConnectionBar connection={connection} />
        <div className="panel">
          <h1>Sala indisponível</h1>
          <p>{error}</p>
          <Link className="btn btn-primary" to="/">
            Voltar ao início
          </Link>
        </div>
      </main>
    )
  }

  if (!room) return null

  if (!room.you) {
    return (
      <main className="page">
        <ConnectionBar connection={connection} code={room.code} />
        <div className="panel">
          <h1>Entre na sala</h1>
          <p>Esta sessão ainda não está na sala {room.code}.</p>
          <Link className="btn btn-primary" to={`/entrar/${room.code}`}>
            Escolher apelido
          </Link>
        </div>
      </main>
    )
  }

  return (
    <main className={`page room-page ${isNarrator ? 'narrator' : 'detective'}`}>
      <ConnectionBar
        connection={connection}
        code={room.code}
        role={isNarrator ? 'Narrador' : 'Detetive'}
      />

      {!inRound ? (
        <LobbyView
          room={room}
          cases={cases}
          isNarrator={isNarrator}
          copied={copied}
          onCopy={copyInvite}
          onStart={onStart}
          sending={sending}
        />
      ) : resolved ? (
        <ResolvedView
          room={room}
          isNarrator={isNarrator}
          onLobby={async () => {
            const snap = await returnLobby(code)
            setRoom(snap)
          }}
        />
      ) : isNarrator ? (
        <NarratorView
          room={room}
          priv={priv}
          showSolution={showSolution}
          setShowSolution={setShowSolution}
          pendingQ={pendingQ}
          pendingT={pendingT}
          oralText={oralText}
          setOralText={setOralText}
          oralAnswer={oralAnswer}
          setOralAnswer={setOralAnswer}
          guidance={guidance}
          setGuidance={setGuidance}
          confirmOpen={confirmOpen}
          setConfirmOpen={setConfirmOpen}
          spokenSummary={spokenSummary}
          setSpokenSummary={setSpokenSummary}
          sending={sending}
          setSending={setSending}
          setMsg={setMsg}
          code={code}
          refresh={refresh}
        />
      ) : (
        <DetectiveView
          room={room}
          question={question}
          setQuestion={setQuestion}
          onAsk={onAsk}
          theoryOpen={theoryOpen}
          setTheoryOpen={setTheoryOpen}
          theoryText={theoryText}
          setTheoryText={setTheoryText}
          onTheory={onTheory}
          sending={sending}
          onCopy={copyInvite}
          copied={copied}
        />
      )}

      {msg ? (
        <p className="toast" role="status">
          {msg}
        </p>
      ) : null}
      {error ? <p className="error-banner">{error}</p> : null}
    </main>
  )
}

function LobbyView({
  room,
  cases,
  isNarrator,
  copied,
  onCopy,
  onStart,
  sending,
}: {
  room: RoomSnapshot
  cases: CasePublic[]
  isNarrator: boolean
  copied: boolean
  onCopy: () => void
  onStart: (id: string) => void
  sending: boolean
}) {
  return (
    <div className="lobby">
      <header className="lobby-head">
        <h1 className="brand-title sm">Última Pista</h1>
        <p className="lede">Modo {room.mode}. Compartilhe o convite e comece quando quiser.</p>
        <div className="invite-row">
          <span className="room-code-lg">{room.code}</span>
          <button type="button" className="btn btn-secondary" onClick={onCopy}>
            {copied ? 'Convite copiado' : 'Copiar convite'}
          </button>
          <a className="btn btn-ghost" href={`/telao/${room.code}`} target="_blank" rel="noreferrer">
            Abrir telão
          </a>
        </div>
      </header>

      <section className="panel">
        <h2>Participantes</h2>
        {room.participants.length === 0 ? (
          <p className="empty">Lobby vazio.</p>
        ) : (
          <ul className="people">
            {room.participants.map((p) => (
              <li key={p.id}>
                <span className="avatar" aria-hidden="true">
                  {p.nickname.slice(0, 1).toUpperCase()}
                </span>
                {p.nickname}
                <span className="muted">{p.role}</span>
              </li>
            ))}
          </ul>
        )}
        {room.mode === 'presencial' ? (
          <p className="hint">No presencial, você pode iniciar só com o narrador conectado.</p>
        ) : (
          <p className="hint">À distância, convide pelo menos um detetive pelo link.</p>
        )}
      </section>

      {isNarrator ? (
        <section className="panel">
          <h2>Escolher caso</h2>
          <ul className="case-list">
            {cases.map((c) => (
              <li key={c.id}>
                <div>
                  <strong>{c.title}</strong>
                  <p>
                    {c.theme} · {c.difficulty}
                  </p>
                </div>
                <button
                  type="button"
                  className="btn btn-primary"
                  disabled={sending}
                  onClick={() => onStart(c.id)}
                >
                  Iniciar
                </button>
              </li>
            ))}
          </ul>
        </section>
      ) : (
        <p className="waiting">Aguardando o narrador escolher o caso…</p>
      )}
    </div>
  )
}

function DetectiveView({
  room,
  question,
  setQuestion,
  onAsk,
  theoryOpen,
  setTheoryOpen,
  theoryText,
  setTheoryText,
  onTheory,
  sending,
  onCopy,
  copied,
}: {
  room: RoomSnapshot
  question: string
  setQuestion: (v: string) => void
  onAsk: (e: FormEvent) => void
  theoryOpen: boolean
  setTheoryOpen: (v: boolean) => void
  theoryText: string
  setTheoryText: (v: string) => void
  onTheory: (e: FormEvent) => void
  sending: boolean
  onCopy: () => void
  copied: boolean
}) {
  const round = room.round!
  return (
    <div className="play detective-play">
      <div className="play-main">
        <div className="invite-row compact">
          <span className="room-code">Sala {room.code}</span>
          <button type="button" className="btn btn-ghost" onClick={onCopy}>
            {copied ? 'Copiado' : 'Copiar convite'}
          </button>
        </div>
        {round.case ? <CaseCard c={round.case} /> : null}
        <section>
          <h2 className="section-title">Perguntas da sala</h2>
          <History questions={round.questions} theories={round.theories} />
        </section>
      </div>

      <div className="composer-dock">
        {theoryOpen ? (
          <form className="composer theory-composer" onSubmit={onTheory}>
            <label className="sr-only" htmlFor="theory">
              Teoria
            </label>
            <textarea
              id="theory"
              value={theoryText}
              onChange={(e) => setTheoryText(e.target.value)}
              maxLength={2000}
              required
              placeholder="Explique a sequência completa…"
              rows={3}
            />
            <div className="composer-actions">
              <button type="button" className="btn btn-ghost" onClick={() => setTheoryOpen(false)}>
                Cancelar
              </button>
              <button className="btn btn-primary" disabled={sending || !theoryText.trim()}>
                Enviar teoria
              </button>
            </div>
          </form>
        ) : (
          <form className="composer" onSubmit={onAsk}>
            <label className="sr-only" htmlFor="q">
              Pergunta
            </label>
            <input
              id="q"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              maxLength={280}
              required
              placeholder="Faça uma pergunta…"
              enterKeyHint="send"
            />
            <button className="btn btn-icon" aria-label="Enviar pergunta" disabled={sending || !question.trim()}>
              →
            </button>
            <button type="button" className="btn btn-primary" onClick={() => setTheoryOpen(true)}>
              Propor teoria
            </button>
          </form>
        )}
      </div>
    </div>
  )
}

function NarratorView({
  room,
  priv,
  showSolution,
  setShowSolution,
  pendingQ,
  pendingT,
  oralText,
  setOralText,
  oralAnswer,
  setOralAnswer,
  guidance,
  setGuidance,
  confirmOpen,
  setConfirmOpen,
  spokenSummary,
  setSpokenSummary,
  sending,
  setSending,
  setMsg,
  code,
  refresh,
}: {
  room: RoomSnapshot
  priv: PrivatePayload | null
  showSolution: boolean
  setShowSolution: (v: boolean) => void
  pendingQ: Question[]
  pendingT: Theory[]
  oralText: string
  setOralText: (v: string) => void
  oralAnswer: string
  setOralAnswer: (v: string) => void
  guidance: string
  setGuidance: (v: string) => void
  confirmOpen: boolean
  setConfirmOpen: (v: boolean) => void
  spokenSummary: string
  setSpokenSummary: (v: string) => void
  sending: boolean
  setSending: (v: boolean) => void
  setMsg: (v: string | null) => void
  code: string
  refresh: (reason?: string) => Promise<void>
}) {
  const round = room.round!
  const current = pendingQ[0]

  async function answer(id: string, answer: string) {
    setSending(true)
    try {
      await answerQuestion(code, id, answer)
      await refresh()
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'Falha ao responder')
    } finally {
      setSending(false)
    }
  }

  return (
    <div className="play narrator-play">
      {round.case ? <CaseCard c={round.case} /> : null}

      <section className="secret-panel">
        <div className="secret-head">
          <span aria-hidden="true">🔒</span>
          <h2>Solução em segredo</h2>
        </div>
        {showSolution && priv ? (
          <div className="secret-body">
            <p>{priv.solution}</p>
            <h3>Fatos essenciais</h3>
            <ul>
              {priv.essentials.map((e) => (
                <li key={e}>{e}</li>
              ))}
            </ul>
            <h3>Orientações</h3>
            <ul>
              {priv.hints.map((e) => (
                <li key={e}>{e}</li>
              ))}
            </ul>
          </div>
        ) : (
          <button type="button" className="btn btn-secondary" onClick={() => setShowSolution(true)}>
            Ver solução
          </button>
        )}
      </section>

      <section className="panel current-q">
        <h2>{current ? `Pergunta de ${current.askerNickname}` : 'Fila de perguntas'}</h2>
        {current ? (
          <>
            <p className="big-q">{current.text}</p>
            <div className="answer-grid">
              <button type="button" className="btn ans sim" disabled={sending} onClick={() => answer(current.id, 'sim')}>
                ● Sim
              </button>
              <button type="button" className="btn ans nao" disabled={sending} onClick={() => answer(current.id, 'nao')}>
                ■ Não
              </button>
              <button
                type="button"
                className="btn ans irrel"
                disabled={sending}
                onClick={() => answer(current.id, 'irrelevante')}
              >
                ▲ Irrelevante
              </button>
            </div>
            {pendingQ.length > 1 ? <p className="hint">+{pendingQ.length - 1} na fila</p> : null}
          </>
        ) : (
          <p className="empty">Fila vazia. Aguarde perguntas ou registre uma oral.</p>
        )}
      </section>

      {pendingT.map((t) => (
        <section key={t.id} className="panel">
          <h2>Teoria de {t.authorNickname}</h2>
          <p>{t.text}</p>
          <label>
            Orientação (opcional)
            <input value={guidance} onChange={(e) => setGuidance(e.target.value)} maxLength={280} />
          </label>
          <div className="row-actions">
            <button
              type="button"
              className="btn btn-secondary"
              disabled={sending}
              onClick={async () => {
                setSending(true)
                try {
                  await continueTheory(code, t.id, guidance)
                  setGuidance('')
                  await refresh()
                } catch (e) {
                  setMsg(e instanceof Error ? e.message : 'Erro')
                } finally {
                  setSending(false)
                }
              }}
            >
              Continuar investigação
            </button>
            <button
              type="button"
              className="btn btn-primary"
              disabled={sending}
              onClick={async () => {
                setSending(true)
                try {
                  await confirmTheory(code, t.id)
                  await refresh()
                } catch (e) {
                  setMsg(e instanceof Error ? e.message : 'Erro')
                } finally {
                  setSending(false)
                }
              }}
            >
              Aceitar e revelar
            </button>
          </div>
        </section>
      ))}

      <section className="panel">
        <h2>Registrar pergunta oral</h2>
        <p className="hint">
          {room.mode === 'presencial'
            ? 'Opcional — a conversa pode seguir só na voz.'
            : 'Também útil se alguém perguntar fora do app.'}
        </p>
        <label>
          Resumo
          <input value={oralText} onChange={(e) => setOralText(e.target.value)} maxLength={280} />
        </label>
        <div className="answer-grid compact">
          {(['sim', 'nao', 'irrelevante'] as const).map((a) => (
            <button
              key={a}
              type="button"
              className={`btn ans ${a === 'nao' ? 'nao' : a === 'sim' ? 'sim' : 'irrel'} ${oralAnswer === a ? 'selected' : ''}`}
              onClick={() => setOralAnswer(a)}
            >
              <AnswerBadge answer={a} />
            </button>
          ))}
        </div>
        <button
          type="button"
          className="btn btn-secondary"
          disabled={sending || !oralText.trim()}
          onClick={async () => {
            setSending(true)
            try {
              await oralQuestion(code, oralText, oralAnswer, idemKey())
              setOralText('')
              await refresh()
            } catch (e) {
              setMsg(e instanceof Error ? e.message : 'Erro')
            } finally {
              setSending(false)
            }
          }}
        >
          Publicar com resposta
        </button>
      </section>

      <section className="panel danger-zone">
        <h2>Encerrar</h2>
        {!confirmOpen ? (
          <button type="button" className="btn btn-primary" onClick={() => setConfirmOpen(true)}>
            Confirmar solução
          </button>
        ) : (
          <div>
            <p>Revelar a solução para todos? Use quando a teoria falada estiver correta.</p>
            <label>
              Resumo opcional
              <input value={spokenSummary} onChange={(e) => setSpokenSummary(e.target.value)} maxLength={2000} />
            </label>
            <div className="row-actions">
              <button type="button" className="btn btn-ghost" onClick={() => setConfirmOpen(false)}>
                Cancelar
              </button>
              <button
                type="button"
                className="btn btn-primary"
                disabled={sending}
                onClick={async () => {
                  setSending(true)
                  try {
                    await confirmSpoken(code, spokenSummary)
                    await refresh()
                  } catch (e) {
                    setMsg(e instanceof Error ? e.message : 'Erro')
                  } finally {
                    setSending(false)
                  }
                }}
              >
                Revelar agora
              </button>
            </div>
          </div>
        )}
      </section>

      <section>
        <h2 className="section-title">Histórico</h2>
        <History questions={round.questions} theories={round.theories} />
      </section>
    </div>
  )
}

function ResolvedView({
  room,
  isNarrator,
  onLobby,
}: {
  room: RoomSnapshot
  isNarrator: boolean
  onLobby: () => Promise<void>
}) {
  const round = room.round!
  return (
    <div className="resolved">
      <p className="eyebrow">Caso resolvido</p>
      <h1>{round.case?.title}</h1>
      {round.case ? <CaseCard c={round.case} revealed={round.revealedSolution} flip /> : null}
      {round.acceptedTheory ? (
        <section className="panel">
          <h2>Teoria aceita</h2>
          <p>{round.acceptedTheory}</p>
        </section>
      ) : null}
      {isNarrator ? (
        <button type="button" className="btn btn-primary" onClick={() => void onLobby()}>
          Escolher outro caso
        </button>
      ) : (
        <p className="hint">Aguardando o narrador para a próxima rodada.</p>
      )}
    </div>
  )
}
