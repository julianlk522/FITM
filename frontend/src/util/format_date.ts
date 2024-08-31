// Input: 2024-05-04T20:08:44Z
// Output: May 4, 2024 8:08:44 PM
export function format_long_date(date: string): string {
	const date_obj = new Date(date)
	return new Intl.DateTimeFormat('en-US', {
		month: 'short',
		day: 'numeric',
		year: 'numeric',
		hour: 'numeric',
		minute: 'numeric',
		hour12: true,
	}).format(date_obj)
}

// Input: 2024-05-04
// Output: May 4, 2024
export function format_short_date(date: string): string {
	const date_obj = new Date(date)
	return new Intl.DateTimeFormat('en-US', {
		month: 'short',
		day: 'numeric',
		year: 'numeric',
		hour12: true,
	}).format(date_obj)
}
