import { useRef, useState, type FormEvent } from 'react'
import { api, extractErrorCode, extractErrorMessage, type ApiPage } from '@/shared/api/client'
import { useLanguage } from '@/shared/i18n/language'
import type { ApiTenant } from '@/features/tenant/types'
import type { ApiRoom } from '@/features/room/types'
import type { ApiContract } from '@/features/contract/types'
import { formatRoomLabel } from '@/features/room/utils'
import type { ApiDocument, DocumentCategory } from './types'
import {
  DOCUMENT_MAX_UPLOAD_BYTES,
  isAllowedUploadName,
  toApiDate,
  toDateInputValue,
} from './utils'

// What a caller already knows when it opens the create form, so the user
// doesn't have to pick it again (e.g. the tenant whose documents are showing).
export type DocumentFormPrefill = { tenant?: ApiTenant; room?: ApiRoom }

// Owns the create/edit document form — its fields, validation and submit — so
// the Documents page and the per-tenant / per-room document sheets share one
// implementation. `sheetProps` spreads straight onto <DocumentFormSheet>.
export function useDocumentForm({ onSaved }: { onSaved: () => void }) {
  const { t } = useLanguage()

  const [open, setOpen] = useState(false)
  const [documentId, setDocumentId] = useState<string | null>(null)
  const [dormitoryId, setDormitoryId] = useState('')
  // The picked dormitory is searched server-side and so isn't necessarily in
  // any loaded list; its name is kept here for the selector's label.
  const [dormitoryName, setDormitoryName] = useState('')
  const [tenantId, setTenantId] = useState('')
  const [tenantDisplayName, setTenantDisplayName] = useState('')
  const [roomId, setRoomId] = useState('')
  const [roomDisplayLabel, setRoomDisplayLabel] = useState('')
  const [name, setName] = useState('')
  const [category, setCategory] = useState<DocumentCategory>('other')
  const [fileUrl, setFileUrl] = useState('')
  const [file, setFile] = useState<File | null>(null)
  // Name of the file already stored for the document being edited, if it has
  // one; it stays unless a new file or link is supplied.
  const [storedFileName, setStoredFileName] = useState<string | null>(null)
  const [uploadedDate, setUploadedDate] = useState('')
  const [note, setNote] = useState('')
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

  // Mirrors of the room/dormitory picks, readable from the async contract
  // lookup below (its closure would otherwise see the values from when the
  // tenant was picked, and could overwrite a room the user chose meanwhile).
  const roomIdRef = useRef('')
  const dormitoryIdRef = useRef('')
  // Bumped on every open/tenant pick so a slow lookup for an earlier tenant
  // can't fill in a form that has since moved on.
  const lookupToken = useRef(0)

  function applyRoom(id: string, label: string) {
    roomIdRef.current = id
    setRoomId(id)
    setRoomDisplayLabel(label)
  }

  function applyDormitory(id: string, dormitory: string) {
    dormitoryIdRef.current = id
    setDormitoryId(id)
    setDormitoryName(dormitory)
  }

  // A room belongs to exactly one dormitory, so when creating, picking it
  // settles that too. An existing document's dormitory is fixed.
  function selectRoom(room: ApiRoom, creating: boolean) {
    applyRoom(room.id, formatRoomLabel(room))
    if (creating) applyDormitory(room.dormitory_id, room.dormitory_name ?? '')
  }

  // Tenants aren't tied to a dormitory themselves, but their active contract
  // is tied to a room — fill in that room and its dormitory when the user
  // hasn't chosen either. Best-effort: on failure the user just picks them.
  function selectTenant(tenant: ApiTenant, creating: boolean) {
    setTenantId(tenant.id)
    setTenantDisplayName(`${tenant.first_name} ${tenant.last_name}`)

    if (!creating || roomIdRef.current) return

    const token = ++lookupToken.current
    api
      .get<ApiPage<ApiContract[]>>('/contracts', {
        params: { tenant_id: tenant.id, status: 'active', per_page: 1 },
      })
      .then(({ data }) => {
        const contract = data.data[0]
        if (!contract || token !== lookupToken.current || roomIdRef.current) return

        applyRoom(contract.room_id, formatRoomLabel(contract))
        if (!dormitoryIdRef.current && contract.dormitory_id) {
          applyDormitory(contract.dormitory_id, contract.dormitory_name ?? '')
        }
      })
      .catch(() => {})
  }

  function clearTenant() {
    lookupToken.current++
    setTenantId('')
    setTenantDisplayName('')
  }

  function clearRoom() {
    applyRoom('', '')
  }

  function pickFile(picked: File | null) {
    setFile(picked)
    // Saves retyping the obvious: an unnamed document takes its file's name.
    if (picked && !name.trim()) setName(picked.name.replace(/\.[^.]+$/, ''))
  }

  function openCreate(prefill?: DocumentFormPrefill) {
    lookupToken.current++
    setDocumentId(null)
    applyDormitory('', '')
    applyRoom('', '')
    setTenantId('')
    setTenantDisplayName('')
    setName('')
    setCategory('other')
    setFileUrl('')
    setFile(null)
    setStoredFileName(null)
    setUploadedDate(toDateInputValue(new Date().toISOString()))
    setNote('')
    setError(null)

    if (prefill?.room) selectRoom(prefill.room, true)
    if (prefill?.tenant) selectTenant(prefill.tenant, true)

    setOpen(true)
  }

  function openEdit(document: ApiDocument) {
    lookupToken.current++
    setDocumentId(document.id)
    applyDormitory(document.dormitory_id, document.dormitory_name ?? '')
    setTenantId(document.tenant_id ?? '')
    setTenantDisplayName(document.tenant_name ?? '')
    applyRoom(document.room_id ?? '', formatRoomLabel(document))
    setName(document.name)
    setCategory(document.category)
    setFileUrl(document.file_url)
    setFile(null)
    setStoredFileName(document.has_file ? document.file_name || document.name : null)
    setUploadedDate(toDateInputValue(document.uploaded_date))
    setNote(document.note)
    setError(null)
    setOpen(true)
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const isEdit = documentId !== null

    if (!isEdit && !dormitoryId) {
      setError(t('documentDormitoryRequired'))
      return
    }
    if (!name.trim()) {
      setError(t('documentNameRequired'))
      return
    }
    if (!file && !fileUrl.trim() && !storedFileName) {
      setError(t('documentFileUrlRequired'))
      return
    }
    if (file) {
      if (file.size > DOCUMENT_MAX_UPLOAD_BYTES) {
        setError(t('documentFileTooLarge'))
        return
      }
      if (file.size === 0) {
        setError(t('documentFileEmpty'))
        return
      }
      if (!isAllowedUploadName(file.name)) {
        setError(t('documentFileUnsupported'))
        return
      }
    }

    setSaving(true)
    setError(null)

    try {
      // A picked file goes up as multipart (the file takes precedence over any
      // link); otherwise it's the plain JSON body. An edit that keeps the
      // stored file and has no link leaves file_url out, so the file isn't
      // replaced by an empty link.
      const link = fileUrl.trim()
      const fields = {
        tenant_id: tenantId || null,
        room_id: roomId || null,
        name: name.trim(),
        category,
        uploaded_date: uploadedDate ? toApiDate(uploadedDate) : undefined,
        note: note.trim(),
      }

      let body: Record<string, unknown> | FormData
      if (file) {
        const form = new FormData()
        if (!isEdit) form.append('dormitory_id', dormitoryId)
        for (const [key, value] of Object.entries(fields)) {
          if (value) form.append(key, value)
        }
        // Empty note is meaningful (it clears it), so it's always sent.
        form.set('note', fields.note)
        form.append('file', file)
        body = form
      } else {
        body = {
          ...(isEdit ? {} : { dormitory_id: dormitoryId }),
          ...fields,
          ...(link || !storedFileName ? { file_url: link } : {}),
        }
      }

      if (!isEdit) {
        await api.post<ApiDocument>('/documents', body)
      } else {
        await api.put<ApiDocument>(`/documents/${documentId}`, body)
      }
      onSaved()
      setOpen(false)
    } catch (err) {
      const fallback = isEdit ? t('documentUpdateError') : t('documentCreateError')
      const codeMessages: Record<string, string> = {
        file_too_large: t('documentFileTooLarge'),
        unsupported_file_type: t('documentFileUnsupported'),
        empty_file: t('documentFileEmpty'),
      }
      setError(codeMessages[extractErrorCode(err) ?? ''] ?? extractErrorMessage(err, fallback))
    } finally {
      setSaving(false)
    }
  }

  return {
    openCreate,
    openEdit,
    sheetProps: {
      open,
      onOpenChange: setOpen,
      isEdit: documentId !== null,
      dormitoryName,
      onDormitorySelect: (dormitory: { id: string; name: string }) => applyDormitory(dormitory.id, dormitory.name),
      tenantDisplayName,
      onSelectTenant: (tenant: ApiTenant) => selectTenant(tenant, documentId === null),
      onClearTenant: clearTenant,
      roomDisplayLabel,
      onSelectRoom: (room: ApiRoom) => selectRoom(room, documentId === null),
      onClearRoom: clearRoom,
      name,
      onNameChange: setName,
      category,
      onCategoryChange: setCategory,
      file,
      onFileChange: pickFile,
      storedFileName,
      fileUrl,
      onFileUrlChange: setFileUrl,
      uploadedDate,
      onUploadedDateChange: setUploadedDate,
      note,
      onNoteChange: setNote,
      saving,
      error,
      onSubmit: handleSubmit,
    },
  }
}
