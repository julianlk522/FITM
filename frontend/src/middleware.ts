
import { defineMiddleware } from "astro:middleware";
import type { VerifyErrors } from "jsonwebtoken";
import jwt from "jsonwebtoken";

export const onRequest = defineMiddleware((context, next) => {
    const req_token = context.cookies.get('token')?.value
    const req_user = context.cookies.get('user')?.value
    
    // if token cookie found, authenticate
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

    // else if no token cookie found but user cookie found, reset user cookie and redirect to login
    } else if (req_user) {
        context.cookies.delete('user')
        return Response.redirect(new URL("/login", context.request.url), 302)

    } 
    
    // if returning to a page that requires login, and redirect_to cookie found, redirect there after authentication
    // TODO: replace localhost URL with production URL
    const redirect_from_login = context.request.headers.get('referer') === 'http://localhost:4321/login'
    if (redirect_from_login) {

        // get redirect URL from cookie
        const redirect_url = context.cookies.get('redirect_to')?.value
        const same_url = redirect_url === context.request.url
        if (redirect_url && !same_url) {
            context.cookies.delete('redirect_to')
            return Response.redirect(new URL(redirect_url, context.request.url), 302)
        }
    }

    // if not redirecting from login and no token or user cookie found, continue
    return next();
});