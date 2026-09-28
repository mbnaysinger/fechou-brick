const MAX_DIMENSION = 1600
const JPEG_QUALITY = 0.8

/**
 * Redimensiona a foto no navegador antes do upload (lado maior <= 1600px,
 * JPEG ~0.8) para manter as linhas BYTEA pequenas no Postgres.
 */
export async function compressImage(file: File): Promise<Blob> {
  const bitmap = await createImageBitmap(file)
  try {
    const scale = Math.min(1, MAX_DIMENSION / Math.max(bitmap.width, bitmap.height))
    const width = Math.round(bitmap.width * scale)
    const height = Math.round(bitmap.height * scale)

    const canvas = document.createElement('canvas')
    canvas.width = width
    canvas.height = height
    const ctx = canvas.getContext('2d')
    if (!ctx) return file

    ctx.drawImage(bitmap, 0, 0, width, height)

    const blob = await new Promise<Blob | null>((resolve) =>
      canvas.toBlob(resolve, 'image/jpeg', JPEG_QUALITY)
    )
    return blob ?? file
  } finally {
    bitmap.close()
  }
}
