if ($('#groups-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[2, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[2]);
        },
        "columnDefs": [
            // Icon / Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[3] + '" class="rounded square" alt="' + row[1] + '"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img');
                },
                "orderable": false
            },
            // Headline
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[4]
                },
                "orderable": false,
            },
            // Members
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderable": false,
            },
            // Link
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return '<a target="_blank" href="https://steamcommunity.com/groups/' + row[6] + '"><i class="fas fa-link"></i></a>';
                },
                "orderable": false,
            },
        ]
    }));
}
