function isIterable(obj) {
    // checks for null and undefined
    if (obj == null) {
        return false;
    }
    return typeof obj[Symbol.iterator] === 'function';
}

function isNumeric(n) {
    return !isNaN(parseFloat(n)) && isFinite(n);
}

function isJson(str) {
    try {
        JSON.parse(str);
    } catch (e) {
        return false;
    }
    return true;
}

// Local logging
function logLocal() {
    if (user.log) {
        console.log(...arguments);
    }
}

function setUrlParam(key, value) {

    const url = new URL(window.location);
    url.searchParams.set(key, value);
    url.searchParams.sort();
    window.history.pushState(null, null, url.pathname + url.search + url.hash);
}

function deleteUrlParam(key) {

    const url = new URL(window.location);
    url.searchParams.delete(key)
    window.history.pushState(null, null, url.pathname + url.search + url.hash);
}

function clearUrlParams() {
    window.history.pushState(null, null, window.location.pathname + window.location.hash);
}
