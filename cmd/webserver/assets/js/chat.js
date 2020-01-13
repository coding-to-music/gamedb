if ($('#chat-page').length > 0) {

    const channel = $('[data-channel-id]').attr('data-channel-id');

    $.ajax({
        url: '/chat/' + channel + '/chat.json',
        dataType: 'json',
        cache: false,
        success: function (data, textStatus, jqXHR) {

            $('.fa-spin').remove();

            if (isIterable(data)) {
                for (const v of data) {
                    chatRow(v, false);
                }
            }
        },
    });

    websocketListener('chat', function (e) {

        const data = JSON.parse(e.data);
        chatRow(data.Data);
    });

    function chatRow(data, addToTop = true) {

        const $container = $('ul[data-channel-id=' + data.channel + ']');

        const fadeClass = (addToTop ? ' fade-green' : '');

        if (!data.content && data.embeds) {
            data.content = '<small>Content not available via our website.</small>';
        }

        $container.json2html(
            data,
            {
                '<>': 'li', 'class': 'media fade-in', 'style': 'animation-delay: ${i}s', 'html': [
                    {'<>': 'img', 'class': 'mr-3 rounded', 'src': 'https://cdn.discordapp.com/avatars/${author_id}/${author_avatar}.png?size=64', 'alt': '${author_user}'},
                    {
                        '<>': 'div', 'class': 'media-body', 'html': [
                            {
                                '<>': 'h5', 'class': 'mt-0 mb-1 rounded' + fadeClass, 'html': '${content}'
                            },
                            //{'<>': 'p', 'class': 'text-muted', 'html': 'By ${author_user} at <span data-livestamp="${timestamp}"></span>'}
                            {'<>': 'p', 'class': 'text-muted', 'html': 'By ${author_user}'}
                        ]
                    }
                ]
            },
            {
                prepend: addToTop,
            }
        );

        $container.find('li').slice(50).remove();
    }
}
