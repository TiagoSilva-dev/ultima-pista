const LABELS: Record<string, { label: string; className: string }> = {
  sim: { label: 'Sim', className: 'badge badge-sim' },
  nao: { label: 'Não', className: 'badge badge-nao' },
  irrelevante: { label: 'Irrelevante', className: 'badge badge-irrel' },
}

export function AnswerBadge({ answer }: { answer?: string }) {
  if (!answer) return <span className="badge badge-pendente">Aguardando</span>
  const meta = LABELS[answer] || { label: answer, className: 'badge' }
  return <span className={meta.className}>{meta.label}</span>
}
