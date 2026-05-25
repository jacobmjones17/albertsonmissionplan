/** Fetch a same-origin image and return a data URL for offline HTML documents. */
export async function fetchImageAsDataUrl(path: string): Promise<string> {
  const res = await fetch(path)
  if (!res.ok) {
    throw new Error(`Could not load image: ${path}`)
  }
  const blob = await res.blob()
  return await new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => resolve(String(reader.result))
    reader.onerror = () => reject(reader.error ?? new Error('Could not read image.'))
    reader.readAsDataURL(blob)
  })
}
