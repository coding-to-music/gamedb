if ($('#chat-bot-page').length > 0) {

    const $container = $('table#recent tbody');

    $.ajax({
        url: '/discord-bot/commands.json',
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
            message,
            {
                '<>': 'tr', 'html': [
                    {
                        '<>': 'td', 'class': 'img thin', 'html': [
                            {
                                '<>': 'div', 'class': 'icon-name', 'html': [
                                    {
                                        '<>': 'div', 'class': 'icon', 'html': [{'<>': 'img', 'data-lazy': 'https://cdn.discordapp.com/avatars/${author_id}/${author_avatar}.png?size=64', 'alt': '', 'data-lazy-alt': '${author_name}'}],
                                    },
                                    {
                                        '<>': 'div', 'class': 'name nowrap', 'html': '${author_name}',
                                    }
                                ]
                            }
                        ]
                    },
                    {'<>': 'td', 'html': '${message}'}
                ],
            },
            {
                prepend: addToTop,
            }
        );

        $container.find('row').slice(2).remove();
        observeLazyImages($container.find('img[data-lazy]'));
        fixBrokenImages();
    }
}
