import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { joinRoom } from '../api'
import { BrandHero } from '../components/BrandHero'

export function JoinPage() {
  const { code: paramCode } = useParams()
  const nav = useNavigate()
  const [nickname, setNickname] = useState('')
  const [code, setCode] = useState((paramCode || '').toUpperCase())
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function onSubmit(e: FormEvent) {
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
    <div className="stage stage-home narrow">
      <BrandHero small />
      <section className="home-col">
        <span className="tag tag-label">Detetive</span>
        <form className="paper-panel" onSubmit={onSubmit}>
          {paramCode ? (
            <>
              <span className="tag">Convite</span>
              <p className="lobby-code">Sala {code}</p>
            </>
          ) : (
            <label className="field">
              Código da sala
              <input
                className="code-input"
                value={code}
                onChange={(e) => setCode(e.target.value.toUpperCase())}
                required
                maxLength={8}
                placeholder="7K9P"
              />
            </label>
          )}
          <label className="field">
            Seu apelido
            <input
              value={nickname}
              onChange={(e) => setNickname(e.target.value)}
              required
              maxLength={24}
              autoComplete="nickname"
              autoFocus
            />
          </label>
          <button className="btn btn-primary btn-block btn-lg" disabled={busy || !nickname.trim() || !code.trim()}>
            Entrar na sala
          </button>
        </form>
      </section>
      {error ? (
        <p className="error-banner" role="alert">
          {error}
        </p>
      ) : null}
      <p className="home-foot">
        <Link to="/">Voltar ao início</Link>
      </p>
    </div>
  )
}
