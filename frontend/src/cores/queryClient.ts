import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
    defaultOptions: {
        queries: {
            staleTime: 1000 * 60 * 5,  // data stays fresh for 5 minutes
            retry: 1,                   // retry failed requests once
            refetchOnWindowFocus: false // set true if you want live-ish data
        },
    },
})