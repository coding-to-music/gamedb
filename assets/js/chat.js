if ($('#chat-page').length > 0) {

    const channel = $('[data-channel-id]').attr('data-channel-id');

    $.ajax({
        url: '/chat/' + channel + '/ajax',
        success: function (data, textStatus, jqXHR) {
            $('.fa-spin').remove();
            if (isIterable(data)) {
                for (const v of data) {
                    chatRow(v, false);
                }
            }
        },
        dataType: 'json',
        cache: false
    });

    websocketListener('chat', function (e) {

        const data = $.parseJSON(e.data);
        chatRow(data.Data);
    });

    function chatRow(data, addToTop = true) {

        const $container = $('ul[data-channel-id=' + data.channel + ']');

        $container.json2html(
            data,
            {
                '<>': 'li', 'class': 'media', 'html': [
                    {'<>': 'img', 'class': 'mr-3 rounded', 'src': 'https://cdn.discordapp.com/avatars/${author_id}/${author_avatar}.png?size=64', 'alt': '${author_user}'},
                    {
                        '<>': 'div', 'class': 'media-body', 'html': [
                            {
                                '<>': 'h5', 'class': function () {
                                    return 'mt-0 mb-1 rounded' + (addToTop ? ' fade-green' : '');
                                }, 'html': '${content}'
                            },
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
