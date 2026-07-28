import { RouterProvider } from 'react-router'
import { Toaster } from 'sonner'
import { router } from './routes'

function App() {
  return (
    <>
      <RouterProvider router={router} />
      <Toaster
        position="top-right"
        richColors
        closeButton
        toastOptions={{
          style: {
            borderRadius: '16px',
            padding: '14px 20px',
            fontSize: '14px',
            fontFamily: "'Geist Variable', sans-serif",
            boxShadow: '0 8px 32px rgba(107, 95, 174, 0.15), 0 2px 8px rgba(0,0,0,0.06)',
            border: '1px solid rgba(107, 95, 174, 0.1)',
          },
          classNames: {
            success: 'toast-success',
            error: 'toast-error',
            info: 'toast-info',
          },
        }}
      />
    </>
  )
}

export default App