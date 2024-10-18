import { useState } from 'preact/hooks'
import { LINKS_ENDPOINT } from '../../constants'
import * as types from '../../types'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import SearchCats from '../SearchFilters/Cats'
import Link from './Link'
import './NewLinks.css'
interface Props {
	Token: string
	User: string
}

export default function NewLinks(props: Props) {
	const [error, set_error] = useState<string | undefined>(undefined)
	const [dupe_url, set_dupe_url] = useState<string | undefined>(undefined)
	const [cats, set_cats] = useState<string[]>([])
	const [submitted_links, set_submitted_links] = useState<types.Link[]>([])

	function handle_url_change(e: InputEvent) {
		const val = (e.currentTarget as HTMLInputElement).value
		if (!val) {
			return
		}

		// auto-populate YT cats if URL is YT link
		if (
			(val.includes('youtu.be') || val.includes('youtube.com')) &&
			!cats.includes('YouTube')
		) {
			const cats_to_be_added = ['YouTube']

			// channel
			if (
				val.includes('youtube.com/@') &&
				!cats.includes('YouTube channels')
			) {
				cats_to_be_added.push('YouTube channels')
			}

			// playlist
			if (
				val.includes('youtube.com/playlist?') &&
				!cats.includes('YouTube playlists')
			) {
				cats_to_be_added.push('YouTube playlists')
			}

			set_cats(
				[...cats, ...cats_to_be_added].sort((a, b) =>
					a.localeCompare(b)
				)
			)
		}
	}

	async function handle_submit(event: SubmitEvent) {
		event.preventDefault()
		const form = event.target as HTMLFormElement
		const formData = new FormData(form)
		const url = formData.get('url')
		if (!url) {
			set_error('Missing URL')
			return
		} else if (!cats.length) {
			set_error('Missing tag')
			return
		}

		const summary = formData.get('summary')

		let resp_body: string

		if (summary) {
			resp_body = JSON.stringify({
				url,
				cats: cats.join(','),
				summary,
			})
		} else {
			resp_body = JSON.stringify({
				url,
				cats: cats.join(','),
			})
		}

		const new_link_resp = await fetch_with_handle_redirect(LINKS_ENDPOINT, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${props.Token}`,
			},
			body: resp_body,
		})
		if (!new_link_resp.Response || new_link_resp.RedirectTo) {
			return (window.location.href = new_link_resp.RedirectTo ?? '/500')
		}
		let new_link_data: types.Link | types.ErrorResponse =
			await new_link_resp.Response.json()

		if (is_error_response(new_link_data)) {
			if (new_link_data.error.includes('already submitted')) {
				const dupe_URL = new_link_data.error.split('See ')[1]
				set_error(new_link_data.error.split('See ')[0])
				set_dupe_url(dupe_URL)
			} else {
				set_error(new_link_data.error)
				set_dupe_url(undefined)
			}

			return
		} else {
			new_link_data.TagCount = 1

			set_submitted_links([...submitted_links, new_link_data])
			set_cats([])
			set_error(undefined)
			set_dupe_url(undefined)
			form.reset()
		}

		return
	}

	return (
		<>
			<section id='new-link'>
				<h2>New Link</h2>
				{error ? (
					<p class='error'>
						{`Error: ${error}`}
						{dupe_url ? (
							<>
								{' '}
								<a href={dupe_url}>View existing</a>
							</>
						) : null}
					</p>
				) : null}
				<form onSubmit={async (e) => await handle_submit(e)}>
					<label for='url'>URL</label>
					<input
						type='text'
						id='url'
						name='url'
						onInput={(e) => handle_url_change(e)}
					/>
					<label for='summary'>Summary (optional)</label>
					<textarea id='summary' name='summary' rows={3} cols={50} />
					<SearchCats
						SelectedCats={cats}
						SetSelectedCats={set_cats}
						SubmittedLinks={submitted_links}
					/>
					<input id='submit-new-link' type='submit' value='Submit' />
				</form>
			</section>
			{submitted_links.length ? (
				<section id='submitted-links'>
					<h2>Submitted</h2>
					<ol>
						{submitted_links.map((link) => (
							<Link
								key={link.ID}
								Link={link}
								IsNewPage
								SetNewLinkCats={set_cats}
								User={props.User}
							/>
						))}
					</ol>
				</section>
			) : null}
		</>
	)
}
