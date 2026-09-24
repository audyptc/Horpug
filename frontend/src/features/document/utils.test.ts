import { describe, expect, it } from 'vitest'
import { getDocumentFileKind, getDocumentPreviewUrl, isAllowedUploadName, isOpenableUrl } from './utils'

describe('isOpenableUrl', () => {
  // Links are free text, and anything that isn't http(s) must never reach an
  // href/src.
  it('accepts only http(s)', () => {
    expect(isOpenableUrl('https://example.com/a.pdf')).toBe(true)
    expect(isOpenableUrl(' http://example.com ')).toBe(true)
    expect(isOpenableUrl('javascript:alert(1)')).toBe(false)
    expect(isOpenableUrl('data:text/html,<script>alert(1)</script>')).toBe(false)
    expect(isOpenableUrl('not a url')).toBe(false)
  })
})

describe('document previews', () => {
  it('detects the file kind from the link', () => {
    expect(getDocumentFileKind('https://example.com/scan.PDF')).toBe('pdf')
    expect(getDocumentFileKind('https://example.com/photo.jpeg')).toBe('image')
    expect(getDocumentFileKind('https://drive.google.com/file/d/abc123/view?usp=sharing')).toBe('google')
    expect(getDocumentFileKind('https://example.com/page')).toBe('other')
  })

  it('turns Google share links into the embeddable preview', () => {
    expect(getDocumentPreviewUrl('https://drive.google.com/file/d/abc123/view?usp=sharing')).toBe(
      'https://drive.google.com/file/d/abc123/preview'
    )
    expect(getDocumentPreviewUrl('https://docs.google.com/document/d/xyz/edit')).toBe(
      'https://docs.google.com/document/d/xyz/preview'
    )
  })

  it('has no preview for unsafe or unknown links', () => {
    expect(getDocumentPreviewUrl('javascript:alert(1)')).toBeNull()
    expect(getDocumentPreviewUrl('https://example.com/page')).toBeNull()
  })
})

describe('isAllowedUploadName', () => {
  it('checks the extension case-insensitively', () => {
    expect(isAllowedUploadName('ID-CARD.JPG')).toBe(true)
    expect(isAllowedUploadName('contract.pdf')).toBe(true)
    expect(isAllowedUploadName('run.exe')).toBe(false)
    expect(isAllowedUploadName('contract.pdf.exe')).toBe(false)
  })
})
