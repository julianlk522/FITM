import { useState } from 'preact/hooks';
import type { LinkData } from '../types';
import format_date from '../util/format_date';
import './Link.css';

interface Props {
	link: LinkData
    is_summary_page: boolean
    token: string | undefined
    user: string | undefined
}

export default function Link(props: Props) {
    const {is_summary_page, token, user} = props
    const {
        ID: id,
        URL: url,
        SubmittedBy: submitted_by,
        SubmitDate: submit_date,
        Categories: categories,
        Summary: summary,
        SummaryCount: summary_count,
        ImgURL: img_url,
    } = props.link

    const [is_liked, set_is_liked] = useState(props.link.IsLiked)
    const [like_count, set_like_count] = useState(props.link.LikeCount)

    async function handle_like() {
        if (!token) {
            document.cookie = `redirect_to=${window.location.pathname.replaceAll('/', '%2F')}; path=/; max-age=3600; SameSite=strict; Secure`
            return window.location.href = '/login'
        }

        // like
        if (!is_liked) {
            const like_resp = await fetch(
                `http://127.0.0.1:8000/links/${id}/like`,
                {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        Authorization: `Bearer ${token}`,
                    },
                }
            )
            const like_data = await like_resp.json()
            if (like_data.ID) {
                set_is_liked(true)
                set_like_count(like_count + 1)
                return
            } else {
                console.error("WTF is this: ",like_data)
            }

        // unlike
        } else {
            const unlike_resp = await fetch(
                `http://127.0.0.1:8000/links/${id}/like`,
                {
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json',
                        Authorization: `Bearer ${token}`,
                    },
                }
            )
            const unlike_data = await unlike_resp.json()
            if (unlike_data.message === 'deleted') {
                set_is_liked(false)
                set_like_count(like_count - 1)
                return
            } else {
                console.error("WTF is this: ", unlike_data)
            }
	    }
    }

    return (
        <li class='link'>
            {img_url
                ? 
                <div class="preview"><
                    img src={img_url} alt={summary ? summary : url} height={100} width={100} />
                    <a href={url}>
                <h2>
                    {summary ? summary : url}
                </h2>
            </a>
                    </div>
                : <a href={url}>
                <h2>
                    {summary ? summary : url}
                </h2>
            </a>}

            {summary ? <p class='url'>{url}</p> : null}
            
            <p>
                Submitted By: <a href={`/map/${submitted_by}`}>{submitted_by}</a> on {
                    format_date(submit_date)
                }
            </p>
            {categories 
                ? 
                    <>
                        <p>Categories: {categories}</p>
                        <a href='/'>Add tag</a>
                    </>
                : 
                        <p>No categories. <a href='/'>Add tag</a></p>
            
            }

            {is_summary_page 
                ? 
                    null
                : 
                    
                    <p>
                        <a href={`/summary/${id}`}>
                            {summary_count > 1
                                ? `View all summaries (${summary_count}) or add new`
                                : 'Add summary'}
                        </a>
                    </p>
                    
            }

            {user !== submitted_by ? <button onClick={handle_like} class={`like-btn${is_liked ? ' liked' : ''}`}
            >
                {is_liked ? `Unlike (${like_count})` : `Like (${like_count})`}
            </button> : null}
        </li>
    );
}