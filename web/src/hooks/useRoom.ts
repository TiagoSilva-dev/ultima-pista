import { useCallback, useEffect, useRef, useState } from 'react'
import { getRoom } from '../api'
import type { RoomSnapshot } from '../types'

type Options = {
  code: string
  telao?: boolean
}

export function useRoom({ code, telao = false }: Options) {
  const [room, setRoom] = useState<RoomSnapshot | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [connection, setConnection] = useState<'connecting' | 'live' | 'reconnecting'>('connecting')
  const versionRef = useRef(0)
  const inflight = useRef(0)
  const wanted = useRef(0)

  const refresh = useCallback(
    async (_reason = 'manual') => {
      const ticket = ++wanted.current
      inflight.current += 1
      try {
        const snap = await getRoom(code, telao)
        // Always discard older snapshots (avoids lobby overwriting an active round).
        if (snap.version < versionRef.current) {
          return
        }
        // If a newer refresh was requested and this response is not newer, skip.
        if (ticket !== wanted.current && snap.version === versionRef.current) {
          return
        }
        versionRef.current = snap.version
        setRoom(snap)
        setError(snap.expired ? 'Sala expirada por inatividade (24h).' : null)
      } catch (e) {
        setError(e instanceof Error ? e.message : 'Falha ao carregar a sala')
      } finally {
        inflight.current -= 1
      }
    },
    [code, telao],
  )

  useEffect(() => {
    versionRef.current = 0
    setRoom(null)
    void refresh('init')
  }, [refresh])

  useEffect(() => {
    let es: EventSource | null = null
    let closed = false
    let retry: number | undefined

    const connect = () => {
      if (closed) return
      setConnection((c) => (c === 'live' ? 'live' : 'connecting'))
      es = new EventSource(`/api/rooms/${code}/events`)
      es.addEventListener('hello', (ev) => {
        setConnection('live')
        try {
          const data = JSON.parse((ev as MessageEvent).data) as { version: number }
          if (data.version !== versionRef.current) void refresh('hello')
          else void refresh('hello-sync')
        } catch {
          void refresh('hello')
        }
      })
      es.addEventListener('invalidate', (ev) => {
        setConnection('live')
        try {
          const data = JSON.parse((ev as MessageEvent).data) as { version: number }
          if (data.version >= versionRef.current) void refresh('sse')
        } catch {
          void refresh('sse')
        }
      })
      es.addEventListener('ping', () => setConnection('live'))
      es.onerror = () => {
        setConnection('reconnecting')
        es?.close()
        retry = window.setTimeout(connect, 1500)
      }
    }
    connect()
    return () => {
      closed = true
      if (retry) window.clearTimeout(retry)
      es?.close()
    }
  }, [code, refresh])

  return { room, error, connection, refresh, setRoom }
}
