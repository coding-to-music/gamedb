if ($('#commits-page').length > 0) {

    const options = {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[3]);
            $(row).attr('data-target', '_blank');
        },
        "columnDefs": [
            // Message
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<a href="' + row[3] + '" target="_blank" class="icon-name"><div class="name">' + row[0] + '</div></a>'
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('id', rowData[4]);
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Time
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[2] + '" data-livestamp="' + row[1] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false,
            },
            // Hash
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[4];
                },
                "orderable": false,
            },
            // Commit
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[5].toLocaleString();
                },
                "orderable": false,
            },
            // Live
            {
                "targets": 4,
                "render": function (data, type, row) {
                    if (row[6]) {
                        return '<i class="fas fa-check text-success"></i>';
                    } else {
                        return '<i class="fas fa-times text-danger"></i>';
                    }
                },
                "orderable": false,
            }
        ]
    };

    let page = null;
    const $table = $('table.table');
    const dt = $table.gdbTable({tableOptions: options});

    dt.on('draw.dt', function (e, settings) {
        page = null;
    });
}
