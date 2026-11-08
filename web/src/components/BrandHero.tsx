export function BrandHero({ small = false }: { small?: boolean }) {
  return (
    <header className={`brand-hero ${small ? 'is-small' : ''}`}>
      <div className="brand-hero-row">
        <span className="brand-rule" aria-hidden="true" />
        <h1 className="brand-title">Última Pista</h1>
        <span className="brand-rule" aria-hidden="true" />
      </div>
      <p className="brand-sub">Um jogo de mistérios em grupo</p>
    </header>
  )
}
