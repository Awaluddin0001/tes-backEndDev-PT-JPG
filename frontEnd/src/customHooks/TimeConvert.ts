export function formatDateToISO(originalDateString: string): string {
  const originalDate = new Date(originalDateString);
  return originalDate.toISOString();
}
