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
