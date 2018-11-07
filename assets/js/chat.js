if ($('#chat-page').length > 0) {

    if (window.WebSocket === undefined) {
        console.log('Your browser does not support WebSockets');
    } else {

        const socket = new WebSocket('ws://' + location.host + "/websocket/chat");
        const $badge = $('#live-badge');

        socket.onopen = function (e) {
            // console.log('WebSocket opened');
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger')
        };
        socket.onclose = function (e) {
            // console.log('WebSocket closed');
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success')
        };
        socket.onmessage = function (e) {
            // console.log('WebSocket recieved');
            // console.log(e.data);

            let data = jQuery.parseJSON(e.data);

            if (data.Page === 'chat') {

                data = data.Data;

                $('ul.list-unstyled').prepend($(
                    '<li class="media">' +
                    '    <img class="mr-3" src="https://cdn.discordapp.com/avatars/' + data.author_id + '/' + data.author_avatar + '.png?size=64" alt="' + data.author_user + '">' +
                    '    <div class="media-body">' +
                    '        <h5 class="mt-0 mb-1">' + data.content + '</h5>' +
                    '        <p class="text-muted">By ' + data.author_user + '</p>' +
                    '    </div>' +
                    '</li>'));

            }
        };
        socket.onerror = function (e) {
            console.log('WebSocket error');
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success')
        };
    }
}
