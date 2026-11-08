import { useState } from 'react'
import type { FormEvent } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { joinRoom } from '../api'

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
    <main className="home narrow">
      <header className="home-hero">
        <h1 className="brand-title">Última Pista</h1>
        <p className="lede">Escolha um apelido para entrar na sala {code || ''}.</p>
      </header>
      <form className="panel" onSubmit={onSubmit}>
        {!paramCode ? (
          <label>
            Código
            <input value={code} onChange={(e) => setCode(e.target.value.toUpperCase())} required maxLength={8} />
          </label>
        ) : (
          <p className="room-code-lg">Sala {code}</p>
        )}
        <label>
          Apelido
          <input value={nickname} onChange={(e) => setNickname(e.target.value)} required maxLength={24} />
        </label>
        <button className="btn btn-primary" disabled={busy}>
          Entrar na sala
        </button>
      </form>
      {error ? <p className="error-banner">{error}</p> : null}
      <p className="home-foot">
        <Link to="/">Voltar ao início</Link>
      </p>
    </main>
  )
}
