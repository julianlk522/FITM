import { useState } from 'preact/hooks';
import type { ErrorResponse, LinkData } from "../types";
import { is_error_response } from "../types";
import Link from './Link';
import './NewLinks.css';
import NewTag from './NewTag';
interface Props {
	Token: string
    User: string
}

export default function NewLinks(props: Props) {
    const {Token: token} = props

    const [error, set_error] = useState<string | undefined>(undefined)
    const [categories, set_categories] = useState<string[]>([])
    const [submitted_links, set_submitted_links] = useState<LinkData[]>([])

    async function handle_submit(event: SubmitEvent) {
        event.preventDefault()
        const form = event.target as HTMLFormElement
        const formData = new FormData(form)
        const url = formData.get('url')
    
        const new_link_resp = await fetch('http://127.0.0.1:8000/links', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`,
            },
            body: JSON.stringify({
                url,
                categories: categories.join(','),
            }),
        })
        if (new_link_resp.statusText === "Unauthorized") {
            window.location.href = '/login'
        }
        let new_link_data: LinkData | ErrorResponse = await new_link_resp.json()

        if (is_error_response(new_link_data)) {
            set_error(new_link_data.error)
            return
        } else {
            new_link_data.IsTagged = true
            set_submitted_links([...submitted_links, new_link_data])
            set_categories([])
            set_error(undefined)
        }
    
        return
    }

    return (
        <section>
            <h2>Enter New Link Details</h2>
            
            {error
                ? 
                    <p class='error'>{`Error: ${error}`}</p>
                :
                    null
            }

            <form onSubmit={async (e) => await handle_submit(e)}>
                <label for='url'>URL</label>
                <input type='text' id='url' name='url' />
                <br />

                <NewTag Categories={categories} SetCategories={set_categories} SetError={set_error} />
                
                <input type='submit' value='Submit' />
            </form>

            {submitted_links.length ? (
                <div id='submitted'>
                    <h2>Submitted Links</h2>
                    <ol>
                        {submitted_links.map((link) => (
                            <Link
                                key={link.ID}
                                Link={link}
                                Token={props.Token}
                                User={props.User}
                                IsSummaryPage={false}
                                IsTagPage={false} 
                            />
                        ))}
                    </ol>
                </div>
            ) : null}
        </section>
    )
    
}