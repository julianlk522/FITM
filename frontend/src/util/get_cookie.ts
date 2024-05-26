export default function get_cookie(name: string, cookie_string: string) {
    name += "=";
    let decoded_cookie = decodeURIComponent(cookie_string);
    let ca = decoded_cookie.split(';');
    for(let i = 0; i <ca.length; i++) {
      let c = ca[i];
      while (c.charAt(0) == ' ') {
        c = c.substring(1);
      }
      if (c.indexOf(name) == 0) {
        return c.substring(name.length, c.length);
      }
    }
    return "";
}