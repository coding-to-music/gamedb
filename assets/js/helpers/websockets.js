function websocketListener(page, onMessage) {

    if (window.WebSocket === undefined) {

        console.log('Your browser does not support websockets');

    } else {

        const socket = new WebSocket((location.protocol === 'https:' ? "wss://gamedb.online" : "ws://" + location.host) + "/websocket/" + page);
        const $badge = $('#live-badge');
        let open = false;

        socket.onopen = function (e) {
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger');
            console.log('Websocket opened');
            open = true;
        };

        socket.onclose = function (e) {
            if (open) {
                $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
                toast(false, 'Live functionality has stopped'); // onerror will trigger too
                console.log('Websocket closed');
            }
        };

        socket.onerror = function (e) {
            if (open) {
                $badge.addClass('badge-danger').removeClass('badge-secondary badge-success');
                toast(false, 'Live functionality has stopped');
            }
        };

        socket.onmessage = function (e) {

            if (user.isLocal) {
                console.log('WS: ' + e.data);
            }

            return onMessage(e)
        };

        // Click to close websocket manually
        // $badge.on('click', function (e) {
        //     if ($(this).hasClass('cursor-pointer')) {
        //         socket.close(1000);
        //         $badge.addClass('badge-danger').removeClass('badge-secondary badge-success cursor-pointer');
        //         toast(false, 'Live functionality has stopped');
        //     }
        // });
    }
}