import { useEffect, useMemo, useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  answerQuestion,
  askQuestion,
  confirmSpoken,
  confirmTheory,
  continueTheory,
  getPrivate,
  idemKey,
  listCases,
  oralQuestion,
  proposeTheory,
  returnLobby,
  startRound,
} from '../api'
import { CaseCard } from '../components/CaseCard'
import { ArrowIcon, LockIcon } from '../components/Icons'
import { Participants } from '../components/Participants'
import { QuestionList } from '../components/QuestionList'
import { RoomHeader } from '../components/RoomHeader'
import { Sheet } from '../components/Sheet'
import { useRoom } from '../hooks/useRoom'
import type { CasePublic, PrivatePayload, Question, RoomSnapshot, Theory } from '../types'

const ANSWERS = [
  { value: 'sim', label: 'Sim', className: 'ans-sim' },
  { value: 'nao', label: 'Não', className: 'ans-nao' },
  { value: 'irrelevante', label: 'Irrelevante', className: 'ans-irrel' },
] as const

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

  useEffect(() => {
    if (!msg) return
    const t = window.setTimeout(() => setMsg(null), 3500)
    return () => window.clearTimeout(t)
  }, [msg])

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

  async function run(action: () => Promise<unknown>, fallback: string) {
    setSending(true)
    setMsg(null)
    try {
      await action()
      return true
    } catch (e) {
      setMsg(e instanceof Error ? e.message : fallback)
      return false
    } finally {
      setSending(false)
    }
  }

  async function onStart(caseId: string) {
    await run(async () => setRoom(await startRound(code, caseId)), 'Falha ao iniciar')
  }

  async function onAsk(e: FormEvent) {
    e.preventDefault()
    const text = question.trim()
    if (!text) return
    const ok = await run(async () => {
      await askQuestion(code, text, idemKey())
      setQuestion('')
      await refresh('ask')
    }, 'Falha no envio')
    if (ok) setMsg('Pergunta enviada ao narrador.')
  }

  async function onTheory(e: FormEvent) {
    e.preventDefault()
    const text = theoryText.trim()
    if (!text) return
    const ok = await run(async () => {
      await proposeTheory(code, text, idemKey())
      setTheoryText('')
      setTheoryOpen(false)
      await refresh('theory')
    }, 'Falha ao enviar teoria')
    if (ok) setMsg('Teoria enviada ao narrador.')
  }

  if (!room) {
    return (
      <Sheet header={<RoomHeader code={code} connection={connection} />} className="sheet-narrow">
        {error ? (
          <div className="notice">
            <h1 className="notice-title">Sala indisponível</h1>
            <p>{error}</p>
            <Link className="btn btn-primary" to="/">
              Voltar ao início
            </Link>
          </div>
        ) : (
          <p className="loading">Abrindo a sala…</p>
        )}
      </Sheet>
    )
  }

  if (!room.you) {
    return (
      <Sheet header={<RoomHeader code={room.code} connection={connection} />} className="sheet-narrow">
        <div className="notice">
          <h1 className="notice-title">Entre na sala</h1>
          <p>Esta sessão ainda não está na sala {room.code}.</p>
          <Link className="btn btn-primary" to={`/entrar/${room.code}`}>
            Escolher apelido
          </Link>
        </div>
      </Sheet>
    )
  }

  const header = (
    <RoomHeader
      code={room.code}
      role={isNarrator ? 'Narrador' : 'Detetive'}
      connection={connection}
      onCopy={copyInvite}
      copied={copied}
    />
  )

  const toast = (
    <>
      {msg ? (
        <p className="toast" role="status">
          {msg}
        </p>
      ) : null}
      {error ? (
        <p className="toast toast-error" role="alert">
          {error}
        </p>
      ) : null}
    </>
  )

  if (!inRound) {
    return (
      <Sheet header={header}>
        <LobbyView room={room} cases={cases} isNarrator={isNarrator} onStart={onStart} sending={sending} />
        {toast}
      </Sheet>
    )
  }

  if (resolved) {
    return (
      <Sheet header={header}>
        <ResolvedView
          room={room}
          isNarrator={isNarrator}
          onLobby={() => run(async () => setRoom(await returnLobby(code)), 'Falha ao voltar ao lobby')}
        />
        {toast}
      </Sheet>
    )
  }

  if (isNarrator) {
    return (
      <Sheet header={header} className="sheet-narrator">
        <NarratorView
          room={room}
          priv={priv}
          showSolution={showSolution}
          setShowSolution={setShowSolution}
          pendingQ={pendingQ}
          pendingT={pendingT}
          sending={sending}
          run={run}
          code={code}
          refresh={refresh}
        />
        {toast}
      </Sheet>
    )
  }

  const round = room.round!
  return (
    <Sheet
      header={header}
      className="sheet-detective"
      footer={
        <Composer
          question={question}
          setQuestion={setQuestion}
          onAsk={onAsk}
          theoryOpen={theoryOpen}
          setTheoryOpen={setTheoryOpen}
          theoryText={theoryText}
          setTheoryText={setTheoryText}
          onTheory={onTheory}
          sending={sending}
        />
      }
    >
      <div className="room-grid">
        <div className="room-case">{round.case ? <CaseCard c={round.case} /> : null}</div>
        <section className="room-side" aria-labelledby="perguntas-title">
          <h2 id="perguntas-title" className="side-title">
            Perguntas da sala
          </h2>
          <QuestionList questions={round.questions} theories={round.theories} />
          <Participants people={room.participants} className="room-people" />
        </section>
      </div>
      {toast}
    </Sheet>
  )
}

function LobbyView({
  room,
  cases,
  isNarrator,
  onStart,
  sending,
}: {
  room: RoomSnapshot
  cases: CasePublic[]
  isNarrator: boolean
  onStart: (id: string) => void
  sending: boolean
}) {
  return (
    <div className="lobby">
      <section className="lobby-invite">
        <span className="tag">Convite</span>
        <p className="lobby-code">{room.code}</p>
        <p className="lobby-lede">
          Modo {room.mode === 'presencial' ? 'presencial' : 'à distância'}. Compartilhe o convite e comece quando
          quiser.
        </p>
        <a className="btn btn-outline" href={`/telao/${room.code}`} target="_blank" rel="noreferrer">
          Abrir telão
        </a>

        <h2 className="side-title sm">Na sala</h2>
        {room.participants.length === 0 ? (
          <p className="empty">Lobby vazio.</p>
        ) : (
          <Participants people={room.participants} showRole className="stacked" />
        )}
        <p className="hint">
          {room.mode === 'presencial'
            ? 'No presencial, você pode iniciar só com o narrador conectado.'
            : 'À distância, convide pelo menos um detetive pelo link.'}
        </p>
      </section>

      <section className="lobby-cases">
        {isNarrator ? (
          <>
            <h2 className="side-title">Escolha o caso</h2>
            <ul className="case-list">
              {cases.map((c) => (
                <li key={c.id}>
                  <div>
                    <strong>{c.title}</strong>
                    <p className="case-list-meta">
                      {c.theme} · {c.difficulty}
                    </p>
                  </div>
                  <button type="button" className="btn btn-primary" disabled={sending} onClick={() => onStart(c.id)}>
                    Iniciar
                  </button>
                </li>
              ))}
            </ul>
          </>
        ) : (
          <div className="waiting">
            <span className="tag">Aguarde</span>
            <p>O narrador está escolhendo o caso…</p>
          </div>
        )}
      </section>
    </div>
  )
}

function Composer({
  question,
  setQuestion,
  onAsk,
  theoryOpen,
  setTheoryOpen,
  theoryText,
  setTheoryText,
  onTheory,
  sending,
}: {
  question: string
  setQuestion: (v: string) => void
  onAsk: (e: FormEvent) => void
  theoryOpen: boolean
  setTheoryOpen: (v: boolean) => void
  theoryText: string
  setTheoryText: (v: string) => void
  onTheory: (e: FormEvent) => void
  sending: boolean
}) {
  if (theoryOpen) {
    return (
      <form className="composer composer-theory" onSubmit={onTheory}>
        <label className="sr-only" htmlFor="theory">
          Teoria
        </label>
        <textarea
          id="theory"
          value={theoryText}
          onChange={(e) => setTheoryText(e.target.value)}
          maxLength={2000}
          required
          autoFocus
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
    )
  }

  return (
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
        placeholder="Faça uma pergunta..."
        enterKeyHint="send"
        autoComplete="off"
      />
      <button className="btn btn-send" aria-label="Enviar pergunta" disabled={sending || !question.trim()}>
        <ArrowIcon className="icon" />
      </button>
      <span className="composer-divider" aria-hidden="true" />
      <button type="button" className="btn btn-outline btn-theory" onClick={() => setTheoryOpen(true)}>
        Propor teoria
      </button>
    </form>
  )
}

function NarratorView({
  room,
  priv,
  showSolution,
  setShowSolution,
  pendingQ,
  pendingT,
  sending,
  run,
  code,
  refresh,
}: {
  room: RoomSnapshot
  priv: PrivatePayload | null
  showSolution: boolean
  setShowSolution: (v: boolean) => void
  pendingQ: Question[]
  pendingT: Theory[]
  sending: boolean
  run: (action: () => Promise<unknown>, fallback: string) => Promise<boolean>
  code: string
  refresh: (reason?: string) => Promise<void>
}) {
  const round = room.round!
  const current = pendingQ[0]
  const [guidance, setGuidance] = useState<Record<string, string>>({})
  const [oralText, setOralText] = useState('')
  const [oralAnswer, setOralAnswer] = useState<string>('sim')
  const [confirmOpen, setConfirmOpen] = useState(false)
  const [spokenSummary, setSpokenSummary] = useState('')

  return (
    <div className="narrator-grid">
      <div className="narrator-controls">
        <section className="secret" aria-labelledby="secret-title">
          <LockIcon className="secret-lock" />
          <h2 id="secret-title" className="secret-title">
            Solução em segredo
          </h2>
          <p className="secret-sub">Só você pode ver</p>
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
              <button type="button" className="btn btn-cream" onClick={() => setShowSolution(false)}>
                Ocultar solução
              </button>
            </div>
          ) : (
            <button type="button" className="btn btn-cream" disabled={!priv} onClick={() => setShowSolution(true)}>
              Ver solução
            </button>
          )}
        </section>

        <hr className="rule" />

        <section className="current-q" aria-live="polite">
          {current ? (
            <>
              <span className="tag">Pergunta de {current.askerNickname}</span>
              <p className="big-q">{current.text}</p>
              <div className="answer-stack">
                {ANSWERS.map((a) => (
                  <button
                    key={a.value}
                    type="button"
                    className={`btn btn-answer ${a.className}`}
                    disabled={sending}
                    onClick={() =>
                      run(async () => {
                        await answerQuestion(code, current.id, a.value)
                        await refresh()
                      }, 'Falha ao responder')
                    }
                  >
                    {a.label}
                  </button>
                ))}
              </div>
              {pendingQ.length > 1 ? <p className="hint">+{pendingQ.length - 1} na fila</p> : null}
            </>
          ) : (
            <>
              <span className="tag">Fila de perguntas</span>
              <p className="empty">Fila vazia. Aguarde perguntas ou registre uma oral.</p>
            </>
          )}
        </section>

        {pendingT.map((t) => (
          <section key={t.id} className="theory-card">
            <span className="tag">Teoria de {t.authorNickname}</span>
            <p className="theory-text">{t.text}</p>
            <label className="field">
              Orientação (opcional)
              <input
                value={guidance[t.id] ?? ''}
                onChange={(e) => setGuidance((g) => ({ ...g, [t.id]: e.target.value }))}
                maxLength={280}
              />
            </label>
            <div className="row-actions">
              <button
                type="button"
                className="btn btn-outline"
                disabled={sending}
                onClick={() =>
                  run(async () => {
                    await continueTheory(code, t.id, guidance[t.id] ?? '')
                    setGuidance((g) => ({ ...g, [t.id]: '' }))
                    await refresh()
                  }, 'Erro ao responder teoria')
                }
              >
                Continuar investigação
              </button>
              <button
                type="button"
                className="btn btn-primary"
                disabled={sending}
                onClick={() =>
                  run(async () => {
                    await confirmTheory(code, t.id)
                    await refresh()
                  }, 'Erro ao aceitar teoria')
                }
              >
                Aceitar e revelar
              </button>
            </div>
          </section>
        ))}

        <hr className="rule" />

        {!confirmOpen ? (
          <button type="button" className="btn btn-primary btn-block btn-lg" onClick={() => setConfirmOpen(true)}>
            Confirmar solução
          </button>
        ) : (
          <section className="confirm-box">
            <p>Revelar a solução para todos? Use quando a teoria falada estiver correta.</p>
            <label className="field">
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
                onClick={() =>
                  run(async () => {
                    await confirmSpoken(code, spokenSummary)
                    await refresh()
                  }, 'Erro ao revelar')
                }
              >
                Revelar agora
              </button>
            </div>
          </section>
        )}

        <details className="fold">
          <summary>Registrar pergunta oral</summary>
          <p className="hint">
            {room.mode === 'presencial'
              ? 'Opcional — a conversa pode seguir só na voz.'
              : 'Também útil se alguém perguntar fora do app.'}
          </p>
          <label className="field">
            Resumo da pergunta
            <input value={oralText} onChange={(e) => setOralText(e.target.value)} maxLength={280} />
          </label>
          <div className="answer-row" role="radiogroup" aria-label="Resposta">
            {ANSWERS.map((a) => (
              <button
                key={a.value}
                type="button"
                role="radio"
                aria-checked={oralAnswer === a.value}
                className={`btn btn-answer sm ${a.className} ${oralAnswer === a.value ? 'is-selected' : ''}`}
                onClick={() => setOralAnswer(a.value)}
              >
                {a.label}
              </button>
            ))}
          </div>
          <button
            type="button"
            className="btn btn-outline btn-block"
            disabled={sending || !oralText.trim()}
            onClick={() =>
              run(async () => {
                await oralQuestion(code, oralText, oralAnswer, idemKey())
                setOralText('')
                await refresh()
              }, 'Erro ao registrar')
            }
          >
            Publicar com resposta
          </button>
        </details>
      </div>

      <div className="narrator-info">
        {round.case ? <CaseCard c={round.case} /> : null}
        <section className="room-side" aria-labelledby="hist-title">
          <h2 id="hist-title" className="side-title">
            Perguntas da sala
          </h2>
          <QuestionList questions={round.questions} theories={round.theories} />
          <Participants people={room.participants} className="room-people" />
        </section>
      </div>
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
  onLobby: () => Promise<unknown>
}) {
  const round = room.round!
  return (
    <div className="room-grid resolved">
      <div className="room-case">
        {round.case ? <CaseCard c={round.case} revealed={round.revealedSolution} label="Caso resolvido" flip /> : null}
      </div>
      <section className="room-side">
        <h2 className="side-title">Caso encerrado</h2>
        {round.acceptedTheory ? (
          <div className="accepted">
            <span className="tag">Teoria aceita</span>
            <p>{round.acceptedTheory}</p>
          </div>
        ) : null}
        <QuestionList questions={round.questions} theories={[]} emptyText="Resolvido sem perguntas!" />
        {isNarrator ? (
          <button type="button" className="btn btn-primary btn-block btn-lg" onClick={() => void onLobby()}>
            Escolher outro caso
          </button>
        ) : (
          <p className="hint">Aguardando o narrador para a próxima rodada.</p>
        )}
      </section>
    </div>
  )
}
