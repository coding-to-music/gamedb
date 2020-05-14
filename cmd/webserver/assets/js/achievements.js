const $achievementsPage = $('#achievements-page');

if ($achievementsPage.length > 0) {

    const options = {
        "pageLength": 100,
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[7]);
        },
        "columnDefs": [
            // Name
            {
                "targets": 0,
                "render": function (data, type, row) {

                    let name = row[5] + ': ' + row[0] + '<br><small>' + row[2] + '</small>';

                    if (row[5]) {
                        name += '<span class="badge badge-danger float-right ml-1">Hidden</span>';
                    }

                    return '<div class="icon-name"><div class="icon"><img class="tall" data-lazy="' + row[1] + '" alt="" data-lazy-alt="' + row[0] + '"></div><div class="name">' + name + '</div></div>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false,
            },
            // Complete %
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[3] + '%';
                },
                "orderable": false,
            },
            // Search Score
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[6].toLocaleString();
                },
                "orderable": false,
                "visible": false,
            },
        ]
    };

    // Init table
    const searchFields = [
        $('#search'),
    ];

    $('table.table').gdbTable({
        tableOptions: options,
        searchFields: searchFields
    });
}
