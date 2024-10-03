export function assign_login_redirect_handler() {
	document
		?.getElementById('login-redirect-link')
		?.addEventListener('click', save_path_then_redirect_to_login)
}

export function save_action_and_path_then_redirect_to_login(
	redirect_action: RedirectAction
) {
	// store authorized action to complete after login as cookie
	const action_target_ID = is_link_redirect_action(redirect_action)
		? redirect_action.LinkID
		: redirect_action.SummaryID

	document.cookie = `redirect_action=${redirect_action.Action} ${action_target_ID}; path=${window.location.pathname}; max-age=300; SameSite=strict; Secure`

	return save_path_then_redirect_to_login()
}

export function save_path_then_redirect_to_login() {
	// store path to return to after login as cookie
	document.cookie = `redirect_to=${window.location.pathname.replaceAll('/', '%2F')}; path=/login; max-age=300; SameSite=strict; Secure`

	return (window.location.href = '/login')
}

function is_link_redirect_action(
	action: RedirectAction
): action is LinkRedirectAction {
	return (action as LinkRedirectAction).LinkID !== undefined
}

type RedirectAction = LinkRedirectAction | SummaryRedirectAction

type LinkRedirectAction = {
	Action: 'like link' | 'copy link'
	LinkID: string
}

type SummaryRedirectAction = {
	Action: 'like summary'
	SummaryID: string
}
