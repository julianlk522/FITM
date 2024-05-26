
import { defineMiddleware } from "astro:middleware";
import jwt from "jsonwebtoken";
import get_cookie from "./util/get_cookie";

export const onRequest = defineMiddleware((context, next) => {

    // TODO: replace localhost URLs with production URLs
    if (context.request.url === 'http://localhost:4321/' && context.request.headers.get('Referer') === 'http://localhost:4321/login') {
        let cookie_string = context.request.headers.get('cookie')
        if (cookie_string) {
            const token = get_cookie('token', cookie_string)
            if (!token) {
                return next();
            }
            
            // check if jwt token is valid
            jwt.verify(token, 'secret', function (err: any, decoded) {
                // @ts-ignore
                if (!err && decoded.login_name) {
                    // @ts-ignore
                    context.locals.user = decoded.login_name
                }
            })
        }   
    }
    return next();
});