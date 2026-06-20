import { create } from 'zustand'

interface FilterState {
    serviceId:        string
    standardId:       string
    assessmentId:     string
    pendingHighlightId:  string | null
    setServiceId:           (id: string) => void
    setStandardId:          (id: string) => void
    setAssessmentId:        (id: string) => void
    setAllFilters:          (serviceId: string, standardId: string, assessmentId: string) => void
    setPendingHighlight:    (documentId: string | null) => void
}

export const useFilterStore = create<FilterState>((set) => ({
    serviceId:       '',
    standardId:      '',
    assessmentId:    '',
    pendingHighlightId: null,
    setServiceId:    (id) => set({ serviceId: id, standardId: '', assessmentId: '' }),
    setStandardId:   (id) => set({ standardId: id, assessmentId: '' }),
    setAssessmentId: (id) => set({ assessmentId: id }),
    setAllFilters: (serviceId, standardId, assessmentId) =>
        set({ serviceId, standardId, assessmentId }),
    setPendingHighlight: (documentId) => set({ pendingHighlightId: documentId }),
}))