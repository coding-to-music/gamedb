if ($('#chat-bot-page').length > 0) {

    const $container = $('table#recent tbody');

    $.ajax({
        url: '/chat-bot/commands.json',
        dataType: 'json',
        cache: false,
        success: function (data, textStatus, jqXHR) {

            if (isIterable(data)) {
                for (const message of data) {
                    messageRow(message, false);
                }
            }
        },
    });

    websocketListener('chat-bot', function (e) {

        const data = JSON.parse(e.data);
        toast(true, data.Data, '', 2);
        messageRow(data.Data);
    });

    function messageRow(message, addToTop = true) {

        const fadeClass = (addToTop ? ' fade-green' : '');

        $container.json2html(
            {message: message},
            {
                '<>': 'tr', 'html': [
                    {'<>': 'td', 'html': '${message}'}
                ],
            },
            {
                prepend: addToTop,
            }
        );

        $container.find('row').slice(2).remove();
    }
}
