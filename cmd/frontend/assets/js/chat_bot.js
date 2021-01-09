if ($('#chat-bot-page').length > 0) {

    const options = {
        "order": [[2, 'desc']],
        "columnDefs": [
            // Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<div class="icon-name">' +
                        '<div class="icon"><img class="tall" alt="" data-lazy="https://cdn.discordapp.com/avatars/' + row[0] + '/' + row[2] + '.png?size=64" data-lazy-alt="' + row[1] + '"></div>' +
                        '<div class="name nowrap">' + row[1] + '<br><small>' + row[6] + '</small></div>' +
                        '</div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img thin');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Message
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3];
                },
                "orderable": false,
            },
            // Time
            {
                "targets": 2,
                "render": function (data, type, row) {
                    if (row[4] && row[4] > 0) {
                        return '<span data-livestamp="' + row[4] + '"></span>'
                            + '<br><small class="text-muted">' + row[5] + '</small>';
                    }
                    return '';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('thin');
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
        ]
    };

    const $table = $('#recent-table');
    const dt = $table.gdbTable({
        tableOptions: options,
    });

    websocketListener('chat-bot', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = JSON.parse(e.data);
            addDataTablesRow(options, data.Data['row_data'], info.length, $table);
        }
    });
}
