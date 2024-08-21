import { useState } from 'preact/hooks'
import * as types from '../../types'
import { format_long_date } from '../../util/format_date'
import './Link.css'

interface Props {
	Link: types.Link
	CatsFromUser?: string
	IsSummaryPage: boolean
	IsTagPage: boolean
	Token: string | undefined
	User: string | undefined
}

export default function Link(props: Props) {
	const {
		CatsFromUser: cats_from_user,
		IsSummaryPage: is_summary_page,
		IsTagPage: is_tag_page,
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
		ImgURL: img_url,
		// IsTagged: is_tagged,
	} = props.Link

	const [is_copied, set_is_copied] = useState(props.Link.IsCopied)
	const [is_liked, set_is_liked] = useState(props.Link.IsLiked)
	const [like_count, set_like_count] = useState(props.Link.LikeCount)

	const split_cats = cats?.split(',')
	let tag_attribution =
		cats && user && cats_from_user === user
			? 'Your Tag'
			: cats_from_user
			? `${cats_from_user}'s Tag`
			: tag_count === 1
			? `${submitted_by}'s Tag`
			: 'Global Tag'
	tag_attribution += ` (${tag_count})`
	const cats_html =
		// depending on if tmap page, link to tmap subcats page or global cats page
		cats_from_user
			? // tag1 ==> <a href='/map/user/cat/tag1'>tag1</a>
			  // tag1,tag2 ==> <a href='/map/user/tag1'>tag1</a>, <a href='/map/user/tag2'>tag2</a>
			  split_cats?.map((cat, i) => {
					if (i === split_cats.length - 1) {
						return (
							<a href={`/map/${cats_from_user}?cats=${cat}`}>
								{cat}
							</a>
						)
					} else {
						return (
							<span>
								<a href={`/map/${cats_from_user}?cats=${cat}`}>
									{cat}
								</a>
								,{' '}
							</span>
						)
					}
			  })
			: // tag1 ==> <a href='/cat/tag1'>tag1</a>
			  // tag1,tag2 ==> <a href='/cat/tag1'>tag1</a>, <a href='/cat/tag2'>tag2</a>
			  split_cats?.map((cat, i) => {
					if (i === split_cats.length - 1) {
						return <a href={`/top?cats=${cat}`}>{cat}</a>
					} else {
						return (
							<span>
								<a href={`/top?cats=${cat}`}>{cat}</a>,{' '}
							</span>
						)
					}
			  })

	async function handle_like() {
		if (!token) {
			document.cookie = `redirect_to=${window.location.pathname.replaceAll(
				'/',
				'%2F'
			)}; path=/login; max-age=21600; SameSite=strict; Secure`
			document.cookie = `redirect_action=like link ${id}; path=${window.location.pathname}; max-age=21600; SameSite=strict; Secure`
			return (window.location.href = '/login')
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
				console.error('WTF is this: ', like_data)
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
				console.error('WTF is this: ', unlike_data)
			}
		}
	}

	async function handle_copy() {
		if (!token) {
			document.cookie = `redirect_to=${window.location.pathname.replaceAll(
				'/',
				'%2F'
			)}; path=/login; max-age=21600; SameSite=strict; Secure`
			document.cookie = `redirect_action=copy link ${id}; path=${window.location.pathname}; max-age=21600; SameSite=strict; Secure`
			return (window.location.href = '/login')
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
				console.error('WTF is this: ', copy_data)
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
				console.error('WTF is this: ', uncopy_data)
			}
		}
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
				Submitted by{' '}
				<a href={`/map/${submitted_by}`} class='submitted-by'>
					{submitted_by}
				</a>{' '}
				on {format_long_date(submit_date)}
			</p>

			{is_tag_page && tag_count === 1 && submitted_by === user ? null : (
				<p class='tags'>
					<a class='tags-page-link' href={`/tag/${id}`}>
						{tag_attribution}
					</a>
					{': '}
					{cats_html}
				</p>
			)}

			{is_summary_page ? null : (
				<p class='summaries'>
					<a href={`/summary/${id}`}>Summaries ({summary_count})</a>
				</p>
			)}

			{user !== submitted_by ? (
				<>
					<button
						onClick={handle_like}
						class={`like-btn${is_liked ? ' liked' : ''}`}
						alt='Like this link'
					>
						{is_liked ? (
							<>
								{/* https://icon-sets.iconify.design/solar/like-bold/ */}
								<svg
									xmlns='http://www.w3.org/2000/svg'
									width='1em'
									height='1em'
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
									width='1em'
									height='1em'
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
						onClick={handle_copy}
						class={`copy-btn${is_copied ? ' copied' : ''}`}
						alt='Copy to treasure map'
					>
						{is_copied ? (
							<>
								{/* https://icon-sets.iconify.design/mingcute/copy-fill/ */}
								<svg
									xmlns='http://www.w3.org/2000/svg'
									width='1em'
									height='1em'
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
									width='1em'
									height='1em'
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
				<div class='same-user-like-count'>
					{/* https://icon-sets.iconify.design/solar/like-outline/ */}
					<svg
						xmlns='http://www.w3.org/2000/svg'
						width='1em'
						height='1em'
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
				</div>
			)}
		</li>
	)
}
