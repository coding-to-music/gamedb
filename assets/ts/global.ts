interface Person {
    userID: string;
    userLevel: number;
    userName: string;
    userEmail: string;
    isLoggedIn: boolean;
    isLocal: boolean;
    isAdmin: boolean;
    showAds: boolean;
    country: string;
    currencySymbol: string;
    flashesGood: Array<string>;
    flashesBad: Array<string>;
    toasts: Array<Toast>;
    session: StringMap;
}

interface Toast {
    Title: string;
    Message: string;
    Link: string;
    Theme: string;
    Timeout: number;
}

interface StringMap {
    [index: string]: string;
}
