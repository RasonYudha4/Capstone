import { useForm, Controller } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { X } from 'lucide-react'
import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogTitle,
    DialogDescription,
} from '@/components/ui/dialog'
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from '@/components/ui/select'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import FileDropzone from '../atoms/FileDropzone'

const uploadSchema = z.object({
    kelompokKerja: z.string().min(1, 'Wajib dipilih'),
    fungsiPelayanan: z.string().min(1, 'Wajib dipilih'),
    standarAkreditasi: z.string().min(1, 'Wajib dipilih'),
    elemenPenilaian: z.string().min(1, 'Wajib dipilih'),
    namaBerkas: z.string().min(1, 'Nama berkas wajib diisi'),
    tipeBerkas: z.string().min(1, 'Wajib dipilih'),
    deskripsi: z.string().optional(),
    file: z.instanceof(File, { message: 'File wajib diunggah' }),
})

type UploadFormValues = z.infer<typeof uploadSchema>

interface UploadModalProps {
    open: boolean
    onOpenChange: (open: boolean) => void
}

const fieldClass = 'rounded-xl border-gray-200 bg-gray-50 focus:ring-[#6B5FAE] focus:border-[#6B5FAE] placeholder:text-gray-400 text-sm'
const labelClass = 'text-sm font-bold text-[#6B5FAE]'
const errorClass = 'text-xs text-red-500 mt-1'

export default function UploadModal({ open, onOpenChange }: UploadModalProps) {
    const {
        register,
        control,
        handleSubmit,
        setValue,
        watch,
        formState: { errors },
        reset,
    } = useForm<UploadFormValues>({
        resolver: zodResolver(uploadSchema),
    })

    const selectedFile = watch('file')

    const onSubmit = async (data: UploadFormValues) => {
        // TODO: plug in actual upload API call
        console.log('Upload submitted:', data)
        reset()
        onOpenChange(false)
    }

    const handleClose = () => {
        reset()
        onOpenChange(false)
    }

    return (
        <Dialog open={open} onOpenChange={handleClose}>
            <DialogContent className="max-w-3xl rounded-3xl bg-gray-50 p-8 gap-0 max-h-[90vh] min-w-175 [&>button]:hidden">
                {/* Header */}
                <DialogHeader className="mb-4">
                    <div className="flex items-start justify-between">
                        <div>
                            <DialogTitle className="text-2xl font-bold text-[#6B5FAE] mb-2">
                                Unggah Berkas
                            </DialogTitle>
                            <DialogDescription className="text-sm text-[#6B5FAE]/80 leading-relaxed max-w-lg">
                                Unggah dan simpan berkas terkait akreditasi rumah sakit ke dalam database.
                                Silahkan isi data yang diperlukan melalui formulir yang tertera.
                            </DialogDescription>
                        </div>
                        <button
                            onClick={handleClose}
                            className="text-gray-400 hover:text-gray-600 transition-colors mt-1 shrink-0"
                        >
                            <X className="w-5 h-5" />
                        </button>
                    </div>
                    <Separator className="mt-4 bg-[#6B5FAE]/30" />
                </DialogHeader>

                <div className="overflow-y-auto max-h-[65vh] pr-2">
                    <form onSubmit={handleSubmit(onSubmit)}>
                        {/* Two column layout: dropzone left, fields right */}

                        <div className="flex flex-col gap-2 mb-6">
                            <p className={labelClass}>Unggah Berkas</p>
                            <Controller
                                control={control}
                                name="file"
                                render={() => (
                                    <FileDropzone
                                        file={selectedFile}
                                        onFileSelect={(f) => setValue('file', f, { shouldValidate: true })}
                                    />
                                )}
                            />
                            {errors.file && <p className={errorClass}>{errors.file.message}</p>}
                        </div>

                        {/* Right — Form fields */}
                        <div className="flex flex-col gap-4">

                            {/* Row 1 */}
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <p className={labelClass}>Kelompok Kerja</p>
                                    <Controller control={control} name="kelompokKerja" render={({ field }) => (
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                                                <SelectValue placeholder="Kelompok Manajemen Rumah Sakit" />
                                            </SelectTrigger>
                                            <SelectContent className="rounded-xl">
                                                <SelectItem value="mrs">Manajemen Rumah Sakit</SelectItem>
                                                <SelectItem value="pbp">Pelayanan Berfokus Pasien</SelectItem>
                                                <SelectItem value="skp">Sasaran Keselamatan Pasien</SelectItem>
                                                <SelectItem value="pn">Program Nasional</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )} />
                                    {errors.kelompokKerja && <p className={errorClass}>{errors.kelompokKerja.message}</p>}
                                </div>

                                <div>
                                    <p className={labelClass}>Fungsi Pelayanan</p>
                                    <Controller control={control} name="fungsiPelayanan" render={({ field }) => (
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                                                <SelectValue placeholder="Tata Kelola Rumah Sakit" />
                                            </SelectTrigger>
                                            <SelectContent className="rounded-xl">
                                                <SelectItem value="tkrs">Tata Kelola Rumah Sakit</SelectItem>
                                                <SelectItem value="mkp">Manajemen Kualitas Pelayanan</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )} />
                                    {errors.fungsiPelayanan && <p className={errorClass}>{errors.fungsiPelayanan.message}</p>}
                                </div>
                            </div>

                            {/* Row 2 */}
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <p className={labelClass}>Standar Akreditasi</p>
                                    <Controller control={control} name="standarAkreditasi" render={({ field }) => (
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                                                <SelectValue placeholder="Standar 1" />
                                            </SelectTrigger>
                                            <SelectContent className="rounded-xl">
                                                {[1, 2, 3, 4].map(n => (
                                                    <SelectItem key={n} value={`s${n}`}>Standar {n}</SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    )} />
                                    {errors.standarAkreditasi && <p className={errorClass}>{errors.standarAkreditasi.message}</p>}
                                </div>

                                <div>
                                    <p className={labelClass}>Element Penilaian</p>
                                    <Controller control={control} name="elemenPenilaian" render={({ field }) => (
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                                                <SelectValue placeholder="Element Penilaian 1" />
                                            </SelectTrigger>
                                            <SelectContent className="rounded-xl">
                                                {[1, 2, 3, 4].map(n => (
                                                    <SelectItem key={n} value={`ep${n}`}>Elemen Penilaian {n}</SelectItem>
                                                ))}
                                            </SelectContent>
                                        </Select>
                                    )} />
                                    {errors.elemenPenilaian && <p className={errorClass}>{errors.elemenPenilaian.message}</p>}
                                </div>
                            </div>

                            {/* Row 3 */}
                            <div className="grid grid-cols-2 gap-4">
                                <div>
                                    <p className={labelClass}>Nama Berkas</p>
                                    <Input
                                        {...register('namaBerkas')}
                                        placeholder="Nama Berkas"
                                        className={`mt-1.5 ${fieldClass}`}
                                    />
                                    {errors.namaBerkas && <p className={errorClass}>{errors.namaBerkas.message}</p>}
                                </div>

                                <div>
                                    <p className={labelClass}>Tipe Berkas</p>
                                    <Controller control={control} name="tipeBerkas" render={({ field }) => (
                                        <Select onValueChange={field.onChange} value={field.value}>
                                            <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                                                <SelectValue placeholder="Surat Keputusan" />
                                            </SelectTrigger>
                                            <SelectContent className="rounded-xl">
                                                <SelectItem value="sk">Surat Keputusan</SelectItem>
                                                <SelectItem value="sop">SOP</SelectItem>
                                                <SelectItem value="laporan">Laporan</SelectItem>
                                                <SelectItem value="panduan">Panduan</SelectItem>
                                                <SelectItem value="notulen">Notulen</SelectItem>
                                            </SelectContent>
                                        </Select>
                                    )} />
                                    {errors.tipeBerkas && <p className={errorClass}>{errors.tipeBerkas.message}</p>}
                                </div>
                            </div>

                            {/* Deskripsi */}
                            <div>
                                <p className={labelClass}>Deskripsi</p>
                                <Textarea
                                    {...register('deskripsi')}
                                    placeholder="Masukan deskripsi tambahan"
                                    className={`mt-1.5 ${fieldClass} resize-none`}
                                    rows={3}
                                />
                            </div>
                        </div>

                        {/* Submit */}
                        <div className="flex justify-end mt-6">
                            <Button
                                type="submit"
                                className="bg-[#6B5FAE] hover:bg-[#5a4f9a] text-white rounded-xl px-10 font-semibold"
                            >
                                Kirim
                            </Button>
                        </div>
                    </form>
                </div>
            </DialogContent>
        </Dialog>
    )
}