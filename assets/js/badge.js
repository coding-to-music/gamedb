const $badgePage = $('#badge-page');

if ($badgePage.length > 0) {

    const options = {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[5]);
        },
        "columnDefs": [
            // Ranks
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return row[0];
                },
            },
            // Icon / Player
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<div class="icon-name"><div class="icon"><img src="' + row[2] + '" alt="' + row[1] + '"></div><div class="name">' + row[1] + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                }
            },
            // Level
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[3].toLocaleString();
                },
            },
            // Time
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            },
        ]
    };

    $('table.table').gdbTable({tableOptions: options});
}
