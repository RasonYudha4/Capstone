// hooks/useAgentCommandExecutor.ts
import { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { useFilterStore } from '@/stores/filterStore'
import type { AgentCommand } from '@/cores/types'

export function useAgentCommandExecutor() {
    const navigate = useNavigate()

    return useCallback((cmd: AgentCommand) => {
        console.log('[executor] received command:', cmd)
        const store = useFilterStore.getState()

        switch (cmd.type) {
            case 'NAVIGATE':
                console.log('[executor] navigating to', cmd.path)
                navigate(cmd.path)
                break
            case 'SET_SERVICE_FILTER':
                console.log('[executor] setting serviceId', cmd.serviceId)
                store.setServiceId(cmd.serviceId)
                break
            case 'SET_STANDARD_FILTER':
                console.log('[executor] setting standardId', cmd.standardId)
                store.setStandardId(cmd.standardId)
                break
            case 'SET_ASSESSMENT_FILTER':
                console.log('[executor] setting assessmentId', cmd.assessmentId)
                store.setAssessmentId(cmd.assessmentId)
                break
            case 'OPEN_DOCUMENT':
                console.log('[executor] setting pendingHighlight', cmd.documentId)
                store.setPendingHighlight(cmd.documentId)
                break
        }
    }, [navigate])
}