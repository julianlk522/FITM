import { useState } from 'preact/hooks'
import { SUMMARIES_ENDPOINT } from '../../constants'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import { format_long_date } from '../../util/format_date'
import { save_action_and_path_then_redirect_to_login } from '../../util/login_redirect'
import SameUserLikeCount from '../Link/SameUserLikeCount'
import './Summary.css'

interface Props {
	ID: string
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

	const expected_like_action_status = 204
	async function handle_like() {
		if (!token) {
			save_action_and_path_then_redirect_to_login({
				Action: 'like summary',
				SummaryID: ID,
			})
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
			} else if (
				like_resp.Response.status !== expected_like_action_status
			) {
				const like_data = await like_resp.Response.json()
				if (is_error_response(like_data)) {
					return set_error(like_data.error)
				}
				return console.error('Whoops: ', like_data)
			}

			set_is_liked(true)
			set_like_count((prev) => prev + 1)
			return

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
			} else if (
				unlike_resp.Response.status !== expected_like_action_status
			) {
				const unlike_data = await unlike_resp.Response.json()
				if (is_error_response(unlike_data)) {
					return set_error(unlike_data.error)
				}
				return console.error('Whoops: ', unlike_data)
			}

			set_is_liked(false)
			set_like_count((prev) => prev - 1)
			return
		}
	}

	const expected_delete_action_status = 205
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
		} else if (
			delete_resp.Response.status !== expected_delete_action_status
		) {
			const delete_data = await delete_resp.Response.json()
			if (is_error_response(delete_data)) {
				return set_error(delete_data.error)
			}
			return console.error('Whoops: ', delete_data)
		}

		return window.location.reload()
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
