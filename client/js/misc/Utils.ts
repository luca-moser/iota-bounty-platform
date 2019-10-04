const urlRegex = new RegExp('^(https?:\\/\\/)?' + // protocol
    '((([a-z\\d]([a-z\\d-]*[a-z\\d])*)\\.?)+[a-z]{2,}|' + // domain name
    '((\\d{1,3}\\.){3}\\d{1,3}))' + // ip (v4) address
    '(\\:\\d+)?(\\/[-a-z\\d%_.~+]*)*' + //port
    '(\\?[;&amp;a-z\\d%_.~+=-]*)?' + // query string
    '(\\#[-a-z\\d_]*)?$', 'i');

const githubFrag = "github.com";

export function isValidGitHubURL(s: string): boolean {
    if (!urlRegex.test(s)) return false;
    if (s.indexOf(githubFrag) === -1) return false;
    return true;
}

const emailRegex = /^(([^<>()\[\]\\.,;:\s@"]+(\.[^<>()\[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/;

export function isValidEmail(email: string) {
    return emailRegex.test(String(email).toLowerCase());
}

const whitespaceRegex = /^\S*$/;

export function hasNoWhitespace(s: string) {
    return whitespaceRegex.test(s);
}

enum LinkKeys {
    Timeout = "timeout_at",
    MultiUse = "multi_use",
    ExpectedAmount = "expected_amount"
}

export enum FetchConst {
    ContentType = "Content-Type",
    Authorization = "Authorization",
    JSONContent = "application/json",
}

export function fetchOpts(payload: string, token?: string): RequestInit {
    let req = {
        method: "POST",
        headers: {
            [FetchConst.ContentType]: FetchConst.JSONContent,
        },
        body: payload,
    };
    if (token) {
        req.headers[FetchConst.Authorization] = `Bearer ${token}`;
    }
    return req;
}

export enum Routes {
    LOGIN = "/user/login",
    REGISTER = "/user/id"
}