import { create } from 'zustand'

interface FilterState {
    serviceId:        string
    standardId:       string
    assessmentId:     string
    setServiceId:     (id: string) => void
    setStandardId:    (id: string) => void
    setAssessmentId:  (id: string) => void
}

export const useFilterStore = create<FilterState>((set) => ({
    serviceId:       '',
    standardId:      '',
    assessmentId:    '',
    setServiceId:    (id) => set({ serviceId: id, standardId: '', assessmentId: '' }),
    setStandardId:   (id) => set({ standardId: id, assessmentId: '' }),
    setAssessmentId: (id) => set({ assessmentId: id }),
}))