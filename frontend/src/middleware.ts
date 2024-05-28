
import { defineMiddleware } from "astro:middleware";
import type { VerifyErrors } from "jsonwebtoken";
import jwt from "jsonwebtoken";
import get_cookie from "./util/get_cookie";

export const onRequest = defineMiddleware((context, next) => {

    // when redirecting from login:
    // TODO: replace localhost URLs with production URLs
    if (context.request.url === 'http://localhost:4321/' && context.request.headers.get('Referer') === 'http://localhost:4321/login') {

        // get JWT token cookie from request headers
        let cookie_string = context.request.headers.get('cookie')
        if (cookie_string) {
            const token = get_cookie('token', cookie_string)
            if (!token) {
                return next();
            }
            
            // verify token
            jwt.verify(token, 'secret', function<JwtPayload> (err: VerifyErrors | null, decoded: JwtPayload | undefined) {
                if (decoded && !err) {

                    // add token and user cookies if verified
                    // @ts-ignore
                    context.locals.user = decoded.login_name
                }
            })
        }   
    }
    // if not redirecting from login or no token or token expired/unverified, continue unauthenticated
    return next();
    // (Layout.astro will run an additional JWT check and reset the cookie if expired/unverified. It cannot be performed in this middleware since not .astro)
});