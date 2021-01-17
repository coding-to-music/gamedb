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
            data.content = '<small>Media content not available on our website, please view in Discord</small>';
        }

        $container.json2html(
            data,
            {
                '<>': 'li', 'class': 'media fade-in', 'style': 'animation-delay: ${i}s', 'html': [
                    {
                        '<>': 'img', 'class': 'mr-3 rounded', 'alt': '', 'data-lazy-alt': '${author_user}', 'data-lazy': function (obj, index) {
                            return obj.author_avatar ? 'https://cdn.discordapp.com/avatars/' + obj.author_id + '/' + obj.author_avatar + '.png?size=64' : '/assets/img/no-app-image-square.jpg';
                        }
                    },
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
        observeLazyImages($container.find('img[data-lazy]'));
    }
}
