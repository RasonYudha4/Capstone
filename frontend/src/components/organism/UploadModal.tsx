import { useForm, Controller } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { X } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from "@/components/ui/dialog";

import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import FileDropzone from "../molecules/FileDropzone"
import { useState, useEffect } from "react"
import ConfirmDialog from "../molecules/ConfirmDialog"
import { useFormOptions } from "@/hooks/useFormOption"
import { useUploadDocument } from "@/hooks/useDocument"
import { useQueryClient } from '@tanstack/react-query'

const uploadSchema = z.object({
  serviceId: z.string().uuid("Wajib dipilih"),
  standardId: z.string().uuid("Wajib dipilih"),
  assessmentId: z.string().uuid("Wajib dipilih"),
  namaBerkas: z.string().min(1, "Nama berkas wajib diisi"),
  tipeBerkas: z.string().min(1, "Wajib dipilih"),
  deskripsi: z.string().optional(),
  file: z.instanceof(File, { message: "File wajib diunggah" }),
});

type UploadFormValues = z.infer<typeof uploadSchema>;

interface UploadModalProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
}

const fieldClass =
  "rounded-xl border-gray-200 bg-gray-50 focus:ring-[#6B5FAE] focus:border-[#6B5FAE] placeholder:text-gray-400 text-sm";
const labelClass = "text-sm font-bold text-[#6B5FAE]";
const errorClass = "text-xs text-red-500 mt-1";

export default function UploadModal({ open, onOpenChange }: UploadModalProps) {
  const {
    services,
    documentTypes,
    getStandards,
    getAssessments,
    isLoading,
    isError,
  } = useFormOptions();
  console.log("services:", services);
  console.log("isLoading:", isLoading);
  console.log("isError", isError);

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
  });

  const selectedFile = watch("file");
  const selectedServiceId = watch("serviceId");
  const selectedStandardId = watch("standardId");

  const standards = getStandards(selectedServiceId);
  const assessments = getAssessments(selectedServiceId, selectedStandardId);

  // Reset downstream when service changes
  useEffect(() => {
    setValue("standardId", "");
    setValue("assessmentId", "");
  }, [selectedServiceId, setValue]);

  // Reset assessment when standard changes
  useEffect(() => {
    setValue("assessmentId", "");
  }, [selectedStandardId, setValue]);

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [pendingData, setPendingData] = useState<UploadFormValues | null>(null);
  const { mutateAsync: uploadDocument, isPending } = useUploadDocument();
  const onSubmit = async (data: UploadFormValues) => {
    setPendingData(data);
    setConfirmOpen(true);
  };

  const onConfirmSubmit = async () => {
    if (!pendingData) return;
    try {
      await uploadDocument({
        body: {
          service_id: pendingData.serviceId,
          standard_id: pendingData.standardId,
          assessment_id: pendingData.assessmentId,
          document_type_id: pendingData.tipeBerkas,
          filename: pendingData.namaBerkas,
          description: pendingData.deskripsi,
        },
        file: pendingData.file,
      });
      setConfirmOpen(false);
      setPendingData(null);
      reset();
      onOpenChange(false);
    } catch (error) {
      console.error("Upload failed:", error);
      // TODO: show toast error
    }
  };

  const handleClose = () => {
    reset();
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="max-w-3xl rounded-3xl bg-gray-50 p-8 gap-0 max-h-[90vh] min-w-175 [&>button]:hidden">
        <ConfirmDialog
          open={confirmOpen}
          onOpenChange={setConfirmOpen}
          title="Unggah Berkas?"
          description="Pastikan semua data yang diisi sudah benar sebelum mengunggah berkas ini."
          confirmLabel="Ya, Kirim"
          variant="default"
          onConfirm={onConfirmSubmit}
        />

        {/* Header */}
        <DialogHeader className="mb-4">
          <div className="flex items-start justify-between">
            <div>
              <DialogTitle className="text-2xl font-bold text-[#6B5FAE] mb-2">
                Unggah Berkas
              </DialogTitle>
              <DialogDescription className="text-sm text-[#6B5FAE]/80 leading-relaxed max-w-lg">
                Unggah dan simpan berkas terkait akreditasi rumah sakit ke dalam
                database. Silahkan isi data yang diperlukan melalui formulir
                yang tertera.
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
            {/* Dropzone */}
            <div className="flex flex-col gap-2 mb-6">
              <p className={labelClass}>Unggah Berkas</p>
              <Controller
                control={control}
                name="file"
                render={() => (
                  <FileDropzone
                    file={selectedFile}
                    onFileSelect={(f) =>
                      setValue("file", f, { shouldValidate: true })
                    }
                  />
                )}
              />
              {errors.file && (
                <p className={errorClass}>{errors.file.message}</p>
              )}
            </div>

            <div className="flex flex-col gap-4">
              {/* Row 1 — Fungsi Pelayanan (Service) */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className={labelClass}>Fungsi Pelayanan</p>
                  <Controller
                    control={control}
                    name="serviceId"
                    render={({ field }) => (
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                        disabled={isLoading}
                      >
                        <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                          <SelectValue
                            placeholder={
                              isLoading ? "Memuat..." : "Pilih Fungsi Pelayanan"
                            }
                          />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                          {services.map((s) => (
                            <SelectItem key={s.id} value={s.id}>
                              {s.code} — {s.description}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  />
                  {errors.serviceId && (
                    <p className={errorClass}>{errors.serviceId.message}</p>
                  )}
                </div>

                {/* Standar Akreditasi (Standard) */}
                <div>
                  <p className={labelClass}>Standar Akreditasi</p>
                  <Controller
                    control={control}
                    name="standardId"
                    render={({ field }) => (
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                        disabled={!selectedServiceId}
                      >
                        <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                          <SelectValue
                            placeholder={
                              selectedServiceId
                                ? "Pilih Standar"
                                : "Pilih Fungsi Pelayanan dulu"
                            }
                          />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                          {standards.map((s) => (
                            <SelectItem key={s.id} value={s.id}>
                              {s.code} — {s.description}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  />
                  {errors.standardId && (
                    <p className={errorClass}>{errors.standardId.message}</p>
                  )}
                </div>
              </div>

              {/* Row 2 — Elemen Penilaian (Assessment) + Nama Berkas */}
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <p className={labelClass}>Elemen Penilaian</p>
                  <Controller
                    control={control}
                    name="assessmentId"
                    render={({ field }) => (
                      <Select
                        value={field.value}
                        onValueChange={field.onChange}
                        disabled={!selectedStandardId}
                      >
                        <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                          <SelectValue
                            placeholder={
                              selectedStandardId
                                ? "Pilih Elemen Penilaian"
                                : "Pilih Standar dulu"
                            }
                          />
                        </SelectTrigger>
                        <SelectContent className="rounded-xl">
                          {assessments.map((a) => (
                            <SelectItem key={a.id} value={a.id}>
                              {a.code} — {a.description}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  />
                  {errors.assessmentId && (
                    <p className={errorClass}>{errors.assessmentId.message}</p>
                  )}
                </div>

                <div>
                  <p className={labelClass}>Nama Berkas</p>
                  <Input
                    {...register("namaBerkas")}
                    placeholder="Nama Berkas"
                    className={`mt-1.5 ${fieldClass}`}
                  />
                  {errors.namaBerkas && (
                    <p className={errorClass}>{errors.namaBerkas.message}</p>
                  )}
                </div>
              </div>

              {/* Row 3 — Tipe Berkas */}
              <div>
                <p className={labelClass}>Tipe Berkas</p>
                <Controller
                  control={control}
                  name="tipeBerkas"
                  render={({ field }) => (
                    <Select
                      value={field.value}
                      onValueChange={field.onChange}
                      disabled={isLoading}
                    >
                      <SelectTrigger className={`mt-1.5 ${fieldClass}`}>
                        <SelectValue
                          placeholder={
                            isLoading ? "Memuat..." : "Pilih Tipe Berkas"
                          }
                        />
                      </SelectTrigger>
                      <SelectContent className="rounded-xl">
                        {documentTypes.map((dt) => (
                          <SelectItem key={dt.id} value={dt.id}>
                            {dt.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  )}
                />
                {errors.tipeBerkas && (
                  <p className={errorClass}>{errors.tipeBerkas.message}</p>
                )}
              </div>

              {/* Deskripsi */}
              <div>
                <p className={labelClass}>Deskripsi</p>
                <Textarea
                  {...register("deskripsi")}
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
                disabled={isPending}
                className="bg-[#6B5FAE] hover:bg-[#5a4f9a] text-white rounded-xl px-10 font-semibold"
              >
                {isPending ? "Mengunggah..." : "Kirim"}
              </Button>
            </div>
          </form>
        </div>
      </DialogContent>
    </Dialog>
  );
}
