
import type { APIContext } from "astro";
import { sequence } from "astro:middleware";
import type { VerifyErrors } from "jsonwebtoken";
import jwt from "jsonwebtoken";

export const onRequest = sequence(handle_jwt_auth, handle_redirect_action);

async function handle_jwt_auth(context: APIContext, next: () => Promise<Response>) {
    const req_token = context.cookies.get('token')?.value
    const req_user = context.cookies.get('user')?.value
    
    // authenticate token cookie if found
    if (req_token) { 
        try {
            jwt.verify(req_token, 'secret', function<JwtPayload> (err: VerifyErrors | null, decoded: JwtPayload | undefined) {

                // @ts-ignore
                if (err || !decoded || !decoded.login_name) {
                    context.cookies.delete('token')
                    context.cookies.delete('user')

                    return Response.redirect(new URL("/login", context.request.url), 302);
                
                // set user cookie if verified
                } else {
                    
                    // @ts-ignore
                    context.cookies.set('user', decoded.login_name, {path: '/', maxAge: 3600, sameSite: 'strict', secure: true})
                }
            })
        } catch(err) {
            console.log("jwt errors: ", err)
            return Response.redirect(new URL("/login", context.request.url), 302);
        }

    // if user cookie found but no token, reset user cookie and redirect to login
    } else if (req_user) {
        context.cookies.delete('user')
        return Response.redirect(new URL("/login", context.request.url))

    }

    return next();
}

async function handle_redirect_action(context: APIContext, next: () => Promise<Response>) {
    const redirect_action = context.cookies.get('redirect_action')?.value
     // user may be in the redirect process; don't delete cookie if so
    // otherwise delete cookie so no accidental actions
    if (!redirect_action || context.request.url === "http://127.0.0.1:4321/login" || context.request.url === "http://localhost:4321/login") {
        return next()
    }

    const token = context.cookies.get('token')?.value
    if (!token) {
        context.cookies.delete('redirect_action')
        return next()
    }
    
    // e.g., "like summary 78" or "copy summary 78"
    const action = redirect_action.split(' ')[0]
    const item = redirect_action.split(' ')[1]
    let api_section
    if (item === 'summary') {
        api_section = 'summaries'
    } else if (item === 'link') {
        api_section = 'links'
    }
    const item_id = redirect_action.split(' ')[2]

    // LINKS
    // r.Post("/links/{link_id}/like", handler.LikeLink)
    // r.Post("/links/{link_id}/copy", handler.CopyLink)

    // SUMMARIES
    // r.Post("/summaries/{summary_id}/like", handler.LikeSummary)

    const api_url = 'http://127.0.0.1:8000'
    const redirect_action_url = `${api_url}/${api_section}/${item_id}/${action}`

    const resp = await fetch(
        redirect_action_url,
        {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                Authorization: `Bearer ${token}`,
            },
        }
    )

    if (resp.status !== 200) {
        console.error("redirect action failed")
    } else {
        context.cookies.delete('redirect_action', {path: context.url.pathname})
    }

    return next()
}