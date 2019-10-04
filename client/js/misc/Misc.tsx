export enum FormState {
    Init,
    Invalid,
    Ok,
    Finished
}

export function mapTextToError(txt: string, errorEnum: any, errorMap: any): string {
    let errorText = "";
    Object.keys(errorEnum).find(key => {
        let val: string = errorEnum[key];
        if (typeof key !== 'string') {
            return false;
        }
        if (txt.includes(val)) {
            errorText = errorMap[val];
            return true;
        }
        return false;
    });
    if (!errorText) {
        return errorMap[errorEnum.Unknown];
    }
    return errorText;
}

export enum CreateError {
    Unknown = "unknown",
    AlreadyExists = "duplicate key",
    NotFound = "404"
}