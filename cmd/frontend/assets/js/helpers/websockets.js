let _websocket;

function websocketListener(page, onMessage, attempt = 1) {

    // Checks
    if (window.WebSocket === undefined) {
        if (user.log) {
            log.console('Your browser does not support websockets');
        }
        return;
    }

    if (_websocket && _websocket.readyState === WebSocket.OPEN) {
        logLocal('Websocket already open');
        return;
    }

    // Connect
    _websocket = new WebSocket((location.protocol === 'https:' ? 'wss://' + window.location.hostname : 'ws://' + location.host) + '/websocket/' + page);

    let $badge = $('#live-badge');

    _websocket.onopen = function (e) {

        logLocal('websocket opened');

        $badge.addClass('badge-success cursor-pointer');
        $badge.removeClass('badge-secondary badge-danger');

        if (attempt > 1) {
            // toast(true, 'Some events may have been missed', 'Live functionality is back');
        }

        attempt = 1;
    };

    _websocket.onclose = function (e) {
        logLocal('Websocket closed', e);
        closeWebsocket($badge, attempt, e, 'onclose');
    };

    _websocket.onerror = function (e) {
        logLocal('Websocket error', e);
        closeWebsocket($badge, attempt, e, 'onerror');
    };

    _websocket.onmessage = function (e) {
        logLocal('WS: ' + e.data);
        return onMessage(e);
    };

    // Click to open/close websocket
    if (attempt === 1) {
        $badge.on('click', function (e) {

            // Open
            if ($(this).hasClass('badge-danger')) {

                logLocal('Websocket opened manually');

                websocketListener(page, onMessage, 2);

                $badge.addClass('badge-success');
                $badge.removeClass('badge-secondary badge-danger');

            } else if ($(this).hasClass('badge-success')) {

                logLocal('Websocket closed manually', e);
                if (_websocket !== null) {
                    _websocket.close(1000);
                }
                e.code = 1000;
                closeWebsocket($badge, 1, e, 'manual');
            }
        });
    }

    const closeWebsocket = function closeWebsocket($badge, attempt, e, type) {

        logLocal(type, attempt);

        _websocket = null;

        $badge.addClass('badge-danger');
        $badge.removeClass('badge-secondary badge-success');

        if (type === 'onclose' || type === 'manual') {

            if (attempt === 1) {
                // toast(false, 'Live functionality has stopped');
            }

            if (e.code !== 1000) {

                setTimeout(function () {
                    websocketListener(page, onMessage, attempt + 1);
                }, 5000);
            }
        }
    };
}
