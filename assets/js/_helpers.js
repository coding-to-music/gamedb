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

function toast(success = true, body, title = '', timeout = 0, link = '') {

    const redirect = function () {
        if (link === 'refresh') {
            link = window.location.href;
        }
        window.location.replace(link);

    };

    const options = {
        timeOut: Number(timeout > 0 ? timeout : 8) * 1000,
        onclick: link ? redirect : null,

        newestOnTop: true,
        preventDuplicates: false,
        extendedTimeOut: 0, // Keep alive on hover
    };

    if (success) {
        toastr.success(body, title, options);
    } else {
        toastr.error(body, title, options);
    }

}

function highLightOwnedGames() {
    if (user.isLoggedIn) {
        let games = localStorage.getItem('games');
        if (games != null) {
            games = JSON.parse(games);
            if (games != null) {
                $('[data-app-id]').each(function () {
                    const id = $(this).attr('data-app-id');
                    if (games.indexOf(parseInt(id)) !== -1) {
                        $(this).addClass('font-weight-bold')
                    }
                });
            }
        }
    }
}
