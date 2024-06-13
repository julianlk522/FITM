// Input: 2024-05-04T20:08:44Z
// Output: May 4, 2024 8:08:44 PM
export default function format_date(date: string): string {
    const date_obj = new Date(date);
    // format to local timezone
    date_obj.setMinutes(date_obj.getMinutes() + date_obj.getTimezoneOffset());
    return date_obj.toLocaleString("en-US", { month: 'short', day: 'numeric', year: 'numeric', hour: 'numeric', minute: 'numeric', hour12: true });
}