$.fn.dataTableExt.oPagination.gamedb = function (page, pages) {
    return ['previous', _numbers(page, pages), 'next'];
};

function _numbers(page, pages) {

    let numbers;
    const buttons = 7;
    const half = Math.floor(buttons / 2);

    if (pages <= buttons) {
        numbers = _range(0, pages);
    } else if (page <= half) {
        numbers = _range(0, buttons - 2);
        numbers.push('ellipsis');
        // numbers.push(pages - 1);
    } else if (page >= pages - 1 - half) {
        numbers = _range(pages - (buttons - 2), pages);
        numbers.splice(0, 0, 'ellipsis'); // no unshift in ie6
        numbers.splice(0, 0, 0);
    } else {
        numbers = _range(page - half + 2, page + half - 1);
        numbers.push('ellipsis');
        // numbers.push(pages - 1);
        numbers.splice(0, 0, 'ellipsis');
        numbers.splice(0, 0, 0);
    }

    numbers.DT_el = 'span';
    return numbers;
}

function _range(len, start) {

    const out = [];
    let end;

    end = start;
    start = len;

    for (let i = start; i < end; i++) {
        out.push(i);
    }

    return out;
}
