import { useQuery } from '@tanstack/react-query'
import { useMemo } from 'react'
import { formOptionsService } from '@/services/formOption_service'
import type { StandardOption, AssessmentOption } from '@/dtos/formOption_dto'

export const formOptionKeys = {
    all: ['form-options'] as const,
}

export const useFormOptions = () => {
    const query = useQuery({
        queryKey: formOptionKeys.all,
        queryFn:  formOptionsService.getAll,
        staleTime: 1000 * 60 * 5,
    })

    const services      = useMemo(() => query.data?.data?.services       ?? [], [query.data])
    const documentTypes = useMemo(() => query.data?.data?.document_types ?? [], [query.data])

    const getStandards = (serviceId: string): StandardOption[] =>
        services.find(s => s.id === serviceId)?.standards ?? []

    const getAssessments = (serviceId: string, standardId: string): AssessmentOption[] =>
        getStandards(serviceId).find(s => s.id === standardId)?.assessments ?? []

    return {
        services,
        documentTypes,
        getStandards,
        getAssessments,
        isLoading: query.isLoading,
        isError:   query.isError,
    }
}