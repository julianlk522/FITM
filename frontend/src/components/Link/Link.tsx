import {
	useEffect,
	useState,
	type Dispatch,
	type StateUpdater,
} from 'preact/hooks'
import { LINKS_ENDPOINT } from '../../constants'
import * as types from '../../types'
import { is_error_response } from '../../types'
import fetch_with_handle_redirect from '../../util/fetch_with_handle_redirect'
import { format_long_date } from '../../util/format_date'
import {
	save_action_and_path_then_redirect_to_login,
	save_path_then_redirect_to_login,
} from '../../util/login_redirect'
import Modal from '../Modal/Modal'
import './Link.css'
import SameUserLikeCount from './SameUserLikeCount'

interface Props {
	Link: types.Link
	CatsFromUser?: string
	IsSummaryPage?: boolean
	IsTagPage?: boolean
	IsTmapPage?: boolean
	IsNewPage?: boolean
	SetNewLinkCats?: Dispatch<StateUpdater<string[]>>
	NSFWCatLinks?: boolean
	Token?: string
	User?: string
}

export default function Link(props: Props) {
	const {
		CatsFromUser: cats_from_user,
		IsSummaryPage: is_summary_page,
		IsTagPage: is_tag_page,
		IsTmapPage: is_tmap_page,
		IsNewPage: is_new_page,
		SetNewLinkCats: set_new_link_cats,
		NSFWCatLinks: nsfw_cat_links,
		Token: token,
		User: user,
	} = props
	const {
		ID: id,
		URL: url,
		SubmittedBy: submitted_by,
		SubmitDate: submit_date,
		Cats: cats,
		Summary: summary,
		SummaryCount: summary_count,
		TagCount: tag_count,
		ImgURL: saved_img_url,
	} = props.Link

	const is_your_link = user !== undefined && submitted_by === user
	const has_only_your_tag = tag_count === 1 && is_your_link

	const [is_copied, set_is_copied] = useState(props.Link.IsCopied)
	const [is_liked, set_is_liked] = useState(props.Link.IsLiked)
	const [like_count, set_like_count] = useState(props.Link.LikeCount)
	const [show_delete_modal, set_show_delete_modal] = useState(false)
	const [img_url, set_img_url] = useState(saved_img_url)

	// hide preview image if URL fails to resolve
	useEffect(() => {
		if (saved_img_url) {
			// use Image constructor so no CORS issues, even locally
			const img = new Image()
			img.onload = () => set_img_url(saved_img_url)
			img.onerror = () => set_img_url(undefined)
			img.src = saved_img_url
		}
	}, [saved_img_url])

	// correct submitted_by to local time if newly created link
	let submit_date_in_local_time = ''
	if (is_new_page) {
		const sd_utc = new Date(submit_date)
		const tz_offset_millis = sd_utc.getTimezoneOffset() * 60000
		submit_date_in_local_time = new Date(
			sd_utc.getTime() - tz_offset_millis
		).toISOString()
	}

	let tag_attribution =
		cats && user && cats_from_user === user
			? 'your tag'
			: cats_from_user
				? `${cats_from_user}'s tag`
				: tag_count === 1
					? `${submitted_by}'s tag`
					: 'global tag'
	tag_attribution += ` (${tag_count})`
	const split_cats = cats.split(',')
	const cats_endpoint =
		is_tmap_page && cats_from_user ? `/map/${cats_from_user}` : '/top'

	const expected_like_or_copy_action_status = 204
	async function handle_like() {
		if (!token) {
			save_action_and_path_then_redirect_to_login({
				Action: 'like link',
				LinkID: id,
			})
		}

		const resp = await fetch_with_handle_redirect(
			LINKS_ENDPOINT + `/${id}/like`,
			{
				method: is_liked ? 'DELETE' : 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
			}
		)
		if (!resp.Response || resp.RedirectTo) {
			return (window.location.href = resp.RedirectTo ?? '/500')
		} else if (
			resp.Response.status !== expected_like_or_copy_action_status
		) {
			const like_data = await resp.Response.json()
			if (is_error_response(like_data)) {
				return console.error('Whoops: ', like_data.error)
			}
			return console.error('Whoops: ', like_data)
		}

		if (is_liked) {
			set_is_liked(false)
			set_like_count((prev) => prev - 1)
		} else {
			set_is_liked(true)
			set_like_count((prev) => prev + 1)
		}

		return
	}

	async function handle_copy() {
		if (!token) {
			save_action_and_path_then_redirect_to_login({
				Action: 'copy link',
				LinkID: id,
			})
		}

		const resp = await fetch_with_handle_redirect(
			LINKS_ENDPOINT + `/${id}/copy`,
			{
				method: is_copied ? 'DELETE' : 'POST',
				headers: {
					'Content-Type': 'application/json',
					Authorization: `Bearer ${token}`,
				},
			}
		)
		if (!resp.Response || resp.RedirectTo) {
			return (window.location.href = resp.RedirectTo ?? '/500')
		} else if (
			resp.Response.status !== expected_like_or_copy_action_status
		) {
			const copy_data = await resp.Response.json()
			if (is_error_response(copy_data)) {
				return console.error('Whoops: ', copy_data.error)
			}
			return console.error('Whoops: ', copy_data)
		}

		set_is_copied(is_copied ? false : true)
		return
	}

	const expected_delete_action_status = 205
	async function handle_delete() {
		if (!token) {
			save_path_then_redirect_to_login()
		} else if (!is_your_link) {
			console.error('not your link')
			return
		}

		const delete_resp = await fetch_with_handle_redirect(LINKS_ENDPOINT, {
			method: 'DELETE',
			headers: {
				'Content-Type': 'application/json',
				Authorization: `Bearer ${token}`,
			},
			body: JSON.stringify({ link_id: id }),
		})
		if (!delete_resp.Response || delete_resp.RedirectTo) {
			return (window.location.href = delete_resp.RedirectTo ?? '/500')
		} else if (
			delete_resp.Response.status !== expected_delete_action_status
		) {
			const delete_data = await delete_resp.Response.json()
			if (is_error_response(await delete_data)) {
				return console.error('Whoops: ', delete_data.error)
			}
			return console.error('Whoops: ', delete_resp.Response)
		}

		// redirect to tmap if tag or summary page
		// since those pages no longer exist for the deleted link
		if (is_tag_page || is_summary_page) {
			return (window.location.href = `/map/${user}`)
		}

		// else reload
		return window.location.reload()
	}

	return (
		<li class={`link${is_summary_page || is_tag_page ? ' single' : ''}`}>
			{img_url ? (
				<div class='preview'>
					<img
						src={img_url}
						alt={summary ? summary : url}
						width={100}
					/>
					<div>
						<a href={url} class='url-anchor'>
							<h3>{summary ? summary : url}</h3>
						</a>
						{summary ? <p class='url'>{url}</p> : null}
					</div>
				</div>
			) : (
				<>
					<a href={url} class='url-anchor'>
						<h3>{summary ? summary : url}</h3>
					</a>

					{summary ? <p class='url'>{url}</p> : null}
				</>
			)}

			<p>
				submitted by{' '}
				<a href={`/map/${submitted_by}`} class='submitted-by'>
					{submitted_by}
				</a>{' '}
				on{' '}
				{format_long_date(
					submit_date_in_local_time
						? submit_date_in_local_time
						: submit_date
				)}
			</p>

			{is_tag_page && has_only_your_tag ? null : (
				<p class='tags'>
					<a class='tags-page-link' href={`/tag/${id}`}>
						{tag_attribution}
					</a>
					{': '}
					{split_cats.map((cat, i) =>
						i === split_cats.length - 1 ? (
							<a
								href={
									cats_endpoint +
									`?cats=${encodeURIComponent(cat)}${nsfw_cat_links ? '&nsfw=true' : ''}`
								}
							>
								{cat}
							</a>
						) : (
							<span>
								<a
									href={
										cats_endpoint +
										`?cats=${encodeURIComponent(cat)}${nsfw_cat_links ? '&nsfw=true' : ''}`
									}
								>
									{cat}
								</a>
								,{' '}
							</span>
						)
					)}
					{set_new_link_cats ? (
						<button
							title='Copy cats to pending new link'
							class='img-btn copy-cats-btn'
							onClick={() => set_new_link_cats(split_cats)}
						>
							{/* https://icon-sets.iconify.design/solar/copy-outline/ */}
							<svg
								xmlns='http://www.w3.org/2000/svg'
								viewBox='2.25 1.25 19.5 21.5'
								width='1rem'
								height='1rem'
							>
								{' '}
								<path
									fill='currentColor'
									fill-rule='evenodd'
									d='M15 1.25h-4.056c-1.838 0-3.294 0-4.433.153c-1.172.158-2.121.49-2.87 1.238c-.748.749-1.08 1.698-1.238 2.87c-.153 1.14-.153 2.595-.153 4.433V16a3.75 3.75 0 0 0 3.166 3.705c.137.764.402 1.416.932 1.947c.602.602 1.36.86 2.26.982c.867.116 1.97.116 3.337.116h3.11c1.367 0 2.47 0 3.337-.116c.9-.122 1.658-.38 2.26-.982s.86-1.36.982-2.26c.116-.867.116-1.97.116-3.337v-5.11c0-1.367 0-2.47-.116-3.337c-.122-.9-.38-1.658-.982-2.26c-.531-.53-1.183-.795-1.947-.932A3.75 3.75 0 0 0 15 1.25m2.13 3.021A2.25 2.25 0 0 0 15 2.75h-4c-1.907 0-3.261.002-4.29.14c-1.005.135-1.585.389-2.008.812S4.025 4.705 3.89 5.71c-.138 1.029-.14 2.383-.14 4.29v6a2.25 2.25 0 0 0 1.521 2.13c-.021-.61-.021-1.3-.021-2.075v-5.11c0-1.367 0-2.47.117-3.337c.12-.9.38-1.658.981-2.26c.602-.602 1.36-.86 2.26-.981c.867-.117 1.97-.117 3.337-.117h3.11c.775 0 1.464 0 2.074.021M7.408 6.41c.277-.277.665-.457 1.4-.556c.754-.101 1.756-.103 3.191-.103h3c1.435 0 2.436.002 3.192.103c.734.099 1.122.28 1.399.556c.277.277.457.665.556 1.4c.101.754.103 1.756.103 3.191v5c0 1.435-.002 2.436-.103 3.192c-.099.734-.28 1.122-.556 1.399c-.277.277-.665.457-1.4.556c-.755.101-1.756.103-3.191.103h-3c-1.435 0-2.437-.002-3.192-.103c-.734-.099-1.122-.28-1.399-.556c-.277-.277-.457-.665-.556-1.4c-.101-.755-.103-1.756-.103-3.191v-5c0-1.435.002-2.437.103-3.192c.099-.734.28-1.122.556-1.399'
									clip-rule='evenodd'
								></path>{' '}
							</svg>
						</button>
					) : null}
				</p>
			)}

			{is_summary_page ? null : (
				<p class='summaries'>
					<a href={`/summary/${id}`}>summaries ({summary_count})</a>
				</p>
			)}

			{user !== submitted_by ? (
				<>
					<button
						title='Like link'
						alt='Like link'
						onClick={handle_like}
						class={`like-btn${is_liked ? ' liked' : ''}`}
					>
						{is_liked ? (
							<>
								{/* https://icon-sets.iconify.design/solar/like-bold/ */}
								<svg
									xmlns='http://www.w3.org/2000/svg'
									width='1rem'
									height='1rem'
									viewBox='0 0 24 24'
								>
									<path
										fill='currentColor'
										d='m20.27 16.265l.705-4.08a1.666 1.666 0 0 0-1.64-1.95h-5.181a.833.833 0 0 1-.822-.969l.663-4.045a4.783 4.783 0 0 0-.09-1.973a1.635 1.635 0 0 0-1.092-1.137l-.145-.047a1.346 1.346 0 0 0-.994.068c-.34.164-.588.463-.68.818l-.476 1.834a7.628 7.628 0 0 1-.656 1.679c-.415.777-1.057 1.4-1.725 1.975l-1.439 1.24a1.67 1.67 0 0 0-.572 1.406l.812 9.393A1.666 1.666 0 0 0 8.597 22h4.648c3.482 0 6.453-2.426 7.025-5.735'
									/>
									<path
										fill='currentColor'
										fill-rule='evenodd'
										d='M2.968 9.485a.75.75 0 0 1 .78.685l.97 11.236a1.237 1.237 0 1 1-2.468.107V10.234a.75.75 0 0 1 .718-.749'
										clip-rule='evenodd'
									/>
								</svg>{' '}
								({like_count})
							</>
						) : (
							<>
								{/* https://icon-sets.iconify.design/solar/like-outline/ */}
								<svg
									xmlns='http://www.w3.org/2000/svg'
									width='1rem'
									height='1rem'
									viewBox='0 0 24 24'
								>
									<path
										fill='currentColor'
										fill-rule='evenodd'
										d='M12.438 2.778a.596.596 0 0 0-.438.03a.515.515 0 0 0-.28.33l-.476 1.834a8.378 8.378 0 0 1-.72 1.844c-.485.907-1.218 1.604-1.898 2.19l-1.438 1.24a.918.918 0 0 0-.315.774l.812 9.393a.916.916 0 0 0 .911.837h4.649c3.136 0 5.779-2.182 6.286-5.113l.705-4.08a.916.916 0 0 0-.901-1.073h-5.181c-.977 0-1.72-.876-1.562-1.84l.663-4.044a4.03 4.03 0 0 0-.076-1.664a.885.885 0 0 0-.596-.611zl.23-.714zm-1.09-1.321a2.096 2.096 0 0 1 1.549-.107l.145.047l-.23.714l.23-.714c.777.25 1.383.87 1.589 1.662c.193.746.229 1.524.104 2.284l-.663 4.044a.083.083 0 0 0 .082.097h5.18c1.5 0 2.636 1.352 2.38 2.829l-.705 4.08c-.638 3.688-3.938 6.357-7.764 6.357H8.596a2.416 2.416 0 0 1-2.405-2.208l-.813-9.393a2.418 2.418 0 0 1 .83-2.04l1.44-1.24c.655-.564 1.206-1.111 1.552-1.76a6.83 6.83 0 0 0 .592-1.514l.476-1.833a2.014 2.014 0 0 1 1.08-1.305m-8.38 8.028a.75.75 0 0 1 .78.685l.97 11.236a1.237 1.237 0 1 1-2.468.107V10.234a.75.75 0 0 1 .718-.75'
										clip-rule='evenodd'
									/>
								</svg>{' '}
								({like_count})
							</>
						)}
					</button>

					<button
						title='Copy link to treasure map'
						alt='Copy link to treasure map'
						onClick={handle_copy}
						class={`copy-btn${is_copied ? ' copied' : ''}`}
					>
						{is_copied ? (
							<>
								{/* https://icon-sets.iconify.design/mingcute/copy-fill/ */}
								<svg
									xmlns='http://www.w3.org/2000/svg'
									width='1rem'
									height='1rem'
									viewBox='0 0 24 24'
								>
									<g fill='none'>
										<path d='M24 0v24H0V0zM12.593 23.258l-.011.002l-.071.035l-.02.004l-.014-.004l-.071-.035q-.016-.005-.024.005l-.004.01l-.017.428l.005.02l.01.013l.104.074l.015.004l.012-.004l.104-.074l.012-.016l.004-.017l-.017-.427q-.004-.016-.017-.018m.265-.113l-.013.002l-.185.093l-.01.01l-.003.011l.018.43l.005.012l.008.007l.201.093q.019.005.029-.008l.004-.014l-.034-.614q-.005-.019-.02-.022m-.715.002a.02.02 0 0 0-.027.006l-.006.014l-.034.614q.001.018.017.024l.015-.002l.201-.093l.01-.008l.004-.011l.017-.43l-.003-.012l-.01-.01z' />
										<path
											fill='currentColor'
											d='M19 2a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-2v2a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h2V4a2 2 0 0 1 2-2zm-9 13H8a1 1 0 0 0-.117 1.993L8 17h2a1 1 0 0 0 .117-1.993zm9-11H9v2h6a2 2 0 0 1 2 2v8h2zm-7 7H8a1 1 0 1 0 0 2h4a1 1 0 1 0 0-2'
										/>
									</g>
								</svg>
								{' (Copied)'}
							</>
						) : (
							<>
								{/* https://icon-sets.iconify.design/mingcute/copy-line/ */}
								<svg
									xmlns='http://www.w3.org/2000/svg'
									width='1rem'
									height='1rem'
									viewBox='0 0 24 24'
								>
									<g fill='none'>
										<path d='M24 0v24H0V0zM12.593 23.258l-.011.002l-.071.035l-.02.004l-.014-.004l-.071-.035q-.016-.005-.024.005l-.004.01l-.017.428l.005.02l.01.013l.104.074l.015.004l.012-.004l.104-.074l.012-.016l.004-.017l-.017-.427q-.004-.016-.017-.018m.265-.113l-.013.002l-.185.093l-.01.01l-.003.011l.018.43l.005.012l.008.007l.201.093q.019.005.029-.008l.004-.014l-.034-.614q-.005-.019-.02-.022m-.715.002a.02.02 0 0 0-.027.006l-.006.014l-.034.614q.001.018.017.024l.015-.002l.201-.093l.01-.008l.004-.011l.017-.43l-.003-.012l-.01-.01z' />
										<path
											fill='currentColor'
											d='M19 2a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2h-2v2a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h2V4a2 2 0 0 1 2-2zm-4 6H5v12h10zm-5 7a1 1 0 1 1 0 2H8a1 1 0 1 1 0-2zm9-11H9v2h6a2 2 0 0 1 2 2v8h2zm-7 7a1 1 0 0 1 .117 1.993L12 13H8a1 1 0 0 1-.117-1.993L8 11z'
										/>
									</g>
								</svg>
								{' (Copy)'}
							</>
						)}
					</button>
				</>
			) : (
				<SameUserLikeCount LikeCount={like_count} />
			)}

			{is_your_link ? (
				<>
					{/* delete button */}
					<button
						title='Delete Link'
						alt='Delete Link'
						class='delete-link-btn img-btn'
						onClick={() => set_show_delete_modal(true)}
					>
						<img src='../../../delete.svg' height={20} width={20} />
					</button>

					{show_delete_modal ? (
						<Modal
							Prompt={'Are you sure you want to delete'}
							IsDeleteConfirmation
							DeleteURL={url}
							HandleDelete={handle_delete}
							SetShowModal={set_show_delete_modal}
						/>
					) : null}
				</>
			) : null}
		</li>
	)
}
