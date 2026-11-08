import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { createRoom, joinRoom } from '../api'

export function HomePage() {
  const nav = useNavigate()
  const [nickname, setNickname] = useState('')
  const [mode, setMode] = useState<'presencial' | 'remoto'>('remoto')
  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function onCreate(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      const room = await createRoom(nickname.trim(), mode)
      nav(`/sala/${room.code}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível criar a sala')
    } finally {
      setBusy(false)
    }
  }

  async function onJoin(e: FormEvent) {
    e.preventDefault()
    setBusy(true)
    setError(null)
    try {
      const room = await joinRoom(code.trim().toUpperCase(), nickname.trim())
      nav(`/sala/${room.code}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Não foi possível entrar')
    } finally {
      setBusy(false)
    }
  }

  return (
    <main className="home">
      <div className="home-glow" aria-hidden="true" />
      <header className="home-hero">
        <p className="eyebrow">Enigmas cooperativos</p>
        <h1 className="brand-title">Última Pista</h1>
        <p className="lede">
          Uma pessoa narra. As outras perguntam. Reconstruam a sequência — sem conta, só um link.
        </p>
      </header>

      <div className="home-panels">
        <form className="panel" onSubmit={onCreate}>
          <h2>Criar sala</h2>
          <label>
            Seu apelido
            <input
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              maxLength={24}
              required
              autoComplete="nickname"
            />
          </label>
          <fieldset>
            <legend>Como vão jogar?</legend>
            <label className="radio">
              <input
                type="radio"
                name="mode"
                checked={mode === 'remoto'}
                onChange={() => setMode('remoto')}
              />
              À distância
            </label>
            <label className="radio">
              <input
                type="radio"
                name="mode"
                checked={mode === 'presencial'}
                onChange={() => setMode('presencial')}
              />
              Presencial
            </label>
          </fieldset>
          <button className="btn btn-primary" disabled={busy || !nickname.trim()}>
            Criar sala
          </button>
        </form>

        <form className="panel" onSubmit={onJoin}>
          <h2>Entrar com código</h2>
          <label>
            Seu apelido
            <input
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              maxLength={24}
              required
              autoComplete="nickname"
            />
          </label>
          <label>
            Código da sala
            <input
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
              maxLength={8}
              required
              autoCapitalize="characters"
              placeholder="Ex.: 7K9P"
            />
          </label>
          <button className="btn btn-secondary" disabled={busy || !nickname.trim() || !code.trim()}>
            Entrar
          </button>
        </form>
      </div>

      {error ? <p className="error-banner" role="alert">{error}</p> : null}
      <p className="home-foot">
        Prefere o telão? Depois de entrar, abra o link de projeção na sala — ou use{' '}
        <Link to="/entrar">um convite direto</Link>.
      </p>
    </main>
  )
}
