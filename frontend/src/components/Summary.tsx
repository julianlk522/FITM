import { useState } from 'preact/hooks';
import './Summary.css';

interface Props {
    Token: string | undefined
	ID: number
	Text: string
	SubmittedBy: string
	LikeCount: number
	IsLiked: boolean | undefined
}

export default function Summary(props: Props) {
	const {ID, Token: token, Text: text, SubmittedBy: submitted_by} = props

	const [is_liked, set_is_liked] = useState(props.IsLiked)
    const [like_count, set_like_count] = useState(props.LikeCount)

	async function handle_like() {
		if (!token) {
			return (window.location.href = '/login')
		}
	
		// like
		if (!is_liked) {
			const like_resp = await fetch(
				`http://127.0.0.1:8000/summaries`,
				{
					method: 'POST',
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`,
					},
					body: JSON.stringify({
						summary_id: ID
					})
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
				`http://127.0.0.1:8000/summaries`,
				{
					method: 'DELETE',
					headers: {
						'Content-Type': 'application/json',
						Authorization: `Bearer ${token}`,
					},
					// body: JSON.stringify({
					// 	slike_id: ID
					// })
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
		<li class='summary'>
			{text}
			<p>Submitted By: {submitted_by}</p>
			<button
				onClick={handle_like}
				class={`like-btn${is_liked ? ' liked' : ''}`}>Like ({like_count})
			</button>
		</li>
	)
}