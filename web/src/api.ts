import type { CasePublic, PrivatePayload, RoomSnapshot } from './types'

export class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function parse<T>(res: Response): Promise<T> {
  const data = await res.json().catch(() => ({}))
  if (!res.ok) {
    throw new ApiError(res.status, (data as { error?: string }).error || res.statusText)
  }
  return data as T
}

export function idemKey(): string {
  if (typeof crypto !== 'undefined' && crypto.randomUUID) return crypto.randomUUID()
  return `${Date.now()}-${Math.random().toString(16).slice(2)}`
}

export async function listCases(): Promise<CasePublic[]> {
  return parse(await fetch('/api/cases'))
}

export async function createRoom(nickname: string, mode: string): Promise<RoomSnapshot> {
  return parse(
    await fetch('/api/rooms', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nickname, mode }),
    }),
  )
}

export async function joinRoom(code: string, nickname: string): Promise<RoomSnapshot> {
  return parse(
    await fetch('/api/rooms/join', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, nickname }),
    }),
  )
}

export async function getRoom(code: string, telao = false): Promise<RoomSnapshot> {
  const path = telao ? `/api/rooms/${code}/telao` : `/api/rooms/${code}`
  return parse(await fetch(path))
}

export async function getPrivate(code: string): Promise<PrivatePayload> {
  return parse(await fetch(`/api/rooms/${code}/private`))
}

export async function startRound(code: string, caseId: string): Promise<RoomSnapshot> {
  return parse(
    await fetch(`/api/rooms/${code}/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ caseId }),
    }),
  )
}

export async function askQuestion(code: string, text: string, key: string) {
  return parse(
    await fetch(`/api/rooms/${code}/questions`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key },
      body: JSON.stringify({ text, idempotencyKey: key }),
    }),
  )
}

export async function oralQuestion(code: string, text: string, answer: string, key: string) {
  return parse(
    await fetch(`/api/rooms/${code}/oral`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key },
      body: JSON.stringify({ text, answer, idempotencyKey: key }),
    }),
  )
}

export async function answerQuestion(code: string, questionId: string, answer: string) {
  return parse(
    await fetch(`/api/rooms/${code}/questions/${questionId}/answer`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ answer }),
    }),
  )
}

export async function proposeTheory(code: string, text: string, key: string) {
  return parse(
    await fetch(`/api/rooms/${code}/theories`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', 'Idempotency-Key': key },
      body: JSON.stringify({ text, idempotencyKey: key }),
    }),
  )
}

export async function continueTheory(code: string, theoryId: string, guidance: string) {
  return parse(
    await fetch(`/api/rooms/${code}/theories/${theoryId}/continue`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ guidance }),
    }),
  )
}

export async function confirmTheory(code: string, theoryId: string) {
  return parse(
    await fetch(`/api/rooms/${code}/theories/${theoryId}/confirm`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    }),
  )
}

export async function confirmSpoken(code: string, summary: string) {
  return parse(
    await fetch(`/api/rooms/${code}/confirm-solution`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ summary }),
    }),
  )
}

export async function returnLobby(code: string): Promise<RoomSnapshot> {
  return parse(
    await fetch(`/api/rooms/${code}/lobby`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: '{}',
    }),
  )
}
