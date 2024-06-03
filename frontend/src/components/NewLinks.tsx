import { useState } from 'preact/hooks';
import type { ErrorResponse, LinkData } from "../types";
import { is_error_response } from "../types";
import Link from './Link';
import './NewLinks.css';
interface Props {
	token: string
    user: string
}

export default function NewLinks(props: Props) {
    const [status, set_status] = useState<"Success" | "Error" | undefined>(undefined)
    const [message, set_message] = useState<string | undefined>(undefined)
    const [submitted_links, set_submitted_links] = useState<LinkData[]>([])

    async function handle_submit(event: SubmitEvent, token: string) {
        event.preventDefault()
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)
        const url = formData.get('url')
        const categories = formData.get('categories')
    
        const new_link_resp = await fetch('http://127.0.0.1:8000/links', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({
                url,
                categories,
            }),
        })
        if (new_link_resp.statusText === "Unauthorized") {
            window.location.href = '/login'
        }
        let new_link_data: LinkData | ErrorResponse = await new_link_resp.json()

        if (is_error_response(new_link_data)) {
            set_status("Error")
            set_message(new_link_data.error)
            return
        } else {
            set_submitted_links([...submitted_links, new_link_data])
            set_status("Success")
        }
    
        return
    }

    return (
        <div id='new_links'>
            <h2>Enter New Link Details</h2>
            {status 
                ? 
                    status === "Success"
                        ?
                            <p class='success'>Submitted</p>
                        :
                            <p class='error'>{`Error: ${message}`}</p>
                : null
            }
            <form id='new_link_form' onSubmit={async (e) => await handle_submit(e, props.token)}>
            <label for='url'>URL</label>
            <input type='text' id='url' name='url' />
            <br />
            <label for='categories'>Tag Category(ies)</label>
            <input type='text' id='categories' name='categories' />
            <input type='submit' value='Submit' />
            </form>

            {submitted_links.length ? (
                <div id='submitted'>
                    <h2>Submitted Links</h2>
                    <ul>
                        {submitted_links.map((link) => (
                            <Link
                                key={link.ID}
                                link={link}
                                token={props.token}
                                user={props.user}
                                is_summary_page={false} />
                        ))}
                    </ul>
                </div>
            ) : null}
        </div>
    )
    
}