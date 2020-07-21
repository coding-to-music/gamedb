function toast(success = true, body, title = '', timeout = 0, link = '') {

    const redirect = function () {
        if (link === 'refresh') {
            location.reload();
        }
    };

    // Default time
    if (timeout === 0) {
        timeout = 5;
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
        options["newestOnTop"] = true;
    } else {
        options["positionClass"] = "toast-top-right";
    }

    if (success) {
        return toastr.success(body, title, options);
    } else {
        return toastr.error(body, title, options);
    }
}
