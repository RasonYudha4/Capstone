import { Suspense, type ComponentType, type ReactElement } from "react"

export default function wrap(Component: ComponentType): ReactElement {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <Component />
    </Suspense>
  )
}