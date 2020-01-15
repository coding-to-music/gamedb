function toast(success = true, body, title = '', timeout = 0, link = '') {

    const redirect = function () {
        if (link === 'refresh') {
            location.reload();
        }
    };

    if (timeout === 0) {
        timeout = 8;
    }

    if (timeout === -1) {
        timeout = 0;
    }

    const options = {
        onclick: link ? redirect : null,
        newestOnTop: true,
        preventDuplicates: false,
        progressBar: true,
        timeOut: timeout * 1000,
        extendedTimeOut: timeout * 1000
    };

    if (isMobile()) {
        options["positionClass"] = "toast-bottom-right";
    } else {
        options["positionClass"] = "toast-top-right";
    }

    if (success) {
        toastr.success(body, title, options);
    } else {
        toastr.error(body, title, options);
    }
}
