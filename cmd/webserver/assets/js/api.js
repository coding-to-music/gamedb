const $apiPage = $('#api-page');

if ($apiPage.length > 0) {

    $('#sidebar').stickySidebar({
        topSpacing: 0,
        bottomSpacing: 16,
    });

    $('.endpoint').on('mouseenter', function () {
        $(this).trigger('select');
    });

    $('button[type=submit]').on('click', function (e) {

        const form = $(this).closest('form');

        // Reset names
        form.find('input[data-name]').each(function (index) {
            $(this).attr('name', $(this).attr('data-name'));
        });

        // Validate fields
        form.find('input:not([type=submit]):not(.endpoint)').each(function (index) {
            if (!$(this).val() && $(this).prop('required')) {
                $(this).addClass('is-invalid');
            } else {
                $(this).removeClass('is-invalid');
            }
        });

        // Remove empty input from submit
        form.find('input.form-control:not(.endpoint)').each(function (index) {
            if ($(this).val() === '') {
                $(this).attr('name', '');
            }
        });

        // Replace variables in form action
        let action = form.find('.endpoint').attr('value');
        action = unescape(action);

        // noinspection RegExpRedundantEscape
        for (const match of action.matchAll(/\{([a-z]+)\}/g)) {
            const $field = form.find('input[data-name="' + match[1] + '"][data-location=path]');
            if ($field.length && $field.val()) {
                action = action.replace(match[0], $field.val());
                $field.removeAttr('name') // Stop it sending query param as well as path param
            }
        }
        form.attr('action', action);
    });
}
