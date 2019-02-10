const $steamApiPage = $('#steam-api-page');

if ($steamApiPage.length > 0) {

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

        const key = $('input[name=key]').val();

        if (key) {
            $steamApiPage.find('table').show();
        }else{
            $steamApiPage.find('table').hide();
        }

        $('input[name=method-key]').val(key);
        $('input[name=method-format]').val($('select[name=format]').val());
    }

    $form.deserialize(localStorage.getItem('settings'));
    setMethodSettings();
}
