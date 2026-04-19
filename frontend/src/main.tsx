import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { AuthProvider } from './cores/AuthContext'
import App from './App'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from './cores/queryClient'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <AuthProvider>
        <App />
      </AuthProvider>
    </QueryClientProvider>
  </StrictMode>
)