import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { documentService, statsService } from '@/services/document_services'
import type {
    CreateDocumentBody,
    UpdateDocumentBody,
    ApprovalRequest,
    PaginationQuery,
    StatsResponse,
} from '@/dtos/document_dto'

// ─── Query Keys ───────────────────────────────────────────────────────────────

export const documentKeys = {
    all:          ['documents'] as const,
    lists:        () => [...documentKeys.all, 'list'] as const,
    list:         (filters?: Record<string, unknown>) => [...documentKeys.lists(), { filters }] as const,
    details:      () => [...documentKeys.all, 'detail'] as const,
    detail:       (id: string) => [...documentKeys.details(), id] as const,
    byType:       (type: string, query?: PaginationQuery) => [...documentKeys.all, 'type', type, query] as const,
    byGroup:      (group: string, query?: PaginationQuery) => [...documentKeys.all, 'group', group, query] as const,
    byService:    (service: string, query?: PaginationQuery) => [...documentKeys.all, 'service', service, query] as const,
    byStandard:   (standard: string, query?: PaginationQuery) => [...documentKeys.all, 'standard', standard, query] as const,
    byAssessment: (assessment: string, query?: PaginationQuery) => [...documentKeys.all, 'assessment', assessment, query] as const,
    byStatus:     (status: string, query?: PaginationQuery) => [...documentKeys.all, 'status', status, query] as const,
    mine:         (query?: PaginationQuery) => [...documentKeys.all, 'mine', query] as const,
    public:       (query?: PaginationQuery) => [...documentKeys.all, 'public', query] as const,
    publicDetail: (id: string) => [...documentKeys.all, 'public', 'detail', id] as const,
    stats:        ['documents', 'stats'] as const,
}

// ─── Queries ──────────────────────────────────────────────────────────────────

export const useDocuments = (query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.list(query),
        queryFn: () => documentService.getAll(query),
        staleTime: 1000 * 60 * 2,
        refetchInterval: 1000 * 60,
        refetchIntervalInBackground: false,
    })
}

export const useDocument = (id: string) => {
    return useQuery({
        queryKey: documentKeys.detail(id),
        queryFn: () => documentService.getById(id),
        enabled: !!id,
    })
}

// ── Public hooks (no auth required) ──────────────────────────────────────────

export const usePublicDocuments = (query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.public(query),
        queryFn: () => documentService.getPublicDocuments(query),
        staleTime: 1000 * 60 * 5,
    })
}

export const usePublicDocumentUrl = (id: string) => {
    return useQuery({
        queryKey: documentKeys.publicDetail(id),
        queryFn: () => documentService.getPublicDocumentById(id),
        enabled: !!id,
        staleTime: 1000 * 60 * 10,
    })
}

// ─── Existing hooks ───────────────────────────────────────────────────────────

export const useDocumentsByType = (type: string, query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.byType(type, query),
        queryFn: () => documentService.getByType(type, query),
        enabled: !!type,
    })
}

export const useDocumentsByGroup = (group: string, query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.byGroup(group, query),
        queryFn: () => documentService.getByGroup(group, query),
        enabled: !!group,
    })
}

export const useDocumentsByService = (service: string, query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.byService(service, query),
        queryFn: () => documentService.getByService(service, query),
        enabled: !!service,
    })
}

export const useDocumentsByStandard = (standard: string, query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.byStandard(standard, query),
        queryFn: () => documentService.getByStandard(standard, query),
        enabled: !!standard,
    })
}

export const useDocumentsByAssessment = (assessment: string, query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.byAssessment(assessment, query),
        queryFn: () => documentService.getByAssessment(assessment, query),
        enabled: !!assessment,
    })
}

export const useMyDocuments = (query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.mine(query),
        queryFn: () => documentService.getMyDocuments(query),
        staleTime: 1000 * 60 * 2,
    })
}

export const useDocumentsByStatus = (status: string, query?: PaginationQuery) => {
    return useQuery({
        queryKey: documentKeys.byStatus(status, query),
        queryFn: () => documentService.getByStatus(status, query),
        enabled: !!status,
        staleTime: 1000 * 60 * 5,
        refetchInterval: 1000 * 30,
        refetchIntervalInBackground: false,
    })
}

// ─── Mutations ────────────────────────────────────────────────────────────────

export const useUploadDocument = () => {
    const queryClient = useQueryClient()
    return useMutation<
        Awaited<ReturnType<typeof documentService.upload>>,
        Error,
        { body: CreateDocumentBody; file: File }
    >({
        mutationFn: ({ body, file }) => documentService.upload(body, file),
        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: documentKeys.all })
        },
    })
}

export const useUpdateDocument = () => {
    const queryClient = useQueryClient()
    return useMutation<
        Awaited<ReturnType<typeof documentService.update>>,
        Error,
        { body: UpdateDocumentBody; file?: File }
    >({
        mutationFn: ({ body, file }) => documentService.update(body, file),
        onSuccess: (_, { body }) => {
            queryClient.invalidateQueries({ queryKey: documentKeys.detail(body.document_id) })
            queryClient.invalidateQueries({ queryKey: documentKeys.all })
        },
    })
}

export const useApproveDocument = () => {
    const queryClient = useQueryClient()
    return useMutation<
        Awaited<ReturnType<typeof documentService.approve>>,
        Error,
        { body: ApprovalRequest; signedFile?: File }
    >({
        mutationFn: ({ body, signedFile }) => documentService.approve(body, signedFile),
        onSuccess: (_, { body }) => {
            queryClient.invalidateQueries({ queryKey: documentKeys.detail(body.document_id) })
            queryClient.invalidateQueries({ queryKey: documentKeys.all })
        },
    })
}

export const useDeleteDocument = () => {
    const queryClient = useQueryClient()
    return useMutation<
        Awaited<ReturnType<typeof documentService.delete>>,
        Error,
        string
    >({
        mutationFn: (documentId) => documentService.delete(documentId),
        onSuccess: (_, documentId) => {
            queryClient.removeQueries({ queryKey: documentKeys.detail(documentId) })
            queryClient.invalidateQueries({ queryKey: documentKeys.all })
        },
    })
}

export function useStats() {
    return useQuery<StatsResponse>({
        queryKey: documentKeys.stats,
        queryFn: statsService.getStats,
        staleTime: 1000 * 60 * 5,
        refetchInterval: 1000 * 30,
        refetchIntervalInBackground: false,
    })
}