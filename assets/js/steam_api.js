const $steamApiPage = $('#steam-api-page');

if ($steamApiPage.length > 0) {

    $('#sidebar').stickySidebar({
        topSpacing: 0,
        bottomSpacing: 16,
    });

    $('.endpoint').on('mouseenter', function () {
        $(this).select();
    });

    const $form = $steamApiPage.find('form#key-form');

    $form.on('submit', function (e) {

        e.preventDefault();
        localStorage.setItem('settings', $form.serialize());
        setMethodSettings();
        toast(true, 'Settings Saved');
    });

    function setMethodSettings() {

        const key = $('#key-form input[name=key]').val();
        const format = $('#key-form select[name=format]').val();

        if (key) {
            $steamApiPage.find('table').show();
        } else {
            $steamApiPage.find('table').hide();
        }

        $('div.interface input[name=key]').val(key);
        $('div.interface input[name=format]').val(format);
    }

    $form.deserialize(localStorage.getItem('settings'));
    setMethodSettings();
}
