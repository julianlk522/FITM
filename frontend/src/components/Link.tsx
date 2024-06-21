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
        IsTagged: is_tagged,
    } = props.link

    const [is_copied, set_is_copied] = useState(props.link.IsCopied)
    const [is_liked, set_is_liked] = useState(props.link.IsLiked)
    const [like_count, set_like_count] = useState(props.link.LikeCount)

    const split_cats = categories?.split(',')
    const categories_html = 
    // tag1 ==> <a href='/cat/tag1'>tag1</a>
    // tag1,tag2 ==> <a href='/cat/tag1'>tag1</a>, <a href='/cat/tag2'>tag2</a>
    split_cats?.map((cat, i) => {
        if (i === split_cats.length - 1) {
            return <a href={`/cat/${cat}`}>{cat}</a>
        } else {
            return <span><a href={`/cat/${cat}`}>{cat}</a>, </span>
        }
    }) 

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

    async function handle_copy() {
        if (!token) {
            document.cookie = `redirect_to=${window.location.pathname.replaceAll('/', '%2F')}; path=/; max-age=3600; SameSite=strict; Secure`
            return window.location.href = '/login'
        }

        if (!is_copied) {
            const copy_resp = await fetch(
                `http://127.0.0.1:8000/links/${id}/copy`,
                {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        Authorization: `Bearer ${token}`,
                    },
                }
            )
            const copy_data = await copy_resp.json()
            if (copy_data.ID) {
                set_is_copied(true)
                return
            } else {
                console.error("WTF is this: ",copy_data)
            }
        } else {
            const uncopy_resp = await fetch(
                `http://127.0.0.1:8000/links/${id}/copy`,
                {
                    method: 'DELETE',
                    headers: {
                        'Content-Type': 'application/json',
                        Authorization: `Bearer ${token}`,
                    },
                }
            )
            const uncopy_data = await uncopy_resp.json()
            if (uncopy_data.message === 'deleted') {
                set_is_copied(false)
                return
            } else {
                console.error("WTF is this: ", uncopy_data)
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
                Submitted by <a href={`/map/${submitted_by}`}>{submitted_by}</a> on {
                    format_date(submit_date)
                }
            </p>
            {categories 
                ? 
                    <>
                        <p>Categories: {categories_html}</p>
                        <a href={`/tag/${id}`}>
                            {is_tagged
                                ?
                                    'Edit Tag'
                                :
                                    'Add Tag'
                            }
                        </a>
                    </>
                : 
                        <p>No categories. <a href={`/tag/${id}`}>Add tag</a></p>
            
            }

            {is_summary_page 
                ? 
                    null
                : 
                    
                    <p>
                        <a href={`/summary/${id}`}>
                            {summary_count > 1
                                ? `View summaries (${summary_count})`
                                : 'Add summary'}
                        </a>
                    </p>
                    
            }

            {user !== submitted_by
                ?
                    <>
                        <button onClick={handle_like} class={`like-btn${is_liked ? ' liked' : ''}`}>
                        
                            {is_liked ? `Unlike (${like_count})` : `Like (${like_count})`}
                        </button>

                        <button onClick={handle_copy} class={`copy-btn${is_copied ? ' copied' : ''}`}>
                            {is_copied ? 'Uncopy' : 'Copy to Treasure Map'}
                        </button>
                    </> 
                : 
                    null
            }
        </li>
    );
}