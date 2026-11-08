import { useState } from 'react'
import type { FormEvent } from 'react'
import { useNavigate } from 'react-router-dom'
import { createRoom, joinRoom } from '../api'
import { BrandHero } from '../components/BrandHero'

export function HomePage() {
  const nav = useNavigate()
  const [nickname, setNickname] = useState('')
  const [mode, setMode] = useState<'presencial' | 'remoto'>('remoto')
  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function submit(action: () => Promise<{ code: string }>, fallback: string) {
    setBusy(true)
    setError(null)
    try {
      const room = await action()
      nav(`/sala/${room.code}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : fallback)
    } finally {
      setBusy(false)
    }
  }

  function onCreate(e: FormEvent) {
    e.preventDefault()
    void submit(() => createRoom(nickname.trim(), mode), 'Não foi possível criar a sala')
  }

  function onJoin(e: FormEvent) {
    e.preventDefault()
    void submit(() => joinRoom(code.trim().toUpperCase(), nickname.trim()), 'Não foi possível entrar')
  }

  return (
    <div className="stage stage-home">
      <BrandHero />

      <p className="home-lede">
        Uma pessoa narra. As outras perguntam. Reconstruam a história — sem conta, só um link.
      </p>

      <label className="home-nick">
        <span className="tag">Seu apelido</span>
        <input
          value={nickname}
          onChange={(e) => setNickname(e.target.value)}
          maxLength={24}
          required
          autoComplete="nickname"
          placeholder="Como vão te chamar?"
        />
      </label>

      <div className="home-panels">
        <section className="home-col">
          <span className="tag tag-label">Narrador</span>
          <form className="paper-panel" onSubmit={onCreate}>
            <h2 className="panel-title">Criar sala</h2>
            <p className="panel-lede">Você escolhe o caso e responde às perguntas.</p>
            <fieldset className="segmented">
              <legend>Como vão jogar?</legend>
              <label className={mode === 'remoto' ? 'is-on' : ''}>
                <input type="radio" name="mode" checked={mode === 'remoto'} onChange={() => setMode('remoto')} />
                À distância
              </label>
              <label className={mode === 'presencial' ? 'is-on' : ''}>
                <input
                  type="radio"
                  name="mode"
                  checked={mode === 'presencial'}
                  onChange={() => setMode('presencial')}
                />
                Presencial
              </label>
            </fieldset>
            <button className="btn btn-primary btn-block btn-lg" disabled={busy || !nickname.trim()}>
              Criar sala
            </button>
          </form>
        </section>

        <section className="home-col">
          <span className="tag tag-label">Detetive</span>
          <form className="paper-panel" onSubmit={onJoin}>
            <h2 className="panel-title">Entrar com código</h2>
            <p className="panel-lede">Recebeu um convite? Digite o código da sala.</p>
            <label className="field">
              Código da sala
              <input
                className="code-input"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
                maxLength={8}
                required
                autoCapitalize="characters"
                placeholder="7K9P"
              />
            </label>
            <button className="btn btn-outline btn-block btn-lg" disabled={busy || !nickname.trim() || !code.trim()}>
              Entrar
            </button>
          </form>
        </section>
      </div>

      {error ? (
        <p className="error-banner" role="alert">
          {error}
        </p>
      ) : null}
      <p className="home-foot">O telão (modo projeção) pode ser aberto de dentro da sala.</p>
    </div>
  )
}
