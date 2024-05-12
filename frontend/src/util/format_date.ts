// Input: 2024-05-04T20:08:44Z
// Output: May 4, 2024 8:08:44 PM
export default function format_date(date: string): string {
    const date_obj = new Date(date);
    return date_obj.toLocaleDateString('en-US', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
        hour: 'numeric',
        minute: 'numeric',
        hour12: true
    });
}