import { useState } from 'preact/hooks'
import { SUMMARIES_ENDPOINT } from '../../constants'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import { format_long_date } from '../../util/format_date'
import SameUserLikeCount from '../Link/SameUserLikeCount'
import './Summary.css'

interface Props {
	ID: number
	Text: string
	SubmittedBy: string
	LastUpdated: string
	LikeCount: number
	IsLiked?: boolean
	Token?: string
	User?: string
}

export default function Summary(props: Props) {
	const {
		ID,
		Token: token,
		User: user,
		Text: text,
		SubmittedBy: submitted_by,
		LastUpdated: last_updated,
	} = props

	const [is_liked, set_is_liked] = useState(props.IsLiked)
	const [like_count, set_like_count] = useState(props.LikeCount)
	const [error, set_error] = useState<string | undefined>(undefined)

	const like_api_url = SUMMARIES_ENDPOINT + `/${ID}/like`

	async function handle_like() {
		if (!token) {
			const encoded_url = window.location.pathname.replaceAll('/', '%2F')
			document.cookie = `redirect_to=${encoded_url}; path=/login; max-age=14400; SameSite=strict; Secure;`
			document.cookie = `redirect_action=like summary ${ID}; path=${window.location.pathname}; max-age=14400; SameSite=strict; Secure`
			return (window.location.href = '/login')
		}

		// like
		if (!is_liked) {
			const like_resp = await fetch_with_handle_redirect(like_api_url, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
			})
			if (!like_resp.Response || like_resp.RedirectTo) {
				return (window.location.href = like_resp.RedirectTo ?? '/500')
			}
			const like_data = await like_resp.Response.json()
			if (is_error_response(like_data)) {
				return set_error(like_data.error)
			} else if (like_data.message === 'liked') {
				set_is_liked(true)
				set_like_count(like_count + 1)
				return
			} else {
				return console.error('Whoops: ', like_data)
			}

			// unlike
		} else {
			const unlike_resp = await fetch_with_handle_redirect(like_api_url, {
				method: 'DELETE',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
			})
			if (!unlike_resp.Response || unlike_resp.RedirectTo) {
				return (window.location.href = unlike_resp.RedirectTo ?? '/500')
			}

			const unlike_data = await unlike_resp.Response.json()
			if (is_error_response(unlike_data)) {
				return set_error(unlike_data.error)
			} else if (unlike_data.message === 'deleted') {
				set_is_liked(false)
				set_like_count(like_count - 1)
			} else {
				return console.error('Whoops: ', unlike_data)
			}

			return
		}
	}

	async function handle_delete() {
		if (!token) {
			return (window.location.href = '/login')
		}
		const delete_resp = await fetch_with_handle_redirect(
			SUMMARIES_ENDPOINT,
			{
				method: 'DELETE',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
				body: JSON.stringify({ summary_id: ID }),
			}
		)
		if (!delete_resp.Response || delete_resp.RedirectTo) {
			return (window.location.href = delete_resp.RedirectTo ?? '/500')
		}

		const delete_data = await delete_resp.Response.json()
		if (is_error_response(delete_data)) {
			return set_error(delete_data.error)
		} else if (delete_data.message === 'deleted') {
			return window.location.reload()
		} else {
			return console.error('Whoops: ', delete_data)
		}
	}
	return (
		<li class='summary'>
			"{text}"
			<p>
				submitted by{' '}
				{submitted_by === 'Auto Summary' ? (
					<span class='auto-summary'>Auto Summary</span>
				) : (
					<a href={`/map/${submitted_by}`}>{submitted_by}</a>
				)}
			</p>
			<p>last updated: {format_long_date(last_updated)}</p>
			{user !== submitted_by ? (
				<button
					onClick={handle_like}
					class={`like-btn${is_liked ? ' liked' : ''}`}
				>
					{is_liked ? 'Unlike' : 'Like'} ({like_count})
				</button>
			) : (
				<>
					<SameUserLikeCount LikeCount={like_count} />
					<button
						id='delete-summary-btn'
						class='img-btn'
						onClick={handle_delete}
					>
						<img src='../../../x-lg.svg' height={20} width={20} />
					</button>
					{error ? <p class='error'>{error}</p> : null}
				</>
			)}
		</li>
	)
}
