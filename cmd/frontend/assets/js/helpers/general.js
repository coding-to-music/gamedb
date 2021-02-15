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

    if (Array.isArray(value)) {
        url.searchParams.delete(key);
        for (const v of value) {
            url.searchParams.append(key, v);
        }
    } else {
        url.searchParams.set(key, value);
    }

    url.searchParams.sort();
    window.history.replaceState(null, null, url.pathname + url.search + url.hash);
}

function deleteUrlParam(key) {

    const url = new URL(window.location);
    url.searchParams.delete(key);
    window.history.replaceState(null, null, url.pathname + url.search + url.hash);
}

function clearUrlParams() {
    window.history.replaceState(null, null, window.location.pathname + window.location.hash);
}

function ordinal(i) {

    let j = i % 10;
    let k = i % 100;

    if (j === 1 && k !== 11) {
        return i.toLocaleString() + "st";
    }
    if (j === 2 && k !== 12) {
        return i.toLocaleString() + "nd";
    }
    if (j === 3 && k !== 13) {
        return i.toLocaleString() + "rd";
    }
    return i.toLocaleString() + "th";
}

function pad(n, width, z) {
    z = z || '0';
    n = n + '';
    return n.length >= width ? n : new Array(width - n.length + 1).join(z) + n;
}

function serialiseTable(searchFields, order) {

    const obj = {};
    $(searchFields).each(function (index, $field) {
        const name = $field.attr('name') ? $field.attr('name') : $field.attr('data-name');
        obj[name] = $field.val();
    });

    obj.order = order;

    return obj
}
