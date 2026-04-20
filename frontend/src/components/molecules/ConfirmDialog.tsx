import {
    AlertDialog,
    AlertDialogAction,
    AlertDialogCancel,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
} from '@/components/ui/alert-dialog'

export type ConfirmVariant = 'danger' | 'warning' | 'default'

interface ConfirmDialogProps {
    open: boolean
    onOpenChange: (open: boolean) => void
    title: string
    description: string
    confirmLabel?: string
    cancelLabel?: string
    variant?: ConfirmVariant
    onConfirm: () => void
}

const confirmButtonClass: Record<ConfirmVariant, string> = {
    danger: 'bg-red-500 hover:bg-red-600 text-white border-0 focus:ring-red-500',
    warning: 'bg-amber-500 hover:bg-amber-600 text-white border-0 focus:ring-amber-500',
    default: 'bg-[#6B5FAE] hover:bg-[#5a4f9a] text-white border-0 focus:ring-[#6B5FAE]',
}

export default function ConfirmDialog({
    open,
    onOpenChange,
    title,
    description,
    confirmLabel = 'Konfirmasi',
    cancelLabel = 'Batal',
    variant = 'default',
    onConfirm,
}: ConfirmDialogProps) {
    return (
        <AlertDialog open={open} onOpenChange={onOpenChange}>
            <AlertDialogContent className="rounded-2xl border border-gray-100 shadow-xl max-w-sm">
                <AlertDialogHeader>
                    <AlertDialogTitle className="text-base font-bold text-gray-900">
                        {title}
                    </AlertDialogTitle>
                    <AlertDialogDescription className="text-sm text-gray-500 leading-relaxed">
                        {description}
                    </AlertDialogDescription>
                </AlertDialogHeader>
                <AlertDialogFooter className="gap-2 mt-2">
                    <AlertDialogCancel className="rounded-xl text-sm font-medium border-gray-200 hover:bg-gray-50">
                        {cancelLabel}
                    </AlertDialogCancel>
                    <AlertDialogAction
                        onClick={onConfirm}
                        className={`rounded-xl text-sm font-semibold ${confirmButtonClass[variant]}`}
                    >
                        {confirmLabel}
                    </AlertDialogAction>
                </AlertDialogFooter>
            </AlertDialogContent>
        </AlertDialog>
    )
}