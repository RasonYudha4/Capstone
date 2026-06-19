// hooks/useAgentCommandExecutor.ts
import { useCallback } from 'react'
import { useNavigate } from 'react-router'
import { useFilterStore } from '@/stores/filterStore'
import type { AgentCommand } from '@/cores/types'

export function useAgentCommandExecutor() {
    const navigate = useNavigate()
    const { setServiceId, setStandardId, setAssessmentId } = useFilterStore()

    return useCallback((cmd: AgentCommand) => {
        switch (cmd.type) {
            case 'NAVIGATE':
                navigate(cmd.path)
                break
            case 'SET_SERVICE_FILTER':
                setServiceId(cmd.serviceId)
                break
            case 'SET_STANDARD_FILTER':
                setStandardId(cmd.standardId)
                break
            case 'SET_ASSESSMENT_FILTER':
                setAssessmentId(cmd.assessmentId)
                break
            case 'OPEN_DOCUMENT':
                // This fires a Zustand action that FileDetailModal listens to
                // useFilterStore.getState().setPendingDocumentId(cmd.documentId)
                break
            // case 'SHOW_TOAST':
            //     // imported from sonner
            //     import('sonner').then(({ toast }) => toast.info(cmd.message))
            //     break
        }
    }, [navigate, setServiceId, setStandardId, setAssessmentId])
}