import type { ReactNode } from 'react'

/** Dark backdrop + the paper "window" that holds every room screen. */
export function Sheet({
  header,
  footer,
  children,
  className = '',
}: {
  header: ReactNode
  footer?: ReactNode
  children: ReactNode
  className?: string
}) {
  return (
    <div className="stage">
      <main className={`sheet ${className}`}>
        {header}
        <div className="sheet-body">{children}</div>
        {footer ? <div className="sheet-footer">{footer}</div> : null}
      </main>
    </div>
  )
}
