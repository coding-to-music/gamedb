if ($('#changes-page').length > 0) {

    // var data = [
    //     {'s': 'json2html'},
    //     {'s': 'is'},
    //     {'s': 'awesome'}
    // ];
    //
    // var transform = {'<>': 'li', 'html': '${s}'};
    //
    // // $('ul#test').json2html(data, transform);
    // var html = json2html.transform(data, transform);
    // console.log(html);


    if (window.WebSocket === undefined) {
        console.log('Your browser does not support WebSockets');
    } else {
        var socket = new WebSocket("ws://" + location.host + "/websocket");
        var $badge = $('#live-badge');

        socket.onopen = function (e) {
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger')
        };
        socket.onclose = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success')
        };
        socket.onmessage = function (e) {
            var data = jQuery.parseJSON(e.data);

            if (data.Page === 'changes') {

                data = data.Data;

                // console.log(data);

                $('ul.list-unstyled').prepend($(
                    '<li class="media">' +
                    '    <div class="media-body">' +
                    '        <h5 class="mt-0 mb-1">Change ' + data.id + '</h5>' +
                    '        <p class="text-muted" style="margin-bottom: 0;">\n' +
                    '            <span data-toggle="tooltip" data-placement="top" title="' + data.created_at_nice + '" data-livestamp="' + data.created_at + '">' + data.created_at_nice + '</span>\n' +
                    '            <a href="/changes/' + data.id + '"><i class="fa fa-paperclip" aria-hidden="true"></i></a>\n' +
                    '        </p>' +
                    '        <p class="text-muted" style="margin-bottom: 0;">Apps: ' + makeAppLinks(data.apps, 'apps') + '</p>' +
                    '        <p class="text-muted">Packages: ' + makeAppLinks(data.packages, 'packages') + '</p>' +
                    '    </div>' +
                    '</li>'));

                $('ul.list-unstyled .media').slice(50).remove();
            }
        };
        socket.onerror = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success')
        };

        function makeAppLinks(apps, path) {

            if (apps == null) {
                return '';
            }

            var list = [];
            $.each(apps, function (k, app) {
                var x = $('<a href="/' + path + '/' + app.id + '">' + app.name + '</a>').prop('outerHTML');
                list.push(x);
            });

            return list.join(', ')
        }
    }
}
