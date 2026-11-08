export type CasePublic = {
  id: string
  title: string
  theme: string
  difficulty: string
  publicText?: string
  illustration?: string
}

export type Participant = {
  id: string
  nickname: string
  role: string
}

export type Question = {
  id: string
  askerNickname: string
  text: string
  kind: string
  status: string
  answer?: string
  createdAt: string
  mine?: boolean
}

export type Theory = {
  id: string
  authorNickname: string
  text: string
  status: string
  guidance?: string
  createdAt: string
  mine?: boolean
}

export type Round = {
  id: string
  number: number
  status: string
  case?: CasePublic
  startedAt: string
  resolvedAt?: string
  questions: Question[]
  theories: Theory[]
  acceptedTheory?: string
  revealedSolution?: string
}

export type YouInfo = {
  participantId: string
  nickname: string
  role: string
  isNarrator: boolean
}

export type RoomSnapshot = {
  code: string
  mode: string
  version: number
  participants: Participant[]
  round?: Round | null
  you?: YouInfo | null
  invitePath: string
  telaoPath: string
  expired?: boolean
}

export type PrivatePayload = {
  solution: string
  essentials: string[]
  hints: string[]
  caseTitle: string
}
