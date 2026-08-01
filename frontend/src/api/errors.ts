export function hasHTTPStatus(error: unknown, status: number): boolean {
  return (
    typeof error === 'object' &&
    error !== null &&
    'status' in error &&
    error.status === status
  )
}
