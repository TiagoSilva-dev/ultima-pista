const LABELS: Record<string, { label: string; symbol: string; className: string }> = {
  sim: { label: 'Sim', symbol: '●', className: 'badge badge-sim' },
  nao: { label: 'Não', symbol: '■', className: 'badge badge-nao' },
  irrelevante: { label: 'Irrelevante', symbol: '▲', className: 'badge badge-irrel' },
}

export function AnswerBadge({ answer }: { answer?: string }) {
  if (!answer) return <span className="badge badge-pendente">Pendente</span>
  const meta = LABELS[answer] || { label: answer, symbol: '?', className: 'badge' }
  return (
    <span className={meta.className} aria-label={meta.label}>
      <span aria-hidden="true">{meta.symbol}</span> {meta.label}
    </span>
  )
}
