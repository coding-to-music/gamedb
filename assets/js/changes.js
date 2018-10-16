if ($('#changes-page').length > 0) {

    var columnDefs = [
        // Change ID
        {
            "targets": 0,
            "render": function (data, type, row) {
                return 'Change ' + row[0].toLocaleString();
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('nowrap', 'nowrap');
            },
            "orderable": false,
            "searchable": false
        },
        // Date
        {
            "targets": 1,
            "render": function (data, type, row) {
                return '<span data-toggle="tooltip" data-placement="left" title="' + row[2] + '" data-livestamp="' + row[1] + '">' + row[2] + '</span>';
            },
            "createdCell": function (td, cellData, rowData, row, col) {
                $(td).attr('nowrap', 'nowrap');
            },
            "orderable": false,
            "searchable": false
        },
        // Apps
        {
            "targets": 2,
            "render": function (data, type, row) {

                var apps = [];
                for (var i in row[3]) {
                    if (row[3].hasOwnProperty(i)) {

                        if (row[3][i].name === '') {
                            apps.push('Unknown App');
                        } else {
                            apps.push('<a href="/games/' + row[3][i].id + '">' + row[3][i].name + '</a>');
                        }
                    }
                }

                return apps.join('<br/>');
            },
            "orderable": false,
            "searchable": false
        },
        // Packages
        {
            "targets": 3,
            "render": function (data, type, row) {

                var packages = [];
                for (var i in row[4]) {
                    if (row[4].hasOwnProperty(i)) {

                        if (row[4][i].name === '') {
                            packages.push('Unknown Package');
                        } else {
                            packages.push('<a href="/packages/' + row[4][i].id + '">' + row[4][i].name + '</a>');
                        }
                    }
                }

                return packages.join('<br/>');
            },
            "orderable": false,
            "searchable": false
        }
    ];

    var $table = $('table.table-datatable2');

    var dt = $table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[0, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]).attr('data-link', '/changes/' + data[0]);
        },
        "columnDefs": columnDefs
    }));

    if (window.WebSocket === undefined) {

        console.log('Your browser does not support WebSockets');

    } else {

        var socket = new WebSocket("wss://" + location.host + "/websocket");
        var $badge = $('#live-badge');

        socket.onopen = function (e) {
            $badge.addClass('badge-success').removeClass('badge-secondary badge-danger')
        };

        socket.onclose = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success')
        };

        socket.onerror = function (e) {
            $badge.addClass('badge-danger').removeClass('badge-secondary badge-success')
        };

        socket.onmessage = function (e) {

            var info = dt.page.info();
            if (info.page === 0) { // Page 1

                var data = $.parseJSON(e.data);

                if (data.Page === 'changes') {

                    for (var i in data.Data) {

                        if (data.Data.hasOwnProperty(i)) {

                            addDataTablesRow(columnDefs, data.Data[i], info.length, $table);
                        }
                    }
                }
            }
        };
    }
}
