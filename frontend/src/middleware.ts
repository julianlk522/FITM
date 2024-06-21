
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
        return Response.redirect(new URL("/login", context.request.url))

    }

    // if redirect_to cookie found outside of login page, delete it
    // (should only be read from login handler)
    if (context.request.url !== 'http://127.0.0.1:4321/login' && context.cookies.get('redirect_to')) {
        context.cookies.delete('redirect_to')
    }

    // if not redirecting from login and no token or user cookie found, continue
    return next();
});